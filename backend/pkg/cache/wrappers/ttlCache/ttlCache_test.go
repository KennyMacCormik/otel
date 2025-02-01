package ttlCache

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/cache"
	mockCache "github.com/KennyMacCormik/HerdMaster/pkg/cache/mocks"
	"github.com/KennyMacCormik/HerdMaster/pkg/log"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
	"time"
)

// TODO: add test that TTL works as expected

func getTtlCacheMock(t *testing.T, opts ...InitOptions) (*mockCache.MockInterface, cache.Interface) {
	c := mockCache.NewMockInterface(t)

	ttl, err := NewTtlCache(c, opts...)
	require.NoError(t, err, "expect no error with default configuration")
	require.NotNil(t, ttl, "expect result not nil with default configuration")
	require.Implements(t, (*cache.Interface)(nil), ttl, "result should implement cache.Interface")

	return c, ttl
}

func typeAssertion(t *testing.T, c cache.Interface) *ttlCache {
	cacheImpl, ok := c.(*ttlCache)
	require.True(t, ok, "expect result to be of type *ttlCache")
	require.NotNil(t, cacheImpl, "expect result to be not nil")
	return cacheImpl
}

func TestTtlCache_New(t *testing.T) {
	c := mockCache.NewMockInterface(t)

	t.Run("default", func(t *testing.T) {
		ttl, err := NewTtlCache(c)
		require.NoError(t, err, "expect no error with default configuration")
		assert.NotNil(t, ttl, "expect result not nil with default configuration")
		assert.Implements(t, (*cache.Interface)(nil), ttl, "result should implement cache.Interface")

		cacheImpl := typeAssertion(t, ttl)
		assert.Equal(t, defaultTTL, cacheImpl.ttl, "expect defaultTTL")
		assert.Equal(t, defaultTickerTTL, cacheImpl.tickerTTL, "expect defaultTickerTTL")
		assert.Equal(t, defaultGetKeysTTL, cacheImpl.getKeysTTL, "expect defaultGetKeysTTL")
		assert.Equal(t, defaultDeleteExpiredKeysTTL, cacheImpl.deleteExpiredKeysTTL, "expect defaultDeleteExpiredKeysTTL")
		assert.Equal(t, defaultSkewPercent, cacheImpl.skewPercent, "expect defaultSkewPercent")
	})

	t.Run("with override default", func(t *testing.T) {
		testTTL := 1 * time.Hour
		testSkew := 50
		ttl, err := NewTtlCache(c,
			WithOverrideDefaults(testTTL, testTTL, testTTL, testTTL, testSkew),
		)
		require.NoError(t, err, "expect no error with default configuration")
		assert.NotNil(t, ttl, "expect result not nil with default configuration")
		assert.Implements(t, (*cache.Interface)(nil), ttl, "result should implement cache.Interface")

		cacheImpl := typeAssertion(t, ttl)
		assert.Equal(t, testTTL, cacheImpl.ttl, "expect testTTL")
		assert.Equal(t, testTTL, cacheImpl.tickerTTL, "expect testTTL")
		assert.Equal(t, testTTL, cacheImpl.getKeysTTL, "expect testTTL")
		assert.Equal(t, testTTL, cacheImpl.deleteExpiredKeysTTL, "expect testTTL")
		assert.Equal(t, testSkew, cacheImpl.skewPercent, "expect testSkew")
	})

	t.Run("with incorrect override default", func(t *testing.T) {
		testTTL := -1 * time.Hour
		testSkew := -50
		ttl, err := NewTtlCache(c,
			WithOverrideDefaults(testTTL, testTTL, testTTL, testTTL, testSkew),
		)
		require.NoError(t, err, "expect no error with default configuration")
		assert.NotNil(t, ttl, "expect result not nil with default configuration")
		assert.Implements(t, (*cache.Interface)(nil), ttl, "result should implement cache.Interface")

		cacheImpl := typeAssertion(t, ttl)
		assert.Equal(t, defaultTTL, cacheImpl.ttl, "expect defaultTTL")
		assert.Equal(t, defaultTickerTTL, cacheImpl.tickerTTL, "expect defaultTickerTTL")
		assert.Equal(t, defaultGetKeysTTL, cacheImpl.getKeysTTL, "expect defaultGetKeysTTL")
		assert.Equal(t, defaultDeleteExpiredKeysTTL, cacheImpl.deleteExpiredKeysTTL, "expect defaultDeleteExpiredKeysTTL")
		assert.Equal(t, defaultSkewPercent, cacheImpl.skewPercent, "expect defaultSkewPercent")
	})

	t.Run("with logger", func(t *testing.T) {
		lg, _ := log.ConfigureLogger(log.WithOutput(io.Discard))
		ttl, err := NewTtlCache(c,
			WithLogger(lg),
		)
		require.NoError(t, err, "expect no error with default configuration")
		assert.NotNil(t, ttl, "expect result not nil with default configuration")
		assert.Implements(t, (*cache.Interface)(nil), ttl, "result should implement cache.Interface")

		cacheImpl := typeAssertion(t, ttl)
		assert.Equal(t, lg, cacheImpl.lg, "expect same logger")
	})

	t.Run("with invalid logger", func(t *testing.T) {
		ttl, err := NewTtlCache(c,
			WithLogger(nil),
		)
		require.NoError(t, err, "expect no error with default configuration")
		assert.NotNil(t, ttl, "expect result not nil with default configuration")
		assert.Implements(t, (*cache.Interface)(nil), ttl, "result should implement cache.Interface")

		cacheImpl := typeAssertion(t, ttl)
		assert.NotNil(t, cacheImpl.lg, "expect working logger")
	})

	t.Run("with nil cache", func(t *testing.T) {
		ttl, err := NewTtlCache(nil)
		require.Error(t, err, "expect an error with with nil cache")
		assert.ErrorIs(t, err, cache.NewErrInvalidValue("", cache.ErrNil, ""), "expect err be ErrInvalidValue")
		assert.Nil(t, ttl, "result should be nil with nil cache")
	})
}

