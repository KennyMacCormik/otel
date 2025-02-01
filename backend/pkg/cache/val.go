package cache

import (
	"context"
	"fmt"
	"reflect"
	"sync/atomic"
)

type ValidateFunc func() error

func ValidateInput(opts ...ValidateFunc) error {
	for _, opt := range opts {
		if err := opt(); err != nil {
			return err
		}
	}
	return nil
}

func WithValueValidation(value any, wrap string) func() error {
	return func() error {
		if err := IsNotNil(value, wrap); err != nil {
			return fmt.Errorf("%s: %w", wrap, err)
		}
		return nil
	}
}

func WithKeyValidation(key string, wrap string) func() error {
	return func() error {
		if err := IsKeyValid(key, wrap); err != nil {
			return fmt.Errorf("%s: %w", wrap, err)
		}
		return nil
	}
}

func WithClosedValidation(closed *atomic.Bool, wrap string) func() error {
	return func() error {
		if closed.Load() {
			return fmt.Errorf("%s: %w", wrap, ErrCacheClosed)
		}
		return nil
	}
}

func WithCtxValidation(ctx context.Context, wrap string) func() error {
	return func() error {
		if ctx == nil || ctx.Err() != nil {
			return NewErrNilOrErrCtx(wrap, ctx)
		}
		return nil
	}
}

func IsKeyValid(key string, callerInfo string) error {
	if key == "" {
		return NewErrInvalidValue(key, ErrEmptyString, callerInfo)
	}
	return nil
}

// IsNotNil checks if a value is nil or nil pointer.
func IsNotNil(value any, callerInfo string) error {
	// The interface itself is nil
	if value == nil {
		return NewErrInvalidValue(value, ErrNil, callerInfo)
	}

	// Check if it is a nil pointer, nil interface, or empty function
	val := reflect.ValueOf(value)
	kind := val.Kind()
	switch kind {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return NewErrInvalidValue(value, ErrNilPointerOrNilInterface, callerInfo)
		}
	case reflect.Func:
		if val.IsNil() {
			return NewErrInvalidValue(value, ErrNilFunc, callerInfo)
		}
	}
	return nil
}
