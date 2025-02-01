package syncMap

import (
	"context"
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

const testKeyCapacity = 128
const testTimeout = time.Second * 1

func initSyncMap() cache.Interface {
	return NewSyncMapCache(WithOverrideDefaults(testKeyCapacity, testTimeout))
}

func typeCastCache(t *testing.T, c cache.Interface) *syncMap {
	sm, ok := c.(*syncMap)
	require.True(t, ok, "type cast shall succeed")
	return sm
}

func TestSyncMap_New(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		sm := NewSyncMapCache()
		require.NotNil(t, sm, "cache shall be created")
		assert.Implements(t, (*cache.Interface)(nil), sm, "cache shall implement cache interface")
		impl := typeCastCache(t, sm)
		assert.Equal(t, defaultKeyCapacity, impl.keyCapacity, "cache shall have keyCapacity = defaultKeyCapacity")
		assert.Equal(t, defaultTimeout, impl.timeout, "cache shall have timeout = defaultTimeout")
	})

	t.Run("with override defaults", func(t *testing.T) {
		sm := NewSyncMapCache(WithOverrideDefaults(testKeyCapacity, testTimeout))
		require.NotNil(t, sm, "cache shall be created")
		assert.Implements(t, (*cache.Interface)(nil), sm, "cache shall implement cache interface")

		impl := typeCastCache(t, sm)
		assert.Equal(t, testKeyCapacity, impl.keyCapacity, "cache shall have keyCapacity = testKeyCapacity")
		assert.Equal(t, testTimeout, impl.timeout, "cache shall have timeout = testTimeout")
	})

	t.Run("with incorrect settings", func(t *testing.T) {
		sm := NewSyncMapCache(WithOverrideDefaults(-testKeyCapacity, -testTimeout))
		require.NotNil(t, sm, "cache shall be created")
		assert.Implements(t, (*cache.Interface)(nil), sm, "cache shall implement cache interface")

		impl := typeCastCache(t, sm)
		assert.Equal(t, defaultKeyCapacity, impl.keyCapacity, "cache shall have keyCapacity = defaultKeyCapacity")
		assert.Equal(t, defaultTimeout, impl.timeout, "cache shall have timeout = defaultTimeout")
	})
}

func TestSyncMap_Set(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		err := sm.Set(context.Background(), "key1", "value1")
		require.NoError(t, err, "Shall return no error for valid input")

		length := 0
		sm.m.Range(func(_, _ any) bool {
			length++
			return true
		})
		assert.Equal(t, 1, length, "Cache shall store supplied value")
	})

	t.Run("cache closed", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		err := sm.Close(context.Background())
		require.NoError(t, err, "Shall return no error for Close()")

		err = sm.Set(context.Background(), "key1", "value1")
		require.Error(t, err, "Shall return error for closed cache")
		assert.ErrorIs(t, err, cache.ErrCacheClosed, "Shall return cache.ErrCacheClosed")

		length := 0
		sm.m.Range(func(_, _ any) bool {
			length++
			return true
		})
		assert.Equal(t, 0, length, "Cache shall NOT store supplied value")
	})

	t.Run("incorrect ctx", func(t *testing.T) {
		t.Run("nil ctx", func(t *testing.T) {
			sm := typeCastCache(t, initSyncMap())

			err := sm.Set(nil, "key1", "value1")
			require.Error(t, err, "Shall return error for nil ctx")
			assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", nil), "Shall return cache.ErrCtx")

			length := 0
			sm.m.Range(func(_, _ any) bool {
				length++
				return true
			})
			assert.Equal(t, 0, length, "Cache shall NOT store supplied value")
		})

		t.Run("expired ctx", func(t *testing.T) {
			sm := typeCastCache(t, initSyncMap())

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			time.Sleep(3 * time.Nanosecond)

			err := sm.Set(ctx, "key1", "value1")
			require.Error(t, err, "Shall return error for closed ctx")
			assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", ctx), "Shall return cache.ErrCtx")

			length := 0
			sm.m.Range(func(_, _ any) bool {
				length++
				return true
			})
			assert.Equal(t, 0, length, "Cache shall NOT store supplied value")
		})
	})

	t.Run("empty key", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		err := sm.Set(context.Background(), "", "value1")
		require.Error(t, err, "Shall return error empty key")
		assert.ErrorIs(t, err, cache.NewErrInvalidValue("", cache.ErrEmptyString, ""), "Shall return cache.ErrInvalidValue")

		length := 0
		sm.m.Range(func(_, _ any) bool {
			length++
			return true
		})
		assert.Equal(t, 0, length, "Cache shall NOT store supplied value")
	})

	t.Run("nil value", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		err := sm.Set(context.Background(), "key1", nil)
		require.Error(t, err, "Shall return error empty key")
		assert.ErrorIs(t, err, cache.NewErrInvalidValue("", cache.ErrNil, ""), "Shall return cache.ErrInvalidValue")

		length := 0
		sm.m.Range(func(_, _ any) bool {
			length++
			return true
		})
		assert.Equal(t, 0, length, "Cache shall NOT store supplied value")
	})
}