func TestTtlCache_GetLength(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		const length = 1
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().GetLength().Return(length, nil)

		i, err := ttl.GetLength()
		require.NoError(t, err, "expect no error")
		assert.Equal(t, i, length, "expect returned value equal length")
	})

	t.Run("negative", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().GetLength().Return(0, assert.AnError)

		i, err := ttl.GetLength()
		require.Error(t, err, "expect an error")
		assert.Equal(t, i, 0, "expect returned value equal 0 in case of error")
		assert.ErrorIs(t, err, assert.AnError, "expect an error of type assert.AnError")
	})

	t.Run("cache closed", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Close(mock.Anything).Return(nil)
		err := ttl.Close(context.Background())
		require.NoError(t, err, "expect close successful")

		i, err := ttl.GetLength()
		require.Error(t, err, "expect an error")
		assert.Equal(t, i, 0, "expect returned value equal 0 in case of error")
		assert.ErrorIs(t, err, cache.ErrCacheClosed, "expect error to be cache.ErrCacheClosed")
	})
}

func TestTtlCache_Close(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Close(mock.Anything).Return(nil)

		t.Run("first close", func(t *testing.T) {
			err := ttl.Close(context.Background())
			require.NoError(t, err, "expect no error with first close")
		})

		t.Run("second close", func(t *testing.T) {
			err := ttl.Close(context.Background())
			require.NoError(t, err, "expect no error with second close")
		})
	})

	t.Run("negative", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Close(mock.Anything).Return(assert.AnError)

		t.Run("first close", func(t *testing.T) {
			err := ttl.Close(context.Background())
			require.Error(t, err, "expect error")
			assert.ErrorIs(t, err, assert.AnError, "expect error to be assert.AnError")
		})

		t.Run("second close", func(t *testing.T) {
			err := ttl.Close(context.Background())
			require.NoError(t, err, "expect no error with second close")
		})
	})
}

