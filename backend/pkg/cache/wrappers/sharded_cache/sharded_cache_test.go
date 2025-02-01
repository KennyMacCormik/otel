package sharded_cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mockCache "github.com/KennyMacCormik/otel/backend/pkg/cache/mocks"
	cache2 "github.com/KennyMacCormik/otel/backend/pkg/models/errors/cache"

	"github.com/KennyMacCormik/otel/backend/pkg/cache"
)

// TODO: update SET method tests

const testShardNum int64 = 3

func initFunc(t *testing.T) cache.CacheInterface {
	return mockCache.NewMockCacheInterface(t)
}

func typeCast(t *testing.T, c cache.CacheInterface) *shardedCache {
	impl, ok := c.(*shardedCache)
	require.True(t, ok, "type cast shall succeed")
	return impl
}

func TestShardedCache_New(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		fn := func() cache.CacheInterface { return initFunc(t) }
		c, err := NewShardedCache(fn)
		require.NoError(t, err)
		require.NotNil(t, c, "should return non")
		require.Implements(t, (*cache.CacheInterface)(nil), c, "result should implement cache.Interface")

		sc := typeCast(t, c)
		require.NotNil(t, sc, "should return shardedCache")
		assert.Equal(t, defaultShardNumber, sc.shardNumber, "shard number should be default")
		assert.Equal(t, defaultShardNumber, int64(len(sc.shards)), "shards should be init and equal to default")
		assert.False(t, sc.closed.Load(), "cache should not be closed")
		assert.NotNil(t, sc.closer, "closer func shall be not nil")
	})

	t.Run("override default", func(t *testing.T) {
		fn := func() cache.CacheInterface { return initFunc(t) }
		c, err := NewShardedCache(fn, WithOverrideDefaults(testShardNum))
		require.NoError(t, err)
		require.NotNil(t, c, "should return non")
		require.Implements(t, (*cache.CacheInterface)(nil), c, "result should implement cache.Interface")

		sc := typeCast(t, c)
		require.NotNil(t, sc, "should return shardedCache")
		assert.Equal(t, testShardNum, sc.shardNumber, "shard number should be default")
		assert.Equal(t, testShardNum, int64(len(sc.shards)), "shards should be init and equal to default")
		assert.False(t, sc.closed.Load(), "cache should not be closed")
		assert.NotNil(t, sc.closer, "closer func shall be not nil")
	})

	t.Run("override default with incorrect value", func(t *testing.T) {
		const shardNumber = -30
		fn := func() cache.CacheInterface { return initFunc(t) }
		c, err := NewShardedCache(fn, WithOverrideDefaults(shardNumber))
		require.NoError(t, err)
		require.NotNil(t, c, "should return non")
		require.Implements(t, (*cache.CacheInterface)(nil), c, "result should implement cache.Interface")

		sc := typeCast(t, c)
		require.NotNil(t, sc, "should return shardedCache")
		assert.Equal(t, defaultShardNumber, sc.shardNumber, "shard number should be default")
		assert.Equal(t, defaultShardNumber, int64(len(sc.shards)), "shards should be init and equal to default")
		assert.False(t, sc.closed.Load(), "cache should not be closed")
		assert.NotNil(t, sc.closer, "closer func shall be not nil")
	})

	t.Run("nil init func", func(t *testing.T) {
		c, err := NewShardedCache(nil)
		require.Error(t, err, "should return err for nil func")
		assert.Empty(t, c, "cache should not be initialized")
		assert.ErrorIs(t, err, cache2.NewErrInvalidValue("", cache2.ErrNilFunc, ""), "should return cache.ErrInvalidValue")
	})
}

func TestShardedCache_GetShardedCacheLen(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fn := func() cache.CacheInterface {
			m := mockCache.NewMockCacheInterface(t)
			m.On("GetLength").Return(int64(10), nil).Once()
			return m
		}

		c, err := NewShardedCache(fn)
		require.NoError(t, err)

		sc := typeCast(t, c)
		length, err := sc.getShardedCacheLen()
		require.NoError(t, err, "getShardedCacheLen should not return an error")
		assert.Equal(t, 10*defaultShardNumber, length, "Total length should equal sum of shard lengths")
	})

	t.Run("ErrorOnShard", func(t *testing.T) {
		fn := func() cache.CacheInterface {
			m := mockCache.NewMockCacheInterface(t)
			m.On("GetLength").Return(int64(0), assert.AnError)
			return m
		}

		c, err := NewShardedCache(fn, WithOverrideDefaults(1))
		require.NoError(t, err)

		sc := typeCast(t, c)
		val, err := sc.getShardedCacheLen()
		require.Error(t, err, "getShardedCacheLen should return an error if one shard fails")
		assert.Empty(t, val, "getShardedCacheLen should not return an empty value if one shard fails")
		assert.ErrorIs(t, err, assert.AnError, "Error message should be assert.AnError")
	})
}

func TestShardedCache_LoopShards(t *testing.T) {
	t.Run("AllShardsSuccessful", func(t *testing.T) {
		fn := func() cache.CacheInterface {
			m := mockCache.NewMockCacheInterface(t)
			m.On("GetKeys", mock.Anything).Return([]string{"key1", "key2"}, nil)
			return m
		}

		c, err := NewShardedCache(fn)
		require.NoError(t, err)

		sc := typeCast(t, c)
		resultCh := make(chan []string, defaultShardNumber)
		errCh := make(chan error, defaultShardNumber)

		ctx := context.Background()
		sc.loopShards(ctx, resultCh, errCh)

		close(resultCh)
		close(errCh)

		assert.Empty(t, errCh, "Error channel should be empty when no shard fails")
		var keys []string
		for res := range resultCh {
			keys = append(keys, res...)
		}
		assert.Len(t, keys, int(2*defaultShardNumber), "Result should contain keys from all shards")
	})

	t.Run("OneShardFails", func(t *testing.T) {
		fn := func() cache.CacheInterface {
			m := mockCache.NewMockCacheInterface(t)
			m.On("GetKeys", mock.Anything).Return(nil, assert.AnError)
			return m
		}

		c, err := NewShardedCache(fn, WithOverrideDefaults(1))
		require.NoError(t, err)

		sc := typeCast(t, c)
		resultCh := make(chan []string, defaultShardNumber)
		errCh := make(chan error, defaultShardNumber)

		ctx := context.Background()
		sc.loopShards(ctx, resultCh, errCh)

		close(resultCh)
		close(errCh)

		assert.NotEmpty(t, errCh, "Error channel should not be empty when one shard fails")
		err = <-errCh
		assert.ErrorIs(t, err, assert.AnError, "Error message should return assert.AnError")
	})
}

func TestShardedCache_GetShardNumber(t *testing.T) {
	t.Run("ValidKey", func(t *testing.T) {
		key := "testKey"
		shard := getShardNumber(key, defaultShardNumber)
		assert.GreaterOrEqual(t, shard, int64(0), "Shard number should be non-negative")
		assert.Less(t, shard, defaultShardNumber, "Shard number should be within range of shard count")
	})

	t.Run("EmptyKey", func(t *testing.T) {
		shard := getShardNumber("", defaultShardNumber)
		assert.Equal(t, fallbackShard, shard, "Empty key should default to fallback shard")
	})
}

// TODO: add impl tests
