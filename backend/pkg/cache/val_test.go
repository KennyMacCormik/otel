package cache

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
)

func TestIsNotNil_NilInterface(t *testing.T) {
	err := IsNotNil(nil, "TestIsNotNil_NilInterface")
	require.Error(t, err, "IsNotNil should return an error for a nil interface")
	assert.ErrorIs(t, err, ErrNil, "Error should indicate a nil value")
}

func TestIsNotNil_NilPointer(t *testing.T) {
	var ptr *int
	err := IsNotNil(ptr, "TestIsNotNil_NilPointer")
	require.Error(t, err, "IsNotNil should return an error for a nil pointer")
	assert.ErrorIs(t, err, ErrNilPointerOrNilInterface, "Error should indicate a nil pointer")
}

func TestIsNotNil_ValidPointer(t *testing.T) {
	ptr := new(int)
	err := IsNotNil(ptr, "TestIsNotNil_ValidPointer")
	require.NoError(t, err, "IsNotNil should not return an error for a valid pointer")
}

func TestIsNotNil_NilFunc(t *testing.T) {
	var fn func()
	err := IsNotNil(fn, "TestIsNotNil_NilFunc")
	require.Error(t, err, "IsNotNil should return an error for a nil function")
	assert.ErrorIs(t, err, ErrNilFunc, "Error should indicate a nil function")
}

func TestIsNotNil_ValidFunc(t *testing.T) {
	fn := func() {}
	err := IsNotNil(fn, "TestIsNotNil_ValidFunc")
	require.NoError(t, err, "IsNotNil should not return an error for a valid function")
}

func TestIsKeyValid_EmptyString(t *testing.T) {
	err := IsKeyValid("", "TestIsKeyValid_EmptyString")
	require.Error(t, err, "IsKeyValid should return an error for an empty string")
	assert.ErrorIs(t, err, ErrEmptyString, "Error should indicate an empty string")
}

func TestIsKeyValid_ValidKey(t *testing.T) {
	err := IsKeyValid("validKey", "TestIsKeyValid_ValidKey")
	require.NoError(t, err, "IsKeyValid should not return an error for a valid key")
}

func TestWithValueValidation_NilValue(t *testing.T) {
	validateFunc := WithValueValidation(nil, "TestWithValueValidation_NilValue")
	err := validateFunc()
	require.Error(t, err, "WithValueValidation should return an error for a nil value")
	assert.ErrorIs(t, err, ErrNil, "Error should indicate a nil value")
}

func TestWithValueValidation_ValidValue(t *testing.T) {
	validateFunc := WithValueValidation("validValue", "TestWithValueValidation_ValidValue")
	err := validateFunc()
	require.NoError(t, err, "WithValueValidation should not return an error for a valid value")
}

func TestWithCtxValidation_NilContext(t *testing.T) {
	validateFunc := WithCtxValidation(nil, "TestWithCtxValidation_NilContext")
	err := validateFunc()
	require.Error(t, err, "WithCtxValidation should return an error for a nil context")
	assert.ErrorIs(t, err, ErrNilCtx, "Error should indicate a nil context")
}

func TestWithCtxValidation_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	validateFunc := WithCtxValidation(ctx, "TestWithCtxValidation_CanceledContext")
	err := validateFunc()
	require.Error(t, err, "WithCtxValidation should return an error for a canceled context")
	assert.ErrorIs(t, err, ctx.Err(), "Error should indicate the context is canceled")
}

func TestWithCtxValidation_ValidContext(t *testing.T) {
	ctx := context.Background()
	validateFunc := WithCtxValidation(ctx, "TestWithCtxValidation_ValidContext")
	err := validateFunc()
	require.NoError(t, err, "WithCtxValidation should not return an error for a valid context")
}

func TestWithKeyValidation_EmptyString(t *testing.T) {
	validateFunc := WithKeyValidation("", "TestWithKeyValidation_EmptyString")
	err := validateFunc()
	require.Error(t, err, "WithKeyValidation should return an error for an empty string")
	assert.ErrorIs(t, err, ErrEmptyString, "Error should indicate an empty string")
}

func TestWithKeyValidation_ValidKey(t *testing.T) {
	validateFunc := WithKeyValidation("validKey", "TestWithKeyValidation_ValidKey")
	err := validateFunc()
	require.NoError(t, err, "WithKeyValidation should not return an error for a valid key")
}

func TestWithClosedValidation_ClosedCache(t *testing.T) {
	closed := atomic.Bool{}
	closed.Store(true)

	validateFunc := WithClosedValidation(&closed, "TestWithClosedValidation_ClosedCache")
	err := validateFunc()
	require.Error(t, err, "WithClosedValidation should return an error if the cache is closed")
	assert.ErrorIs(t, err, ErrCacheClosed, "Error should indicate the cache is closed")
}

func TestWithClosedValidation_OpenCache(t *testing.T) {
	closed := atomic.Bool{}
	closed.Store(false)

	validateFunc := WithClosedValidation(&closed, "TestWithClosedValidation_OpenCache")
	err := validateFunc()
	require.NoError(t, err, "WithClosedValidation should not return an error if the cache is open")
}

func TestValidateInput_NoErrors(t *testing.T) {
	validateFuncs := []ValidateFunc{
		func() error { return nil },
		func() error { return nil },
	}

	err := ValidateInput(validateFuncs...)
	require.NoError(t, err, "ValidateInput should not return an error if all validation functions pass")
}

func TestValidateInput_OneError(t *testing.T) {
	validateFuncs := []ValidateFunc{
		func() error { return nil },
		func() error { return fmt.Errorf("validation error") },
	}

	err := ValidateInput(validateFuncs...)
	require.Error(t, err, "ValidateInput should return an error if one of the validation functions fails")
	assert.EqualError(t, err, "validation error", "Error message should match the failing validation function")
}

func TestValidateInput_MultipleErrors(t *testing.T) {
	validateFuncs := []ValidateFunc{
		func() error { return fmt.Errorf("error 1") },
		func() error { return fmt.Errorf("error 2") },
	}

	err := ValidateInput(validateFuncs...)
	require.Error(t, err, "ValidateInput should return the first error if multiple validation functions fail")
	assert.EqualError(t, err, "error 1", "Error message should match the first failing validation function")
}