func TestTtlCache_GetKeys(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().GetKeys(mock.Anything).Return([]string{"key1", "key2"}, nil)

		val, err := ttl.GetKeys(context.Background())
		require.NoError(t, err, "expect no error with default configuration")
		assert.Equal(t, []string{"key1", "key2"}, val, "expect result not nil with default configuration")
	})
	t.Run("Negative", func(t *testing.T) {
		t.Run("return err", func(t *testing.T) {
			c, ttl := getTtlCacheMock(t)
			c.EXPECT().GetKeys(mock.Anything).Return(nil, assert.AnError)

			val, err := ttl.GetKeys(context.Background())
			require.Error(t, err, "expect an error")
			assert.Nil(t, val, "expect nil result with error")
			assert.ErrorIs(t, err, assert.AnError, "expect error to be assert.AnError")
		})

		t.Run("nil ctx", func(t *testing.T) {
			_, ttl := getTtlCacheMock(t)

			val, err := ttl.GetKeys(nil)
			require.Error(t, err, "expect an error")
			assert.Nil(t, val, "expect nil result with error")
			assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", nil), "expect error to be ErrCtx")
		})

		t.Run("cache closed", func(t *testing.T) {
			c, ttl := getTtlCacheMock(t)
			c.EXPECT().Close(mock.Anything).Return(nil)

			err := ttl.Close(context.Background())
			require.NoError(t, err, "expect no error with cache closed")

			val, err := ttl.GetKeys(context.Background())
			require.Error(t, err, "expect an error")
			assert.Nil(t, val, "expect nil result with error")
			assert.ErrorIs(t, err, cache.ErrCacheClosed, "expect error to be assert.AnError")
		})
	})

}

func TestTtlCache_Get(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		value1 := cacheEntry{"value1", time.Now().Add(10 * time.Second)}
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Get(mock.Anything, "key1").Return(&value1, nil)
		val, err := ttl.Get(context.Background(), "key1")
		require.NoError(t, err, "expect no error with default configuration")
		assert.Equal(t, value1.Value, val, "expect result and value match")
	})

	t.Run("negative", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Get(mock.Anything, "key1").Return(nil, assert.AnError)
		val, err := ttl.Get(context.Background(), "key1")
		require.Error(t, err, "expect error")
		assert.Nil(t, val, "expect nil result with error")
		assert.ErrorIs(t, err, assert.AnError, "expect error to be assert.AnError")
	})

	t.Run("expired", func(t *testing.T) {
		value1 := cacheEntry{"value1", time.Now().Add(-10 * time.Second)}
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Get(mock.Anything, "key1").Return(&value1, nil)
		val, err := ttl.Get(context.Background(), "key1")
		require.Error(t, err, "expect err with expired record")
		assert.Nil(t, val, "expect nil result with error")
		assert.ErrorIs(t, err, NewErrTimeout("", "", time.Now()), "expect error to be cache.ErrTimeout")
	})

	t.Run("closed", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Close(mock.Anything).Return(nil)
		err := ttl.Close(context.Background())
		require.NoError(t, err, "expect no error with cache closed")
		val, err := ttl.Get(context.Background(), "key1")
		require.Error(t, err, "expect error")
		assert.Nil(t, val, "expect nil result with error")
		assert.ErrorIs(t, err, cache.ErrCacheClosed, "expect error to be cache.ErrCacheClosed")
	})

	t.Run("nil ctx", func(t *testing.T) {
		_, ttl := getTtlCacheMock(t)
		val, err := ttl.Get(nil, "key1")
		require.Error(t, err, "expect error")
		assert.Nil(t, val, "expect nil result with error")
		assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", nil), "expect error to be ErrCtx")
	})

	t.Run("empty key", func(t *testing.T) {
		_, ttl := getTtlCacheMock(t)
		val, err := ttl.Get(context.Background(), "")
		require.Error(t, err, "expect error")
		assert.Nil(t, val, "expect nil result with error")
		assert.ErrorIs(t, err, cache.NewErrInvalidValue("", cache.ErrEmptyString, ""), "expect error to be cache.ErrInvalidValue")
	})

	t.Run("typecast failure", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Get(mock.Anything, "key1").Return("value1", nil)
		val, err := ttl.Get(context.Background(), "key1")
		require.Error(t, err, "expect error")
		assert.Nil(t, val, "expect nil result with error")
		assert.ErrorIs(t, err, cache.NewErrTypeCastFailed("", "", ""), "expect error to be cache.ErrTypeCastFailed")
	})
}

