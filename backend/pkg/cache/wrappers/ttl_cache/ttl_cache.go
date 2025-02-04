package ttl_cache

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/KennyMacCormik/common/log"

	"github.com/KennyMacCormik/otel/backend/pkg/cache"
	cacheErrors "github.com/KennyMacCormik/otel/backend/pkg/models/errors/cache"
	ttlCacheErrors "github.com/KennyMacCormik/otel/backend/pkg/models/errors/ttl_cache"
	ttlCacheModels "github.com/KennyMacCormik/otel/backend/pkg/models/ttl_cache"
)

const (
	defaultTTL                        = time.Minute
	defaultTickerTTL                  = 30 * time.Second
	defaultGetKeysTTL                 = 10 * time.Second
	defaultDeleteExpiredKeysTTL       = 1 * time.Second
	defaultSkewPercent          int64 = 10
)

type ttlCache struct {
	impl cache.CacheInterface

	ttl, tickerTTL, getKeysTTL, deleteExpiredKeysTTL time.Duration
	skewPercent                                      int64

	ticker     *time.Ticker
	closedOnce sync.Once
	closed     atomic.Bool
	closeCh    chan struct{}
}

type InitOptions func(t *ttlCache)

func WithOverrideDefaults(ttl, tickerTTL, getKeysTTL,
	deleteExpiredKeysTTL time.Duration, skewPercent int64) InitOptions {
	return func(t *ttlCache) {
		if ttl <= 0 {
			ttl = defaultTTL
		}

		if tickerTTL <= 0 {
			tickerTTL = defaultTickerTTL
		}

		if getKeysTTL <= 0 {
			getKeysTTL = defaultGetKeysTTL
		}

		if deleteExpiredKeysTTL <= 0 {
			deleteExpiredKeysTTL = defaultDeleteExpiredKeysTTL
		}

		if skewPercent <= 0 {
			skewPercent = defaultSkewPercent
		}

		t.ttl = ttl
		t.tickerTTL = tickerTTL
		t.getKeysTTL = getKeysTTL
		t.deleteExpiredKeysTTL = deleteExpiredKeysTTL
		t.skewPercent = skewPercent
	}
}

func NewTtlCache(impl cache.CacheInterface, opts ...InitOptions) (cache.CacheInterface, error) {
	const wrap = "NewTtlCache"

	err := cache.WithValueValidation(impl, wrap)()
	if err != nil {
		return nil, err
	}

	t := &ttlCache{
		impl:                 impl,
		ttl:                  defaultTTL,
		tickerTTL:            defaultTickerTTL,
		getKeysTTL:           defaultGetKeysTTL,
		deleteExpiredKeysTTL: defaultDeleteExpiredKeysTTL,
		skewPercent:          defaultSkewPercent,
		closeCh:              make(chan struct{}),
	}

	for _, opt := range opts {
		opt(t)
	}

	t.ticker = time.NewTicker(t.tickerTTL)
	go t.expireCache()

	return t, nil
}

