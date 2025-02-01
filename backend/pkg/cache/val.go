package cache

import (
	"context"
	"fmt"
	"reflect"
	"sync/atomic"

	cache2 "github.com/KennyMacCormik/otel/backend/pkg/models/errors/cache"
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
			return fmt.Errorf("%s: %w", wrap, cache2.ErrCacheClosed)
		}
		return nil
	}
}

func WithCtxValidation(ctx context.Context, wrap string) func() error {
	return func() error {
		if ctx == nil || ctx.Err() != nil {
			return cache2.NewErrNilOrErrCtx(wrap, ctx)
		}
		return nil
	}
}

func IsKeyValid(key string, callerInfo string) error {
	if key == "" {
		return cache2.NewErrInvalidValue(key, cache2.ErrEmptyString, callerInfo)
	}
	return nil
}

// IsNotNil checks if a value is nil or nil pointer.
func IsNotNil(value any, callerInfo string) error {
	// The interface itself is nil
	if value == nil {
		return cache2.NewErrInvalidValue(value, cache2.ErrNil, callerInfo)
	}

	// Check if it is a nil pointer, nil interface, or empty function
	val := reflect.ValueOf(value)
	kind := val.Kind()

	// TODO fix warning
	switch kind {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return cache2.NewErrInvalidValue(value, cache2.ErrNilPointerOrNilInterface, callerInfo)
		}
	case reflect.Func:
		if val.IsNil() {
			return cache2.NewErrInvalidValue(value, cache2.ErrNilFunc, callerInfo)
		}
	}

	return nil
}