func TestTtlCache_Set(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Set(mock.Anything, "key1", mock.Anything).Return(nil)
		err := ttl.Set(context.Background(), "key1", "value1")
		require.NoError(t, err, "expect no error with default configuration")
	})

	t.Run("negative", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Set(mock.Anything, "key1", mock.Anything).Return(assert.AnError)
		err := ttl.Set(context.Background(), "key1", "value1")
		require.Error(t, err, "expect an error")
		assert.ErrorIs(t, err, assert.AnError, "expect error to be assert.AnError")
	})

	t.Run("closed", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Close(mock.Anything).Return(nil)
		err := ttl.Close(context.Background())
		require.NoError(t, err, "expect no error with cache closed")
		err = ttl.Set(context.Background(), "key1", "value1")
		require.Error(t, err, "expect an error")
		assert.ErrorIs(t, err, cache.ErrCacheClosed, "expect error to be cache.ErrCacheClosed")
	})

	t.Run("nil ctx", func(t *testing.T) {
		_, ttl := getTtlCacheMock(t)
		err := ttl.Set(nil, "key1", "value1")
		require.Error(t, err, "expect an error")
		assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", nil), "expect error to be ErrCtx")
	})

	t.Run("empty key", func(t *testing.T) {
		_, ttl := getTtlCacheMock(t)
		err := ttl.Set(context.Background(), "", "value1")
		require.Error(t, err, "expect an error")
		assert.ErrorIs(t, err, cache.NewErrInvalidValue("", cache.ErrEmptyString, ""), "expect error to be cache.ErrInvalidValue")
	})

	t.Run("empty value", func(t *testing.T) {
		_, ttl := getTtlCacheMock(t)
		err := ttl.Set(context.Background(), "key1", nil)
		require.Error(t, err, "expect an error")
		assert.ErrorIs(t, err, cache.NewErrInvalidValue("", cache.ErrNil, ""), "expect error to be cache.ErrInvalidValue")
	})
}

func TestTtlCache_Delete(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Delete(mock.Anything, "key1").Return(nil)
		err := ttl.Delete(context.Background(), "key1")
		require.NoError(t, err, "expect no error with default configuration")
	})

	t.Run("negative", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Delete(mock.Anything, "key1").Return(assert.AnError)
		err := ttl.Delete(context.Background(), "key1")
		require.Error(t, err, "expect an error")
		assert.ErrorIs(t, err, assert.AnError, "expect error to be assert.AnError")
	})

	t.Run("closed", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().Close(mock.Anything).Return(nil)
		err := ttl.Close(context.Background())
		require.NoError(t, err, "expect no error with cache closed")
		err = ttl.Delete(context.Background(), "key1")
		require.Error(t, err, "expect an error")
		assert.ErrorIs(t, err, cache.ErrCacheClosed, "expect error to be cache.ErrCacheClosed")
	})

	t.Run("nil ctx", func(t *testing.T) {
		_, ttl := getTtlCacheMock(t)
		err := ttl.Delete(nil, "key1")
		require.Error(t, err, "expect an error")
		assert.ErrorIs(t, err, cache.NewErrNilOrErrCtx("", nil), "expect error to be ErrCtx")
	})

	t.Run("empty key", func(t *testing.T) {
		_, ttl := getTtlCacheMock(t)
		err := ttl.Delete(context.Background(), "")
		require.Error(t, err, "expect an error")
		assert.ErrorIs(t, err, cache.NewErrInvalidValue("", cache.ErrEmptyString, ""), "expect error to be cache.ErrInvalidValue")
	})
}

