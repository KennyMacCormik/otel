package shardedCache

import (
	"context"
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"github.com/KennyMacCormik/HerdMaster/pkg/conv"
	"hash/fnv"
	"sync"
	"sync/atomic"
)

const defaultShardNumber = 10
const fallbackShard = 0

type shardedCache struct {
	shardNumber int

	shards []cache.Interface
	mtx    sync.RWMutex

	closer     func(ctx context.Context) error
	closed     atomic.Bool
	closedOnce sync.Once
}

type InitOptions func(t *shardedCache)

func WithOverrideDefaults(shardNumber int) InitOptions {
	return func(s *shardedCache) {
		if shardNumber < 1 {
			shardNumber = defaultShardNumber
		}
		s.shardNumber = shardNumber
	}
}

func NewShardedCache(initFn func() cache.Interface, opts ...InitOptions) (cache.Interface, error) {
	const wrap = "NewShardedCache"
	err := cache.WithValueValidation(initFn, wrap)()
	if err != nil {
		return nil, err
	}

	s := &shardedCache{shardNumber: defaultShardNumber}

	for _, opt := range opts {
		opt(s)
	}

	s.shards = make([]cache.Interface, 0, s.shardNumber)

	for i := 0; i < s.shardNumber; i++ {
		shard := initFn()
		s.shards = append(s.shards, shard)
		s.closer = wrapCloser(s.closer, shard.Close)
	}

	return s, nil
}

func wrapCloser(closers ...func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		var err error
		for _, fn := range closers {
			if fn != nil {
				err1 := fn(ctx)
				if err1 != nil {
					err = fmt.Errorf("%w: %w", err, err1)
				}
			}
		}
		return err
	}
}

func (s *shardedCache) Get(ctx context.Context, key string) (any, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	const wrap = "ttlCache/Get"
	if err := cache.ValidateInput(
		cache.WithClosedValidation(&s.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
		cache.WithKeyValidation(key, wrap),
	); err != nil {
		return nil, err
	}

	return s.shards[getShardNumber(key, s.shardNumber)].Get(ctx, key)
}

func (s *shardedCache) Set(ctx context.Context, key string, value any) error {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	const wrap = "ttlCache/Set"
	if err := cache.ValidateInput(
		cache.WithClosedValidation(&s.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
		cache.WithKeyValidation(key, wrap),
		cache.WithValueValidation(value, wrap),
	); err != nil {
		return err
	}

	return s.shards[getShardNumber(key, s.shardNumber)].Set(ctx, key, value)
}

func (s *shardedCache) Delete(ctx context.Context, key string) error {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	const wrap = "ttlCache/Delete"
	if err := cache.ValidateInput(
		cache.WithClosedValidation(&s.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
		cache.WithKeyValidation(key, wrap),
	); err != nil {
		return err
	}

	return s.shards[getShardNumber(key, s.shardNumber)].Delete(ctx, key)
}

func (s *shardedCache) Close(ctx context.Context) error {
	var err error
	s.closedOnce.Do(func() {
		s.mtx.Lock()
		defer s.mtx.Unlock()

		err = s.closer(ctx)

		s.shards = nil
		s.closed.Store(true)
	})

	return err
}

func (s *shardedCache) GetLength() (int, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	const wrap = "ttlCache/GetKeys"
	if err := cache.ValidateInput(
		cache.WithClosedValidation(&s.closed, wrap),
	); err != nil {
		return 0, err
	}

	return s.getShardedCacheLen()
}

func (s *shardedCache) getShardedCacheLen() (int, error) {
	const wrap = "ttlCache/getShardedCacheLen"
	var i int
	for shardNum := range s.shards {
		num, err := s.shards[shardNum].GetLength()
		if err != nil {
			return 0, fmt.Errorf("%s: shard %d :%w", wrap, shardNum, err)
		}
		i += num
	}
	return i, nil
}

func (s *shardedCache) GetKeys(ctx context.Context) ([]string, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	const wrap = "ttlCache/GetKeys"
	if err := cache.ValidateInput(
		cache.WithClosedValidation(&s.closed, wrap),
		cache.WithCtxValidation(ctx, wrap),
	); err != nil {
		return nil, err
	}

	ln, err := s.GetLength()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, ln)
	resultCh := make(chan []string, s.shardNumber)
	errCh := make(chan error, s.shardNumber)

	s.loopShards(ctx, resultCh, errCh)

	close(resultCh)
	close(errCh)

	if len(errCh) > 0 {
		return nil, <-errCh
	}
	for keys := range resultCh {
		result = append(result, keys...)
	}

	return result, nil
}

func (s *shardedCache) loopShards(ctx context.Context, resultCh chan []string, errCh chan error) {
	var wg sync.WaitGroup

	for i := range s.shards {
		wg.Add(1)
		go func(shardIdx int) {
			defer wg.Done()
			keys, err := s.getKeys(ctx, shardIdx)
			if err != nil {
				errCh <- err
				return
			}
			resultCh <- keys
		}(i)
	}

	wg.Wait()
}

func (s *shardedCache) getKeys(ctx context.Context, shard int) ([]string, error) {
	const wrap = "ttlCache/getKeys"

	result, err := s.shards[shard].GetKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: shard %d: %w", wrap, shard, err)
	}

	return result, nil
}

func getShardNumber(key string, shardNumber int) int {
	if key == "" {
		return fallbackShard
	}
	hasher := fnv.New64a()
	_, err := hasher.Write(conv.StrToBytes(key))
	if err != nil {
		return fallbackShard
	}
	return int(hasher.Sum64() % uint64(shardNumber))
}