func TestSyncMap_Delete(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		err := sm.Set(context.Background(), "key1", "value1")
		require.NoError(t, err, "Shall return no error for valid input")

		err = sm.Delete(context.Background(), "key1")
		require.NoError(t, err, "Shall return no error for valid input")

		length := 0
		sm.m.Range(func(_, _ any) bool {
			length++
			return true
		})
		assert.Equal(t, 0, length, "Cache shall delete value")
	})

	t.Run("negative", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		err := sm.Set(context.Background(), "key1", "value1")
		require.NoError(t, err, "Shall return no error for valid input")

		err = sm.Delete(context.Background(), "key2")
		require.NoError(t, err, "Shall return no error for valid input")

		length := 0
		sm.m.Range(func(_, _ any) bool {
			length++
			return true
		})
		assert.Equal(t, 1, length, "Cache shall retain stored value")
	})

	t.Run("closed", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		err := sm.Set(context.Background(), "key1", "value1")
		require.NoError(t, err, "Shall return no error for valid input")

		err = sm.Close(context.Background())
		require.NoError(t, err, "Shall return no error for Close()")

		err = sm.Delete(context.Background(), "key1")
		require.Error(t, err, "expect an error for closed cache")
		assert.ErrorIs(t, err, cache.ErrCacheClosed, "expect error to be cache.ErrCacheClosed")
	})

	t.Run("incorrect ctx", func(t *testing.T) {
		t.Run("nil ctx", func(t *testing.T) {
			sm := typeCastCache(t, initSyncMap())
			err := sm.Set(context.Background(), "key1", "value1")
			require.NoError(t, err, "Shall return no error for valid input")

			err = sm.Delete(nil, "key1")
			require.Error(t, err, "expect an error for closed cache")
			assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", nil), "expect error to be ErrCtx")
		})

		t.Run("expired ctx", func(t *testing.T) {
			sm := typeCastCache(t, initSyncMap())
			err := sm.Set(context.Background(), "key1", "value1")
			require.NoError(t, err, "Shall return no error for valid input")

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			time.Sleep(3 * time.Nanosecond)

			err = sm.Delete(ctx, "key1")
			require.Error(t, err, "expect an error for closed cache")
			assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", ctx), "expect error to be ErrCtx")
		})
	})

	t.Run("empty key", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		err := sm.Set(context.Background(), "key1", "value1")
		require.NoError(t, err, "Shall return no error for valid input")

		err = sm.Delete(context.Background(), "")
		require.Error(t, err, "expect an error for closed cache")
		assert.ErrorIs(t, err, cache.NewErrInvalidValue("", cache.ErrEmptyString, ""), "expect error to be cache.ErrInvalidValue")
	})
}