func TestErrTimeout(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		key := "test-key"
		callerInfo := "test-caller"
		expirationTime := time.Now()

		errTimeout := NewErrTimeout(key, callerInfo, expirationTime)

		require.NotNil(t, errTimeout, "NewErrTimeout should return a non-nil error object")
		assert.Equal(t, key, errTimeout.GetKey(), "key should match the input value")
		assert.Equal(t, callerInfo, errTimeout.GetCallerInfo(), "callerInfo should match the input value")
		assert.Equal(t, expirationTime, errTimeout.GetTtl(), "expirationTime should match the input value")
		assert.Equal(t, ErrExpired, errTimeout.Unwrap(), "signalErr should match ErrExpired")
	})

	t.Run("error message formatting", func(t *testing.T) {
		key := "test-key"
		callerInfo := "test-caller"
		expirationTime := time.Now()

		errTimeout := NewErrTimeout(key, callerInfo, expirationTime)
		expectedMessage := fmt.Sprintf("%s: key [%s] ttl [%s]: %s", callerInfo, key, expirationTime, ErrExpired)

		assert.EqualError(t, errTimeout, expectedMessage, "Error message should match the expected format")
	})

	t.Run("matching ErrTimeout type", func(t *testing.T) {
		key := "test-key"
		callerInfo := "test-caller"
		expirationTime := time.Now()

		errTimeout := NewErrTimeout(key, callerInfo, expirationTime)
		targetErr := &ErrTimeout{}

		assert.True(t, errors.Is(errTimeout, targetErr), "Is should return true for ErrTimeout type")
	})

	t.Run("non-matching error type", func(t *testing.T) {
		key := "test-key"
		callerInfo := "test-caller"
		expirationTime := time.Now()

		errTimeout := NewErrTimeout(key, callerInfo, expirationTime)
		targetErr := errors.New("random error")

		assert.False(t, errors.Is(errTimeout, targetErr), "Is should return false for non-ErrTimeout type")
	})

	t.Run("unwrap to ErrExpired", func(t *testing.T) {
		key := "test-key"
		callerInfo := "test-caller"
		expirationTime := time.Now()

		errTimeout := NewErrTimeout(key, callerInfo, expirationTime)

		assert.Equal(t, ErrExpired, errTimeout.Unwrap(), "Unwrap should return ErrExpired")
	})
}

func TestTtlCache_GetTtl(t *testing.T) {
	var iterations = 100
	_, ttl := getTtlCacheMock(t)
	cacheImpl, ok := ttl.(*ttlCache)
	require.True(t, ok, "expect result to be of type *ttlCache")
	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration %d", i+1), func(t *testing.T) {
			t.Parallel()
			require.WithinDuration(t,
				time.Now().Add(cacheImpl.ttl),
				cacheImpl.getTtl(),
				cacheImpl.ttl*time.Duration(cacheImpl.skewPercent)/100,
				"getTtl should return a time within default TTL duration",
			)
		})
	}
}

func TestTtlCache_ExpireCache(t *testing.T) {
	t.Run("expireCache runs and stops gracefully", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t, WithOverrideDefaults(3*time.Second, 500*time.Millisecond, 1*time.Second, 100*time.Millisecond, 10))
		c.EXPECT().GetKeys(mock.Anything).Return([]string{"key1"}, nil)
		c.EXPECT().Get(mock.Anything, "key1").Return(
			&cacheEntry{
				"expiredValue",
				time.Now().Add(1 * time.Second),
			},
			nil,
		)
		c.EXPECT().Delete(mock.Anything, "key1").Return(nil)
		c.EXPECT().Close(mock.Anything).Return(nil).Once()

		cacheImpl := typeAssertion(t, ttl)
		time.Sleep(2 * time.Second)

		err := cacheImpl.Close(context.Background())
		require.NoError(t, err)
		// Validate ticker stopped without panic or error
		require.NotPanics(t, func() {
			cacheImpl.expireCache()
		}, "expireCache should not panic when the close channel is closed")
	})
}

