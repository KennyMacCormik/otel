package client

import (
	"bytes"
	"context"
	"encoding/json"
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
	"time"
)

type Interface interface {
	Get(ctx context.Context, key, requestId string) (any, error)
	Set(ctx context.Context, key string, value any) error
	Delete(ctx context.Context, key string) error
}

type client struct {
	client  *http.Client
	timeout time.Duration
	backend string
}

func NewClient(backend string, timeout time.Duration) Interface {
	return &client{
		backend: backend,
		timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *client) prepare(ctx context.Context, method, key string) (*http.Request, error) {
	body := map[string]string{"key": key}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(jsonBody)

	return http.NewRequestWithContext(ctx, method, c.backend, reader)
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
		return nil, fmt.Errorf("server-side error")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *client) Get(ctx context.Context, key, requestId string) (any, error) {
	const wrap = "client/Get"
	// prep trace
	tracer := otel.Tracer("backend/get")
	ctx, span := tracer.Start(ctx, "get")
	defer span.End()

	r, err := c.prepare(ctx, http.MethodGet, key)
	if err != nil {
		err = fmt.Errorf("%s, failed prepare request: %w", wrap+"/prepare", err)
		span.AddEvent("failed prepare request", trace.WithAttributes(attribute.String("error", err.Error())))
		return nil, err
	}

	r.Header.Set(middleware.RequestIDKey, requestId)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	b, err := c.invoke(r)
	if err != nil {
		err = fmt.Errorf("%s, failed invoke request: %w", wrap+"/invoke", err)
		span.AddEvent("failed invoke request", trace.WithAttributes(attribute.String("error", err.Error())))
		return nil, err
	}
	return conv.BytesToStr(b), nil
}

func (c *client) Set(ctx context.Context, key string, value any) error {
	//TODO implement me
	//panic("implement me")
	return nil
}

func (c *client) Delete(ctx context.Context, key string) error {
	//TODO implement me
	//panic("implement me")
	return nil
}
