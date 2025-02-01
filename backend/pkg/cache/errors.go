package cache

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

var ErrNilFunc = errors.New("nil function")
var ErrNilPointerOrNilInterface = errors.New("nil pointer or nil interface")
var ErrNil = errors.New("nil value")
var ErrEmptyString = errors.New("empty string")
var ErrCacheClosed = errors.New("cache closed")
var ErrNotFound = errors.New("not found")
var ErrNilCtx = errors.New("nil context")
var ErrTypeCast = errors.New("internal type cast error")

type ErrTypeCastFailed struct {
	key           any
	value         any
	callerInfo    string
	underlyingErr error
	errStr        string
}

func NewErrTypeCastFailed(key, value any, callerInfo string) *ErrTypeCastFailed {
	e := &ErrTypeCastFailed{key: key, value: value, callerInfo: callerInfo}

	e.error()

	return e
}

func (e *ErrTypeCastFailed) Unwrap() error {
	return e.underlyingErr
}
func (e *ErrTypeCastFailed) Error() string {
	return e.errStr
}

func (e *ErrTypeCastFailed) error() {
	if e.key == nil {
		e.errStr = fmt.Errorf("%s: key err: %w: %w", e.callerInfo, ErrNil, ErrTypeCast).Error()
		e.underlyingErr = fmt.Errorf("%w: %w", ErrNil, ErrTypeCast)
		return
	}
	if val := reflect.ValueOf(e.key); val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			e.errStr = fmt.Errorf("%s: key err: %w: %w", e.callerInfo, ErrNilPointerOrNilInterface, ErrTypeCast).Error()
			e.underlyingErr = fmt.Errorf("%w: %w", ErrNilPointerOrNilInterface, ErrTypeCast)
			return
		}
	}

	if e.value == nil {
		e.errStr = fmt.Errorf("%s: key %v: value err: %w: %w", e.callerInfo, e.key, ErrNil, ErrTypeCast).Error()
		e.underlyingErr = fmt.Errorf("%w: %w", ErrNil, ErrTypeCast)
		return
	}
	if val := reflect.ValueOf(e.value); val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			e.errStr = fmt.Errorf("%s: key %v: value err: %w: %w", e.callerInfo, e.key, ErrNilPointerOrNilInterface, ErrTypeCast).Error()
			e.underlyingErr = fmt.Errorf("%w: %w", ErrNilPointerOrNilInterface, ErrTypeCast)
			return
		}
	}

	e.errStr = fmt.Errorf("%s: key %v: value %v: %w", e.callerInfo, e.key, e.value, ErrTypeCast).Error()
	e.underlyingErr = ErrTypeCast
}

// Is function only checks for an ErrTypeCastFailed type and don't compare for an underlying values
func (e *ErrTypeCastFailed) Is(target error) bool {
	_, ok := target.(*ErrTypeCastFailed)
	return ok
}

type ErrInvalidValue struct {
	invalidValue any
	signalError  error
	callerInfo   string
}

func NewErrInvalidValue(invalidValue any, signalError error, callerInfo string) *ErrInvalidValue {
	return &ErrInvalidValue{invalidValue: invalidValue, signalError: signalError, callerInfo: callerInfo}
}

func (e *ErrInvalidValue) GetInvalidValue() any {
	return e.invalidValue
}

func (e *ErrInvalidValue) GetInvalidValueType() reflect.Type {
	return reflect.TypeOf(e.invalidValue)
}

func (e *ErrInvalidValue) GetCallerInfo() string {
	return e.callerInfo
}

// Is function only checks for an ErrInvalidValue type and signalError equality.
// It doesn't compare underlying keys
func (e *ErrInvalidValue) Is(target error) bool {
	err, ok := target.(*ErrInvalidValue)
	if !ok {
		return false
	}
	return e.signalError.Error() == err.signalError.Error()
}

func (e *ErrInvalidValue) Unwrap() error {
	return e.signalError
}

func (e *ErrInvalidValue) Error() string {
	return fmt.Errorf("%s: value [%s]: %w", e.callerInfo, e.invalidValue, e.signalError).Error()
}

type ErrKeyNotFound struct {
	key string
	err error
}

func NewErrKeyNotFound(key string) *ErrKeyNotFound {
	return &ErrKeyNotFound{key: key, err: ErrNotFound}
}

func (e *ErrKeyNotFound) Error() string {
	return fmt.Errorf("key %s: %w", e.key, e.err).Error()
}

// Is function only checks for an ErrKeyNotFound type and don't compare for an underlying key
func (e *ErrKeyNotFound) Is(target error) bool {
	_, ok := target.(*ErrKeyNotFound)
	return ok
}

func (e *ErrKeyNotFound) Unwrap() error {
	return e.err
}

func (e *ErrKeyNotFound) GetKey() string {
	return e.key
}

type ErrCtx struct {
	callerInfo string
	ctxErr     error
}

func NewErrNilOrErrCtx(callerInfo string, ctx context.Context) *ErrCtx {
	if ctx == nil {
		return &ErrCtx{callerInfo: callerInfo, ctxErr: ErrNilCtx}
	}
	return &ErrCtx{callerInfo: callerInfo, ctxErr: ctx.Err()}
}

// Is function only checks for an ErrKeyNotFound type and underlying ctxErr match
func (e *ErrCtx) Is(target error) bool {
	err, ok := target.(*ErrCtx)
	if !ok {
		return false
	}
	return e.ctxErr.Error() == err.ctxErr.Error()
}

func (e *ErrCtx) Error() string {
	return fmt.Errorf("%s: %w", e.callerInfo, e.ctxErr).Error()
}

func (e *ErrCtx) Unwrap() error {
	return e.ctxErr
}

func (e *ErrCtx) GetCallerInfo() string { return e.callerInfo }
