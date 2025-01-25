package compute

import (
	"api/internal/client"
	"context"
	"errors"
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"log/slog"
)

type Interface interface {
	Get(ctx context.Context, key, requestId string) (any, error)
	Set(ctx context.Context, key string, value any) error
	Delete(ctx context.Context, key string) error
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
	val, err := l.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return l.client.Get(ctx, key, requestId)
		} else {
			return nil, err
		}
	}

	return val, nil
}

func (l *layer) Set(ctx context.Context, key string, value any) error {
	const wrap = "compute.layer/set"
	err := l.cache.Set(ctx, key, value)
	if err != nil {
		return fmt.Errorf("%s: could not update cache: %w", wrap, err)
	}

	return l.client.Set(ctx, key, value)
}

func (l *layer) Delete(ctx context.Context, key string) error {
	const wrap = "compute.layer/delete"
	err := l.cache.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("%s: could not update cache: %w", wrap, err)
	}

	return l.client.Delete(ctx, key)
}