func (t *ttlCache) Get(ctx context.Context, key string) (any, error) {
	const wrap = "ttlCache/Get"
	if err := cache.ValidateInput(
		cache.WithClosedValidation(&t.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
		cache.WithKeyValidation(key, wrap),
	); err != nil {
		return nil, err
	}

	val, err := t.impl.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	castedValue, ok := val.(*ttlCacheModels.TtlCacheEntry)
	if !ok {
		err = cacheErrors.NewErrTypeCastFailed(key, val, wrap)
		return nil, err
	}

	if ttlExpired(castedValue.ExpiresAt) {
		return nil, NewErrTimeout(key, wrap, castedValue.ExpiresAt)
	}

	return castedValue.Value, nil
}

func (t *ttlCache) Set(ctx context.Context, key string, value any) (int, error) {
	const wrap = "ttlCache/Set"
	if err := cache.ValidateInput(
		cache.WithClosedValidation(&t.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
		cache.WithKeyValidation(key, wrap),
		cache.WithValueValidation(value, wrap),
	); err != nil {
		return 0, err
	}

	return t.impl.Set(ctx, key, &ttlCacheModels.TtlCacheEntry{Value: value, ExpiresAt: t.getTtl()})
}

func (t *ttlCache) Delete(ctx context.Context, key string) error {
	const wrap = "ttlCache/Delete"
	if err := cache.ValidateInput(
		cache.WithClosedValidation(&t.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
		cache.WithKeyValidation(key, wrap),
	); err != nil {
		return err
	}

	return t.impl.Delete(ctx, key)
}

func (t *ttlCache) Close(ctx context.Context) error {
	var err error
	t.closedOnce.Do(func() {
		t.closed.Store(true)
		close(t.closeCh)
		err = t.impl.Close(ctx)
	})

	return err
}

func (t *ttlCache) GetKeys(ctx context.Context) ([]string, error) {
	const wrap = "ttlCache/GetKeys"
	if err := cache.ValidateInput(
		cache.WithClosedValidation(&t.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
	); err != nil {
		return nil, err
	}

	return t.impl.GetKeys(ctx)
}

func (t *ttlCache) GetLength() (int64, error) {
	const wrap = "ttlCache/GetKeys"
	if err := cache.ValidateInput(
		cache.WithClosedValidation(&t.closed, wrap),
	); err != nil {
		return 0, err
	}

	return t.impl.GetLength()
}

func (t *ttlCache) getTtl() time.Time {
	return getRandomSkew(getTime().Add(t.ttl), t.skewPercent)
}

func (t *ttlCache) expireCache() {
	for {
		select {
		case <-t.ticker.C:
			list, _ := t.getImplKeys()
			for _, key := range list {
				t.deleteExpiredKey(key)
			}
		case <-t.closeCh:
			t.ticker.Stop()
			return
		}
	}

}

func (t *ttlCache) deleteExpiredKey(key string) {
	const wrap = "ttlCache/deleteExpiredKey"

	ctx, cancel := context.WithTimeout(context.Background(), t.deleteExpiredKeysTTL)
	defer cancel()

	// presumably always succeed
	val, err := t.impl.Get(ctx, key)
	if err != nil {
		log.Error(fmt.Sprintf("%s: failed to get key: %s", wrap, key), "key", key, "err", err)
		return
	}

	castedValue, ok := val.(*ttlCacheModels.TtlCacheEntry)
	if !ok {
		log.Error(fmt.Sprintf("%s: failed to type cast: key [%s]", wrap, key), "key", key, "err", err)
		return
	}

	if !ttlExpired(castedValue.ExpiresAt) {
		return
	}

	err = t.impl.Delete(ctx, key)
	if err != nil {
		log.Error(fmt.Sprintf("%s: failed to delete key", wrap), "key", key, "err", err)
	}
}

func (t *ttlCache) getImplKeys() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.getKeysTTL)
	defer cancel()

	return t.impl.GetKeys(ctx)
}

func ttlExpired(ttl time.Time) bool {
	return getTime().After(ttl)
}

func getTime() time.Time {
	return time.Now()
}

func getRandomSkew(expiresAt time.Time, skewPercent int64) time.Time {
	// to be thread-safe
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	ttl := time.Until(expiresAt)
	skewRange := ttl.Nanoseconds() * skewPercent / 100
	skew := rnd.Int63n(2*skewRange+1) - skewRange

	return expiresAt.Add(time.Duration(skew))
}

type ErrTimeout struct {
	key            string
	expirationTime time.Time
	callerInfo     string
	signalErr      error
}

func NewErrTimeout(key, callerInfo string, expirationTime time.Time) *ErrTimeout {
	return &ErrTimeout{key: key, expirationTime: expirationTime, signalErr: ttlCacheErrors.ErrExpired, callerInfo: callerInfo}
}

func (e *ErrTimeout) GetCallerInfo() string {
	return e.callerInfo
}

func (e *ErrTimeout) GetKey() string {
	return e.key
}

func (e *ErrTimeout) GetTtl() time.Time {
	return e.expirationTime
}

func (e *ErrTimeout) Error() string {
	return fmt.Errorf("%s: key [%s] ttl [%s]: %w", e.callerInfo, e.key, e.expirationTime, e.signalErr).Error()
}

// Is function only checks for an ErrTimeout type and don't compare for an underlying key or ttl
func (e *ErrTimeout) Is(target error) bool {
	_, ok := target.(*ErrTimeout)

	return ok
}

func (e *ErrTimeout) Unwrap() error {
	return e.signalErr
}
