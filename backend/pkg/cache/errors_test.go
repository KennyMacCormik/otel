package cache

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestNewErrTypeCastFailed(t *testing.T) {
	const callerInfo = "TestCaller"
	key := "testKey"
	value := "testValue"

	err := NewErrTypeCastFailed(key, value, callerInfo)

	require.NotNil(t, err, "NewErrTypeCastFailed should return a non-nil error")
	assert.Equal(t, callerInfo, err.callerInfo, "CallerInfo should match the input")
	assert.Equal(t, key, err.key, "Key should match the input")
	assert.Equal(t, value, err.value, "Value should match the input")
	assert.ErrorIs(t, err, ErrTypeCast, "Error should wrap ErrTypeCast")
}

func TestErrTypeCastFailed_error(t *testing.T) {
	t.Run("key nil", func(t *testing.T) {
		const callerInfo = "TestCaller"
		value := "testValue"

		err := NewErrTypeCastFailed(nil, value, callerInfo)

		assert.Equal(t, nil, err.key, "Error message should include caller info")
		assert.Equal(t, value, err.value, "Error message should include caller info")
		assert.Equal(t, callerInfo, err.callerInfo, "Error message should include caller info")
		assert.Equal(t,
			fmt.Errorf("%w: %w", ErrNil, ErrTypeCast),
			err.underlyingErr,
			"Error message should include caller info",
		)
		assert.Equal(t,
			fmt.Errorf("%s: key err: %w: %w", err.callerInfo, ErrNil, ErrTypeCast).Error(),
			err.errStr,
			"Error message should include caller info",
		)
	})

	t.Run("key empty interface", func(t *testing.T) {
		const callerInfo = "TestCaller"
		var key *int
		value := "testValue"

		err := NewErrTypeCastFailed(key, value, callerInfo)

		assert.Equal(t, key, err.key, "Error message should include caller info")
		assert.Equal(t, value, err.value, "Error message should include caller info")
		assert.Equal(t, callerInfo, err.callerInfo, "Error message should include caller info")
		assert.Equal(t,
			fmt.Errorf("%w: %w", ErrNilPointerOrNilInterface, ErrTypeCast),
			err.underlyingErr,
			"Error message should include caller info",
		)
		assert.Equal(t,
			fmt.Errorf("%s: key err: %w: %w", err.callerInfo, ErrNilPointerOrNilInterface, ErrTypeCast).Error(),
			err.errStr,
			"Error message should include caller info",
		)
	})

	t.Run("val nil", func(t *testing.T) {
		const callerInfo = "TestCaller"
		key := "key1"

		err := NewErrTypeCastFailed(key, nil, callerInfo)

		assert.Equal(t, key, err.key, "Error message should include caller info")
		assert.Equal(t, nil, err.value, "Error message should include caller info")
		assert.Equal(t, callerInfo, err.callerInfo, "Error message should include caller info")
		assert.Equal(t,
			fmt.Errorf("%w: %w", ErrNil, ErrTypeCast),
			err.underlyingErr,
			"Error message should include caller info",
		)
		assert.Equal(t,
			fmt.Errorf("%s: key %v: value err: %w: %w", err.callerInfo, err.key, ErrNil, ErrTypeCast).Error(),
			err.errStr,
			"Error message should include caller info",
		)
	})

	t.Run("val empty interface", func(t *testing.T) {
		const callerInfo = "TestCaller"
		var value any = (*int)(nil)
		key := "testKey"

		err := NewErrTypeCastFailed(key, value, callerInfo)

		assert.Equal(t, key, err.key, "Error message should include caller info")
		assert.Equal(t, value, err.value, "Error message should include caller info")
		assert.Equal(t, callerInfo, err.callerInfo, "Error message should include caller info")
		assert.Equal(t,
			fmt.Errorf("%w: %w", ErrNilPointerOrNilInterface, ErrTypeCast),
			err.underlyingErr,
			"Error message should include caller info",
		)
		assert.Equal(t,
			fmt.Errorf("%s: key %v: value err: %w: %w", err.callerInfo, err.key, ErrNilPointerOrNilInterface, ErrTypeCast).Error(),
			err.errStr,
			"Error message should include caller info",
		)
	})

	t.Run("other", func(t *testing.T) {
		const callerInfo = "TestCaller"
		value := "testValue"
		key := "testKey"

		err := NewErrTypeCastFailed(key, value, callerInfo)

		assert.Equal(t, key, err.key, "Error message should include caller info")
		assert.Equal(t, value, err.value, "Error message should include caller info")
		assert.Equal(t, callerInfo, err.callerInfo, "Error message should include caller info")
		assert.Equal(t,
			ErrTypeCast,
			err.underlyingErr,
			"Error message should include caller info",
		)
		assert.Equal(t,
			fmt.Errorf("%s: key %v: value %v: %w", err.callerInfo, err.key, err.value, ErrTypeCast).Error(),
			err.errStr,
			"Error message should include caller info",
		)
	})
}

func TestErrTypeCastFailed_Error(t *testing.T) {
	const callerInfo = "TestCaller"
	key := "testKey"
	value := "testValue"

	err := NewErrTypeCastFailed(key, value, callerInfo)
	errMessage := err.Error()

	assert.Contains(t, errMessage, callerInfo, "Error message should include caller info")
	assert.Contains(t, errMessage, key, "Error message should include key")
	assert.Contains(t, errMessage, value, "Error message should include value")
}