func TestSyncMap_Get(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		err := sm.Set(context.Background(), "key1", "value1")
		require.NoError(t, err, "Shall return no error for valid input")

		val, err := sm.Get(context.Background(), "key1")
		require.NoError(t, err, "Shall return no error for valid input")
		assert.Equal(t, "value1", val, "Get value shall match set value")

		length := 0
		sm.m.Range(func(_, _ any) bool {
			length++
			return true
		})
		assert.Equal(t, 1, length, "Cache shall not delete value")
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("cache full", func(t *testing.T) {
			sm := typeCastCache(t, initSyncMap())
			err := sm.Set(context.Background(), "key1", "value1")
			require.NoError(t, err, "Shall return no error for valid input")

			val, err := sm.Get(context.Background(), "key2")
			require.Error(t, err, "Shall return no error for nonexistent key")
			assert.ErrorIs(t, err, cache.ErrNotFound, "expect error to be cache.ErrNotFound")
			assert.Empty(t, val, "Get value shall be nil")

			length := 0
			sm.m.Range(func(_, _ any) bool {
				length++
				return true
			})
			assert.Equal(t, 1, length, "Cache shall retain stored value")
		})

		t.Run("cache empty", func(t *testing.T) {
			sm := typeCastCache(t, initSyncMap())

			val, err := sm.Get(context.Background(), "key2")
			require.Error(t, err, "Shall return no error for nonexistent key")
			assert.ErrorIs(t, err, cache.ErrNotFound, "expect error to be cache.ErrNotFound")
			assert.Empty(t, val, "Get value shall be nil")

			length := 0
			sm.m.Range(func(_, _ any) bool {
				length++
				return true
			})
			assert.Equal(t, 0, length, "Cache shall remain empty")
		})
	})

	t.Run("closed", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		err := sm.Close(context.Background())
		require.NoError(t, err, "Shall return no error close operation")

		val, err := sm.Get(context.Background(), "key2")
		require.Error(t, err, "Shall return no error for nonexistent key")
		assert.ErrorIs(t, err, cache.ErrCacheClosed, "expect error to be cache.ErrCacheClosed")
		assert.Empty(t, val, "Get value shall be nil")
	})

	t.Run("incorrect ctx", func(t *testing.T) {
		t.Run("nil ctx", func(t *testing.T) {
			sm := typeCastCache(t, initSyncMap())

			val, err := sm.Get(nil, "key2")
			require.Error(t, err, "expect an error for nil ctx")
			assert.Empty(t, val, "Get value shall be nil")
			assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", nil), "expect error to be ErrCtx")
		})

		t.Run("expired ctx", func(t *testing.T) {
			sm := typeCastCache(t, initSyncMap())

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			time.Sleep(3 * time.Nanosecond)

			val, err := sm.Get(ctx, "key2")
			require.Error(t, err, "expect an error for closed ctx")
			assert.Empty(t, val, "Get value shall be nil")
			assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", ctx), "expect error to be ErrCtx")
		})
	})

	t.Run("empty key", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		err := sm.Set(context.Background(), "key1", "value1")
		require.NoError(t, err, "Shall return no error for valid input")

		val, err := sm.Get(context.Background(), "")
		require.Error(t, err, "expect an error for closed cache")
		assert.Empty(t, val, "Get value shall be nil")
		assert.ErrorIs(t, err, cache.NewErrInvalidValue("", cache.ErrEmptyString, ""), "expect error to be cache.ErrInvalidValue")
	})
}

func TestSyncMap_Close(t *testing.T) {
	t.Run("close cache successfully", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		ctx := context.Background()
		err := sm.Close(ctx)
		require.NoError(t, err, "Close should not return an error for a valid context")
		assert.True(t, sm.closed.Load(), "Cache should be marked as closed")
	})

	t.Run("close cache with nil context", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		err := sm.Close(nil)
		require.NoError(t, err, "Close should handle nil context gracefully")
		assert.True(t, sm.closed.Load(), "Cache should be marked as closed")
	})

	t.Run("close cache multiple times", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		ctx := context.Background()
		err := sm.Close(ctx)
		require.NoError(t, err, "First Close call should not return an error")
		assert.True(t, sm.closed.Load(), "Cache should be marked as closed after the first call")

		err = sm.Close(ctx)
		assert.NoError(t, err, "Subsequent Close calls should not return an error")
		assert.True(t, sm.closed.Load(), "Cache should remain closed after subsequent calls")
	})

	t.Run("cache is cleared during Close", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		// Add some entries to verify clearance
		sm.m.Store("key1", "value1")
		sm.m.Store("key2", "value2")

		ctx := context.Background()
		err := sm.Close(ctx)
		require.NoError(t, err, "Close should not return an error")

		length := 0
		sm.m.Range(func(_, _ any) bool {
			length++
			return true
		})
		assert.Equal(t, 0, length, "Map should be cleared during Close")
	})
}

