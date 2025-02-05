package client_impl

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

	"github.com/KennyMacCormik/common/conv"
	"github.com/KennyMacCormik/otel/backend/pkg/gin/gin_request_id"
	cacheErrors "github.com/KennyMacCormik/otel/backend/pkg/models/errors/cache"
	otelHelpers "github.com/KennyMacCormik/otel/backend/pkg/otel/helpers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/KennyMacCormik/otel/api/internal/client"
)

type clientImpl struct {
	client  *http.Client
	timeout time.Duration
	backend string
}

func NewBackendClient(backend string, timeout time.Duration) client.BackendClientInterface {
	return &clientImpl{
		backend: backend,
		timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *clientImpl) Get(ctx context.Context, key, requestId string) (any, error) {
	const (
		spanName = "client.get"
	)

	ctx, span := otelHelpers.StartSpanWithCtx(ctx, spanName, spanName)
	defer span.End()

	r, err := c.prepareWithUrlPath(ctx, http.MethodGet, key)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".prepare", err)
		otelHelpers.SetSpanExceptionWithErr(span, err)
		return nil, err
	}

	r.Header.Set(gin_request_id.RequestIDKey, requestId)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	b, err := c.invoke(r)
	if err != nil {
		if errors.Is(err, cacheErrors.ErrNotFound) {
			return nil, err
		}
		err = fmt.Errorf("%s: %w", spanName+".invoke", err)
		otelHelpers.SetSpanExceptionWithErr(span, err)
		return nil, err
	}
	// TODO: fix incorrect return
	return nil, validateResponse(b)
}

// TODO: fix return codes

func (c *clientImpl) Set(ctx context.Context, key string, value any, requestId string) error {
	const (
		spanName = "client.set"
	)

	ctx, span := otelHelpers.StartSpanWithCtx(ctx, spanName, spanName)
	defer span.End()

	r, err := c.prepareWithBody(ctx, http.MethodPut, key, value)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".prepare", err)
		otelHelpers.SetSpanExceptionWithErr(span, err)
		return err
	}

	r.Header.Set(gin_request_id.RequestIDKey, requestId)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	_, err = c.invoke(r)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".invoke", err)
		otelHelpers.SetSpanExceptionWithErr(span, err)
		return err
	}
	return nil
}

func (c *clientImpl) Delete(ctx context.Context, key string, requestId string) error {
	const (
		spanName = "client.delete"
	)

	ctx, span := otelHelpers.StartSpanWithCtx(ctx, spanName, spanName)
	defer span.End()

	r, err := c.prepareWithUrlPath(ctx, http.MethodDelete, key)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".prepare", err)
		otelHelpers.SetSpanExceptionWithErr(span, err)
		return err
	}

	r.Header.Set(gin_request_id.RequestIDKey, requestId)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	_, err = c.invoke(r)
	if err != nil {
		err = fmt.Errorf("%s: %w", spanName+".invoke", err)
		otelHelpers.SetSpanExceptionWithErr(span, err)
		return err
	}
	return nil
}

func (c *clientImpl) prepareWithBody(ctx context.Context, method, key string, val any) (*http.Request, error) {
	body := map[string]any{"key": key, "value": val}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(jsonBody)

	return http.NewRequestWithContext(ctx, method, c.backend, reader)
}

func (c *clientImpl) prepareWithUrlPath(ctx context.Context, method, key string) (*http.Request, error) {
	encodedKey := url.QueryEscape(key)

	path, err := url.JoinPath(c.backend, encodedKey)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, path, nil)
}

func (c *clientImpl) invoke(r *http.Request) ([]byte, error) {
	resp, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, cacheErrors.ErrNotFound
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
		return cacheErrors.ErrNotFound
	}
	return nil
}