func TestTtlCache_DeleteExpiredKey(t *testing.T) {
	t.Run("deleteExpiredKey deletes expired key successfully", func(t *testing.T) {
		r, w, _ := os.Pipe()
		defer func() {
			_ = r.Close()
			_ = w.Close()
		}()

		lg, _ := log.ConfigureLogger(log.WithOutput(w))
		c, ttl := getTtlCacheMock(t, WithLogger(lg))
		c.EXPECT().Get(mock.Anything, "expiredKey").Return(
			&cacheEntry{
				"expiredValue",
				time.Now().Add(-5 * time.Second),
			},
			nil,
		).Once()
		c.EXPECT().Delete(mock.Anything, "expiredKey").Return(nil).Once()

		cacheImpl := typeAssertion(t, ttl)
		cacheImpl.deleteExpiredKey("expiredKey")

		_ = w.Close()
		out := &bytes.Buffer{}
		_, _ = io.Copy(out, r)

		require.Equal(t, "", out.String(), "expect not to log anything")
	})

	t.Run("deleteExpiredKey logs error on failure", func(t *testing.T) {
		r, w, _ := os.Pipe()
		defer func() {
			_ = r.Close()
			_ = w.Close()
		}()

		lg, _ := log.ConfigureLogger(log.WithOutput(w))
		c, ttl := getTtlCacheMock(t, WithLogger(lg))
		c.EXPECT().Get(mock.Anything, "failedKey").Return(nil, assert.AnError).Once()

		cacheImpl := typeAssertion(t, ttl)
		cacheImpl.deleteExpiredKey("failedKey")

		_ = w.Close()
		out := &bytes.Buffer{}
		_, _ = io.Copy(out, r)

		require.Contains(t, out.String(), assert.AnError.Error(), "expect not to log anything")
	})
}

func TestTtlCache_GetImplKeys(t *testing.T) {
	t.Run("getImplKeys retrieves keys successfully", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().GetKeys(mock.Anything).Return([]string{"key1", "key2"}, nil).Once()

		cacheImpl := typeAssertion(t, ttl)
		keys, err := cacheImpl.getImplKeys()

		require.NoError(t, err, "getImplKeys should not return an error")
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys, "getImplKeys should return expected keys")
	})

	t.Run("getImplKeys handles error gracefully", func(t *testing.T) {
		c, ttl := getTtlCacheMock(t)
		c.EXPECT().GetKeys(mock.Anything).Return(nil, assert.AnError).Once()

		cacheImpl := typeAssertion(t, ttl)
		keys, err := cacheImpl.getImplKeys()

		assert.Error(t, err, "getImplKeys should return an error if GetKeys fails")
		assert.ErrorIs(t, err, assert.AnError, "getImplKeys should return an error if GetKeys fails")
		assert.Nil(t, keys, "getImplKeys should return nil keys if GetKeys fails")
	})
}

func TestTtlCache_TtlExpired(t *testing.T) {
	t.Run("ttlExpired returns true for past time", func(t *testing.T) {
		pastTime := time.Now().Add(-1 * time.Second)
		assert.True(t, ttlExpired(pastTime), "ttlExpired should return true for a past time")
	})

	t.Run("ttlExpired returns false for future time", func(t *testing.T) {
		futureTime := time.Now().Add(1 * time.Second)
		assert.False(t, ttlExpired(futureTime), "ttlExpired should return false for a future time")
	})
}
