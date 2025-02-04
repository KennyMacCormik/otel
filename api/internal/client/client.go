package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"github.com/KennyMacCormik/HerdMaster/pkg/conv"
	"github.com/KennyMacCormik/HerdMaster/pkg/gin/middleware"
	"github.com/KennyMacCormik/otel/backend/pkg/gin/gin_request_id"
	otelCustomFuncs "github.com/KennyMacCormik/otel/backend/pkg/otel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type BackendClientInterface interface {
	Get(ctx context.Context, key, requestId string) (any, error)
	Set(ctx context.Context, key string, value any, requestId string) error
	Delete(ctx context.Context, key, requestId string) error
}

type client struct {
	client  *http.Client
	timeout time.Duration
	backend string
}

func NewBackendClient(backend string, timeout time.Duration) BackendClientInterface {
	return &client{
		backend: backend,
		timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *client) Get(ctx context.Context, key, requestId string) (any, error) {
	const (
		spanName = "client.get"
	)

	ctx, span := otelCustomFuncs.StartSpanWithCtx(ctx, spanName, spanName)
	defer span.End()

	r, err := c.prepareWithUrlPath(ctx, http.MethodGet, key)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".prepare", err)
		otelCustomFuncs.SetSpanErr(span, err)
		return nil, err
	}

	r.Header.Set(gin_request_id.RequestIDKey, requestId)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	b, err := c.invoke(r)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".invoke", err)
		otelCustomFuncs.SetSpanErr(span, err)
		return nil, err
	}
	return nil, validateResponse(b)
}

func (c *client) Set(ctx context.Context, key string, value any, requestId string) error {
	const (
		traceName = "api.client.set"
		spanName  = "client.set"
	)
	// prep trace
	tracer := otel.Tracer(traceName)
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	r, err := c.prepareWithBody(ctx, http.MethodPost, key, value)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".prepare", err)
		span.AddEvent("prepare request error", trace.WithAttributes(attribute.String("error", err.Error())))
		return err
	}

	r.Header.Set(middleware.RequestIDKey, requestId)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	_, err = c.invoke(r)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".invoke", err)
		span.AddEvent("invoke request error", trace.WithAttributes(attribute.String("error", err.Error())))
		return err
	}
	return nil
}

func (c *client) Delete(ctx context.Context, key string, requestId string) error {
	const (
		traceName = "api.client.delete"
		spanName  = "client.delete"
	)
	// prep trace
	tracer := otel.Tracer(traceName)
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	r, err := c.prepareWithUrlPath(ctx, http.MethodDelete, key)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".prepare", err)
		span.AddEvent("prepare request error", trace.WithAttributes(attribute.String("error", err.Error())))
		return err
	}

	r.Header.Set(middleware.RequestIDKey, requestId)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	_, err = c.invoke(r)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".invoke", err)
		span.AddEvent("invoke request error", trace.WithAttributes(attribute.String("error", err.Error())))
		return err
	}
	return nil
}

func (c *client) prepareWithBody(ctx context.Context, method, key string, val any) (*http.Request, error) {
	body := map[string]any{"key": key, "value": val}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(jsonBody)

	return http.NewRequestWithContext(ctx, method, c.backend, reader)
}

func (c *client) prepareWithUrlPath(ctx context.Context, method, key string) (*http.Request, error) {
	encodedKey := url.PathEscape(key)
	path, err := url.JoinPath(c.backend, encodedKey)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, path, nil)
}

func (c *client) invoke(r *http.Request) ([]byte, error) {
	resp, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, cache.ErrNotFound
	}
	if resp.StatusCode == http.StatusInternalServerError {
		return nil, errors.New("internal server error")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func validateResponse(b []byte) error {
	str := conv.BytesToStr(b)
	if strings.Contains(str, "malformed request") {
		return errors.New("malformed request")
	}
	if strings.Contains(str, "internal server error") {
		return errors.New("internal server error")
	}
	if strings.Contains(str, "not found") {
		return cache.ErrNotFound
	}
	return nil
}

func startSpanWithCtx(ctx context.Context, traceName, spanName string) (context.Context, trace.Span) {
	tracer := otel.Tracer(traceName)
	newCtx, span := tracer.Start(ctx, spanName)
	return newCtx, span
}

func setSpanErr(span trace.Span, err error) {
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
	span.SetAttributes(attribute.String("error.message", err.Error()))
}
