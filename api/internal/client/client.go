package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"github.com/KennyMacCormik/HerdMaster/pkg/conv"
	"github.com/KennyMacCormik/HerdMaster/pkg/gin/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Interface interface {
	Get(ctx context.Context, key, requestId string) (any, error)
	Set(ctx context.Context, key string, value any, requestId string) error
	Delete(ctx context.Context, key, requestId string) error
}

type client struct {
	client  *http.Client
	timeout time.Duration
	backend string
}

// Config represents the configuration for the backend HTTP client used in the application.
// It provides essential settings to configure the backend endpoint and request timeout.
//
// Fields:
//   - BackendEndpoint: Specifies the URL of the backend endpoint (e.g., REST API endpoint).
//     Validates as a valid URL and is a required field.
//   - BackendRequestTimeout: Specifies the maximum duration to wait for a backend request to complete.
//     Validates as a duration between 100 ms and 30 s (inclusive).
//
// Usage:
// This struct is designed to integrate seamlessly with the `cfg` and `val` packages for centralized
// configuration management and validation. It ensures the backend client is properly configured
// for reliable communication with the backend services.
type Config struct {
	BackendEndpoint       string        `mapstructure:"backend_endpoint" validate:"url,required"`
	BackendRequestTimeout time.Duration `mapstructure:"backend_request_timeout" validate:"min=100ms,max=30s"`
}

func NewClient(backend string, timeout time.Duration) Interface {
	return &client{
		backend: backend,
		timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
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

func (c *client) Get(ctx context.Context, key, requestId string) (any, error) {
	const (
		traceName = "api.client.get"
		spanName  = "client.get"
	)
	// prep trace
	tracer := otel.Tracer(traceName)
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	r, err := c.prepareWithUrlPath(ctx, http.MethodGet, key)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".prepare", err)
		span.AddEvent("prepare request error", trace.WithAttributes(attribute.String("error", err.Error())))
		return nil, err
	}

	r.Header.Set(middleware.RequestIDKey, requestId)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	b, err := c.invoke(r)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".invoke", err)
		span.AddEvent("invoke request error", trace.WithAttributes(attribute.String("error", err.Error())))
		return nil, err
	}
	return nil, validateResponse(b)
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