func TestSyncMap_GetLength(t *testing.T) {
	t.Run("get length of non-empty map", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		sm.m.Store("key1", "value1")
		sm.m.Store("key2", "value2")

		length, err := sm.GetLength()
		require.NoError(t, err, "GetLength should not return an error for a non-empty map")
		assert.Equal(t, 2, length, "GetLength should return the correct count of entries")
	})

	t.Run("get length of empty map", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		length, err := sm.GetLength()
		require.NoError(t, err, "GetLength should not return an error for an empty map")
		assert.Equal(t, 0, length, "GetLength should return 0 for an empty map")
	})

	t.Run("get length of closed map", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		ctx := context.Background()
		err := sm.Close(ctx)
		require.NoError(t, err, "Close should not return an error")

		length, err := sm.GetLength()
		require.Error(t, err, "GetLength should return an error for a closed map")
		assert.ErrorIs(t, err, cache.ErrCacheClosed, "GetLength should return an cache.ErrCacheClosed error")
		assert.Equal(t, 0, length, "GetLength should return 0 for a closed map")
	})
}

func TestSyncMap_GetKeys(t *testing.T) {
	t.Run("retrieve all keys successfully", func(t *testing.T) {
		s := initSyncMap()
		sm := typeCastCache(t, s)
		sm.m.Store("key1", "value1")
		sm.m.Store("key2", "value2")
		sm.m.Store("key3", "value3")

		ctx := context.Background()
		keys, err := s.GetKeys(ctx)
		require.NoError(t, err, "GetKeys should not return an error for a valid context")
		assert.ElementsMatch(t, []string{"key1", "key2", "key3"}, keys, "Keys should match the stored keys")
	})

	t.Run("context timeout during retrieval", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		sm.m.Store("key1", "value1")
		sm.m.Store("key2", "value2")

		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		// Simulate timeout
		time.Sleep(2 * testTimeout)
		keys, err := sm.GetKeys(ctx)
		require.Error(t, err, "GetKeys should return an error if the context is done")
		assert.Nil(t, keys, "Keys should be nil if context times out")
		assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", ctx), "GetKeys should return an ErrNilOrErrCtx")
	})

	t.Run("empty map", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		ctx := context.Background()
		keys, err := sm.GetKeys(ctx)
		require.NoError(t, err, "GetKeys should not return an error for an empty map")
		assert.Empty(t, keys, "Keys should be empty for an empty map")
	})

	t.Run("nil context", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		sm.m.Store("key1", "value1")

		keys, err := sm.GetKeys(nil)
		require.Error(t, err, "GetKeys should NOT handle nil context")
		assert.Nil(t, keys, "Keys should be nil if error returned")
		assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", nil), "GetKeys should return an ErrNilOrErrCtx")
	})

	t.Run("key type casting failure", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		sm.m.Store(123, "value")

		ctx := context.Background()
		keys, err := sm.GetKeys(ctx)
		assert.Error(t, err, "GetKeys should return an error if a key type is invalid")
		assert.Nil(t, keys, "Keys should be nil if there is a type casting failure")
		assert.ErrorIs(t, err, &cache.ErrTypeCastFailed{}, "Error should be of type ErrTypeCastFailed")
	})

	t.Run("cache closed", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		sm.m.Store(123, "value")

		ctx := context.Background()
		err := sm.Close(ctx)
		require.NoError(t, err, "Close should not return an error")

		keys, err := sm.GetKeys(ctx)
		assert.Error(t, err, "GetKeys should return an error if a key type is invalid")
		assert.Nil(t, keys, "Keys should be nil if there is a type casting failure")
		assert.ErrorIs(t, err, cache.ErrCacheClosed, "Error should be of type ErrCacheClosed")
	})
}

