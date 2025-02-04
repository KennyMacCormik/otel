package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/KennyMacCormik/otel/backend/pkg/cache"
	cacheErrors "github.com/KennyMacCormik/otel/backend/pkg/models/errors/cache"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/KennyMacCormik/otel/api/internal/client"
)

type ServiceInterface interface {
	Get(ctx context.Context, key, requestId string) (any, error)
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

func (l *serviceLayer) Get(ctx context.Context, key, requestId string) (any, error) {
	const (
		traceName = "api.compute.get"
		spanName  = "compute.get"
	)
	tracer := otel.Tracer(traceName)
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	val, err := l.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, cacheErrors.ErrNotFound) {
			span.AddEvent("cache miss")
			return l.client.Get(ctx, key, requestId)
		} else {
			return nil, err
		}
	}
	span.AddEvent("cache hit")
	return val, nil
}

func (l *serviceLayer) Set(ctx context.Context, key string, value any, requestId string) (int, error) {
	const (
		traceName = "api.compute.set"
		spanName  = "compute.set"
	)
	tracer := otel.Tracer(traceName)
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	if code, err := l.cache.Set(ctx, key, value); err != nil {
		span.AddEvent("could not update cache", trace.WithAttributes(attribute.String("error", err.Error())))
		return 0, fmt.Errorf("%s: could not update cache: %w", spanName, err)
	}

	return l.client.Set(ctx, key, value, requestId)
}

func (l *serviceLayer) Delete(ctx context.Context, key, requestId string) error {
	const (
		traceName = "api.compute.delete"
		spanName  = "compute.delete"
	)
	tracer := otel.Tracer(traceName)
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	if err := l.cache.Delete(ctx, key); err != nil {
		span.AddEvent("could not update cache", trace.WithAttributes(attribute.String("error", err.Error())))
		return fmt.Errorf("%s: could not update cache: %w", spanName, err)
	}

	return l.client.Delete(ctx, key, requestId)
}
