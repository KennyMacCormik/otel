package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/KennyMacCormik/otel/backend/pkg/cache"
	cacheErrors "github.com/KennyMacCormik/otel/backend/pkg/models/errors/cache"
	otelHelpers "github.com/KennyMacCormik/otel/backend/pkg/otel/helpers"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/KennyMacCormik/otel/api/internal/client"
)

type ServiceInterface interface {
	Get(ctx context.Context, key, requestId string, lg *slog.Logger) (any, error)
	Set(ctx context.Context, key string, value any, requestId string) (int, error)
	Delete(ctx context.Context, key, requestId string) error
}

type serviceLayer struct {
	cache  cache.CacheInterface
	client client.BackendClientInterface
}

func NewServiceLayer(cache cache.CacheInterface, client client.BackendClientInterface) ServiceInterface {
	return &serviceLayer{cache: cache, client: client}
}

func (l *serviceLayer) Get(ctx context.Context, key, requestId string, lg *slog.Logger) (any, error) {
	const (
		spanName = "service.get"
	)

	ctx, span := otelHelpers.StartSpanWithCtx(ctx, spanName, spanName)
	defer span.End()

	val, err := l.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, cacheErrors.ErrNotFound) {
			span.AddEvent("cache miss")
			lg.Debug("cache miss")

			return l.invokeClientAndStoreValue(ctx, key, requestId)
		} else {
			otelHelpers.SetSpanExceptionWithoutErr(span, err)
			lg.Error("cache error", "error", err)

			return l.invokeClientAndStoreValue(ctx, key, requestId)
		}
	}

	span.AddEvent("cache hit")
	lg.Debug("cache hit")

	return val, nil
}

func (l *serviceLayer) Set(ctx context.Context, key string, value any, requestId string) (int, error) {
	const (
		spanName = "compute.set"
	)

	ctx, span := otelHelpers.StartSpanWithCtx(ctx, spanName, spanName)
	defer span.End()

	if _, err := l.cache.Set(ctx, key, value); err != nil {
		span.AddEvent("could not update cache", trace.WithAttributes(attribute.String("error", err.Error())))
		return 0, fmt.Errorf("%s: could not update cache: %w", spanName, err)
	}

	return 0, l.client.Set(ctx, key, value, requestId)
}

func (l *serviceLayer) Delete(ctx context.Context, key, requestId string) error {
	const (
		spanName = "compute.delete"
	)

	ctx, span := otelHelpers.StartSpanWithCtx(ctx, spanName, spanName)
	defer span.End()

	if err := l.cache.Delete(ctx, key); err != nil {
		span.AddEvent("could not update cache", trace.WithAttributes(attribute.String("error", err.Error())))
		return fmt.Errorf("%s: could not update cache: %w", spanName, err)
	}

	return l.client.Delete(ctx, key, requestId)
}

func (l *serviceLayer) invokeClientAndStoreValue(ctx context.Context, key, requestId string) (any, error) {
	val, err := l.client.Get(ctx, key, requestId)
	if err != nil {
		return nil, err
	}

	_, _ = l.cache.Set(ctx, key, val)

	return val, nil
}