func TestSyncMap_ClearSyncMapWithTimeout(t *testing.T) {
	t.Run("successful map clearance within timeout", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		sm.m.Store("key1", "value1")
		sm.m.Store("key2", "value2")
		sm.m.Store("key3", "value3")

		ctx := context.Background()
		err := sm.clearSyncMapWithTimeout(ctx)
		require.NoError(t, err, "clearSyncMapWithTimeout should not return an error")

		length := 0
		sm.m.Range(func(_, _ any) bool {
			length++
			return true
		})
		assert.Equal(t, 0, length, "Map should be cleared")
	})

	t.Run("context timeout during map clearance", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		sm.m.Store("key1", "value1")
		sm.m.Store("key2", "value2")

		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		time.Sleep(2 * testTimeout)
		err := sm.clearSyncMapWithTimeout(ctx)
		assert.NoError(t, err, "no error is returned")

		length := 0
		sm.m.Range(func(_, _ any) bool {
			length++
			return true
		})
		assert.Equal(t, 0, length, "Map should be cleared")
	})

	t.Run("nil context ", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())
		sm.m.Store("key1", "value1")
		sm.m.Store("key2", "value2")

		err := sm.clearSyncMapWithTimeout(nil)
		require.NoError(t, err, "clearSyncMapWithTimeout should handle nil context gracefully")

		length := 0
		sm.m.Range(func(_, _ any) bool {
			length++
			return true
		})
		assert.Equal(t, 0, length, "Map should be cleared")
	})
}

func TestSyncMap_NormalizeCtx(t *testing.T) {
	t.Run("valid context", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		resultCtx, resultCancel := sm.normalizeCtx(ctx)
		if resultCancel != nil {
			defer resultCancel()
		}

		assert.Equal(t, ctx, resultCtx, "normalizeCtx should return same ctx")
	})

	t.Run("background context", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		ctx := context.Background()
		resultCtx, cancel := sm.normalizeCtx(ctx)
		if cancel != nil {
			defer cancel()
		}

		assert.Equal(t, ctx, resultCtx, "normalizeCtx should return context.Background()")
	})

	t.Run("nil context", func(t *testing.T) {
		sm := typeCastCache(t, initSyncMap())

		resultCtx, cancel := sm.normalizeCtx(nil)
		if cancel != nil {
			defer cancel()
		}

		assert.NotNil(t, resultCtx, "normalizeCtx should return a new context if input is nil")
		assert.NotEqual(t, nil, resultCtx, "result ctx should not match nil")
		assert.NotEqual(t, context.Background(), resultCtx, "result ctx should not match nil context.Background()")
	})

	t.Run("expired context", func(t *testing.T) {
		sm := &syncMap{
			timeout:     1 * time.Second,
			keyCapacity: 128,
		}

		expiredCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
		defer cancel()

		resultCtx, cancelNew := sm.normalizeCtx(expiredCtx)
		if cancel != nil {
			defer cancelNew()
		}

		assert.NotNil(t, resultCtx, "normalizeCtx should replace expired context with a new one")
		assert.NotEqual(t, expiredCtx, resultCtx, "normalizeCtx should return a new context if the input context is expired")
	})
}

func TestSyncMap_Concurrent(t *testing.T) {
	const numConcurrent = 100
	sm := typeCastCache(t, initSyncMap())

	t.Run("set", func(t *testing.T) {
		for i := 0; i < numConcurrent; i++ {
			t.Run(fmt.Sprintf("concurrent set %d", i), func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
				defer cancel()
				key := "key" + strconv.Itoa(i)
				value := "value" + strconv.Itoa(i)
				err := sm.Set(ctx, key, value)
				require.NoError(t, err, "Set should not return an error")
			})
		}
	})

	t.Run("get", func(t *testing.T) {
		for i := 0; i < numConcurrent; i++ {
			t.Run(fmt.Sprintf("concurrent get %d", i), func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
				defer cancel()
				key := "key" + strconv.Itoa(i)
				value := "value" + strconv.Itoa(i)
				res, err := sm.Get(ctx, key)
				require.NoError(t, err, "Set should not return an error")
				assert.Equal(t, value, res, "Get should return the same value")
			})
		}
	})

	t.Run("del", func(t *testing.T) {
		for i := 0; i < numConcurrent; i++ {
			t.Run(fmt.Sprintf("concurrent del %d", i), func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
				defer cancel()
				key := "key" + strconv.Itoa(i)
				err := sm.Delete(ctx, key)
				require.NoError(t, err, "Delete should not return an error")
			})
		}
	})
}
