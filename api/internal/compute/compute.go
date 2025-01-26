package compute

import (
	"api/internal/client"
	"context"
	"errors"
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
)

type Interface interface {
	Get(ctx context.Context, key, requestId string) (any, error)
	Set(ctx context.Context, key string, value any, requestId string) error
	Delete(ctx context.Context, key, requestId string) error
}

type layer struct {
	lg     *slog.Logger
	cache  cache.Interface
	client client.Interface
}

func NewComputeLayer(cache cache.Interface, client client.Interface, lg *slog.Logger) Interface {
	return &layer{cache: cache, lg: lg, client: client}
}

func (l *layer) Get(ctx context.Context, key, requestId string) (any, error) {
	const (
		traceName = "api.compute.get"
		spanName  = "compute.get"
	)
	tracer := otel.Tracer(traceName)
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	val, err := l.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			span.AddEvent("cache miss")
			return l.client.Get(ctx, key, requestId)
		} else {
			return nil, err
		}
	}
	span.AddEvent("cache hit")
	return val, nil
}

func (l *layer) Set(ctx context.Context, key string, value any, requestId string) error {
	const (
		traceName = "api.compute.set"
		spanName  = "compute.set"
	)
	tracer := otel.Tracer(traceName)
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	if err := l.cache.Set(ctx, key, value); err != nil {
		span.AddEvent("could not update cache", trace.WithAttributes(attribute.String("error", err.Error())))
		return fmt.Errorf("%s: could not update cache: %w", spanName, err)
	}

	return l.client.Set(ctx, key, value, requestId)
}

func (l *layer) Delete(ctx context.Context, key, requestId string) error {
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