func TestErrTypeCastFailed_Is(t *testing.T) {
	const callerInfo = "TestCaller"
	key := "testKey"
	value := "testValue"

	err := NewErrTypeCastFailed(key, value, callerInfo)
	target := &ErrTypeCastFailed{}

	assert.True(t, err.Is(target), "Is should return true for ErrTypeCastFailed type")
}

func TestNewErrInvalidValue(t *testing.T) {
	const callerInfo = "TestCaller"
	value := "invalidValue"
	signalError := ErrNil

	err := NewErrInvalidValue(value, signalError, callerInfo)

	require.NotNil(t, err, "NewErrInvalidValue should return a non-nil error")
	assert.Equal(t, callerInfo, err.GetCallerInfo(), "CallerInfo should match the input")
	assert.Equal(t, value, err.GetInvalidValue(), "InvalidValue should match the input")
	assert.Equal(t, signalError, err.Unwrap(), "SignalError should match the input")
	assert.ErrorIs(t, err, signalError, "Error should wrap the signal error")
}

func TestErrInvalidValue_GetInvalidValueType(t *testing.T) {
	const callerInfo = "TestCaller"
	value := "invalidValue"
	signalError := ErrNil

	err := NewErrInvalidValue(value, signalError, callerInfo)

	require.Error(t, err, "it is error")
	assert.ErrorIs(t, err, NewErrInvalidValue("", signalError, ""), "error should be of type ErrInvalidValue")
	assert.Equal(t, reflect.TypeOf(value), err.GetInvalidValueType(), "Error message should include invalid value")
}

func TestErrInvalidValue_Error(t *testing.T) {
	const callerInfo = "TestCaller"
	value := "invalidValue"
	signalError := ErrNil

	err := NewErrInvalidValue(value, signalError, callerInfo)
	errMessage := err.Error()

	assert.Contains(t, errMessage, callerInfo, "Error message should include caller info")
	assert.Contains(t, errMessage, value, "Error message should include invalid value")
}

func TestErrInvalidValue_Is(t *testing.T) {
	const callerInfo = "TestCaller"
	value := "invalidValue"
	signalError := ErrNil

	err := NewErrInvalidValue(value, signalError, callerInfo)
	target := &ErrInvalidValue{signalError: ErrNil}

	assert.True(t, err.Is(target), "Is should return true for ErrInvalidValue type and matching signalError")
}

func TestNewErrKeyNotFound(t *testing.T) {
	key := "missingKey"

	err := NewErrKeyNotFound(key)

	require.NotNil(t, err, "NewErrKeyNotFound should return a non-nil error")
	assert.Equal(t, key, err.GetKey(), "Key should match the input")
	assert.ErrorIs(t, err, ErrNotFound, "Error should wrap ErrNotFound")
}

func TestErrKeyNotFound_Error(t *testing.T) {
	key := "missingKey"

	err := NewErrKeyNotFound(key)
	errMessage := err.Error()

	assert.Contains(t, errMessage, key, "Error message should include the missing key")
	assert.Contains(t, errMessage, "not found", "Error message should include 'not found'")
}

func TestErrKeyNotFound_Is(t *testing.T) {
	key := "missingKey"

	err := NewErrKeyNotFound(key)
	target := &ErrKeyNotFound{}

	assert.True(t, err.Is(target), "Is should return true for ErrKeyNotFound type")
}

func TestNewErrNilOrErrCtx(t *testing.T) {
	const callerInfo = "TestCaller"

	t.Run("Nil Context", func(t *testing.T) {
		err := NewErrNilOrErrCtx(callerInfo, nil)

		require.NotNil(t, err, "NewErrNilOrErrCtx should return a non-nil error for nil context")
		assert.Equal(t, ErrNilCtx, err.Unwrap(), "Error should wrap ErrNilCtx for nil context")
		assert.Equal(t, callerInfo, err.GetCallerInfo(), "CallerInfo should match the input")
	})

	t.Run("Canceled Context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := NewErrNilOrErrCtx(callerInfo, ctx)

		require.NotNil(t, err, "NewErrNilOrErrCtx should return a non-nil error for canceled context")
		assert.ErrorIs(t, err, context.Canceled, "Error should wrap context.Canceled for canceled context")
		assert.Equal(t, callerInfo, err.GetCallerInfo(), "CallerInfo should match the input")
	})
}

func TestErrCtx_Error(t *testing.T) {
	const callerInfo = "TestCaller"
	ctx := context.Background()

	err := NewErrNilOrErrCtx(callerInfo, ctx)
	errMessage := err.Error()

	assert.Contains(t, errMessage, callerInfo, "Error message should include caller info")
}

func TestErrCtx_Is(t *testing.T) {
	const callerInfo = "TestCaller"
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewErrNilOrErrCtx(callerInfo, ctx)
	target := &ErrCtx{ctxErr: ErrNilCtx}

	assert.False(t, err.Is(target), "Is should return false for ErrCtx type with different ctxErr")
}
