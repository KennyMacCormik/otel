package service_impl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/KennyMacCormik/otel/backend/pkg/cache"
	cacheErrors "github.com/KennyMacCormik/otel/backend/pkg/models/errors/cache"
	otelHelpers "github.com/KennyMacCormik/otel/backend/pkg/otel/helpers"

	"github.com/KennyMacCormik/otel/api/internal/client"
	"github.com/KennyMacCormik/otel/api/internal/service"
)

type serviceLayer struct {
	cache  cache.CacheInterface
	client client.BackendClientInterface
}

func NewServiceLayer(cache cache.CacheInterface, client client.BackendClientInterface) service.ServiceInterface {
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
		}
		otelHelpers.SetSpanExceptionWithoutErr(span, err)
		lg.Warn("cache error", "error", err)

		return l.invokeClientAndStoreValue(ctx, key, requestId)
	}

	span.AddEvent("cache hit")
	lg.Debug("cache hit")

	return val, nil
}

func (l *serviceLayer) Set(ctx context.Context, key string, value any, requestId string, lg *slog.Logger) (int, error) {
	const (
		spanName = "compute.set"
	)

	ctx, span := otelHelpers.StartSpanWithCtx(ctx, spanName, spanName)
	defer span.End()

	if _, err := l.cache.Set(ctx, key, value); err != nil {
		otelHelpers.SetSpanExceptionWithoutErr(span, err)
		lg.Error(fmt.Sprintf("%s: could not update cache", spanName), "error", err)
	}

	return l.client.Set(ctx, key, value, requestId)
}

func (l *serviceLayer) Delete(ctx context.Context, key, requestId string, lg *slog.Logger) error {
	const (
		spanName = "compute.delete"
	)

	ctx, span := otelHelpers.StartSpanWithCtx(ctx, spanName, spanName)
	defer span.End()

	if err := l.cache.Delete(ctx, key); err != nil {
		otelHelpers.SetSpanExceptionWithoutErr(span, err)
		lg.Error(fmt.Sprintf("%s: could not update cache", spanName), "error", err)
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
