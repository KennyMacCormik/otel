package sync_map

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/KennyMacCormik/otel/backend/pkg/cache"
	cacheErrors "github.com/KennyMacCormik/otel/backend/pkg/models/errors/cache"
)

const (
	defaultKeyCapacity int64 = 128
	defaultTimeout           = 30 * time.Second
)

type InitOptions func(sm *syncMap)
type syncMap struct {
	m         sync.Map
	closeOnce sync.Once
	closed    atomic.Bool

	keyCapacity int64
	timeout     time.Duration
}

func NewSyncMapCache(opts ...InitOptions) cache.CacheInterface {
	c := &syncMap{keyCapacity: defaultKeyCapacity, timeout: defaultTimeout}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithOverrideDefaults(KeyCapacity int64, Timeout time.Duration) InitOptions {
	return func(sm *syncMap) {
		if KeyCapacity <= 0 {
			KeyCapacity = defaultKeyCapacity
		}

		if Timeout <= 0 {
			Timeout = defaultTimeout
		}

		sm.keyCapacity = KeyCapacity
		sm.timeout = Timeout
	}
}

func (sm *syncMap) Get(ctx context.Context, key string) (any, error) {
	const wrap = "syncMap/Get"

	if err := cache.ValidateInput(
		cache.WithClosedValidation(&sm.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
		cache.WithKeyValidation(key, wrap),
	); err != nil {
		return nil, err
	}

	value, ok := sm.m.Load(key)
	if !ok {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%s: %w", wrap, ctx.Err())
		}
		return nil, cacheErrors.NewErrKeyNotFound(key)
	}

	return value, nil
}

func (sm *syncMap) Set(ctx context.Context, key string, value any) (int, error) {
	const wrap = "syncMap/Set"

	if err := cache.ValidateInput(
		cache.WithClosedValidation(&sm.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
		cache.WithKeyValidation(key, wrap),
		cache.WithValueValidation(value, wrap),
	); err != nil {
		return 0, err
	}

	// 201 Created
	val, ok := sm.m.Load(key)
	if !ok {
		sm.m.Store(key, value)
		return 201, nil
	}

	// 204 No Content
	if val == value {
		return 204, nil
	}

	sm.m.Store(key, value)
	return 200, nil
}

func (sm *syncMap) Delete(ctx context.Context, key string) error {
	const wrap = "syncMap/Delete"

	if err := cache.ValidateInput(
		cache.WithClosedValidation(&sm.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
		cache.WithKeyValidation(key, wrap),
	); err != nil {
		return err
	}

	sm.m.Delete(key)
	return nil
}

// Close stops the cache.
// Even if the error was returned,
// cache is still closed, but might not release all its content.
// If you pass incorrect ctx,
// it will be replaced with ctx with default timeout to ensure cache is properly deallocated.
func (sm *syncMap) Close(ctx context.Context) error {
	var err error

	ctx, cancel := sm.normalizeCtx(ctx)
	if cancel != nil {
		defer cancel()
	}

	sm.closeOnce.Do(func() {
		sm.closed.Store(true)
		err = sm.clearSyncMapWithTimeout(ctx)
	})

	return err
}

func (sm *syncMap) GetLength() (int64, error) {
	const wrap = "syncMap/Delete"

	if err := cache.ValidateInput(
		cache.WithClosedValidation(&sm.closed, wrap),
	); err != nil {
		return 0, err
	}

	var i int64

	sm.m.Range(func(k, v any) bool {
		i++
		return true
	})

	return i, nil
}

func (sm *syncMap) GetKeys(ctx context.Context) ([]string, error) {
	const wrap = "syncMap/GetKeys"

	if err := cache.ValidateInput(
		cache.WithClosedValidation(&sm.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
	); err != nil {
		return nil, err
	}

	keys := make([]string, 0, sm.keyCapacity)
	ctx, cancel := sm.normalizeCtx(ctx)
	if cancel != nil {
		defer cancel()
	}

	var err error
	sm.m.Range(func(key, value any) bool {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		strKey, ok := key.(string)
		if !ok {
			err = cacheErrors.NewErrTypeCastFailed(key, value, wrap)
			return false
		}

		keys = append(keys, strKey)

		return true
	})
	if err != nil {
		return nil, err
	}

	if ctx.Err() != nil {
		return nil, fmt.Errorf("%s: %w", wrap, ctx.Err())
	}

	return keys, nil
}

// clearSyncMapWithTimeout clears map content.
// In case nil, time-out or invalid context is supplied,
// it will be replaced with valid ctx with defaultTimeout
func (sm *syncMap) clearSyncMapWithTimeout(ctx context.Context) error {
	const wrap = "syncMap/clearSyncMapWithTimeout"

	ctx, cancel := sm.normalizeCtx(ctx)
	if cancel != nil {
		defer cancel()
	}

	sm.m.Range(func(key, value any) bool {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		sm.m.Delete(key)

		return true
	})

	if ctx.Err() != nil {
		return fmt.Errorf("%s: %w", wrap, ctx.Err())
	}

	return nil
}

// normalizeCtx returns updated ctx if one deemed not valid. Always check returned func for nil before using it
//
// Example:
//
// ctx, cancel := sm.normalizeCtx(ctx)
//
//	if cancel != nil {
//	 defer cancel()
//	}
func (sm *syncMap) normalizeCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == context.Background() {
		return ctx, nil
	}

	if ctx == nil || ctx.Err() != nil {
		return context.WithTimeout(context.Background(), sm.timeout)
	}

	return ctx, nil
}
