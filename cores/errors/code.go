package errors

import (
	"errors"
	"fmt"
)

const (
	unknownCode = 500
)

type Error struct {
	code    int
	message string
}

func (e *Error) Error() string { return fmt.Sprintf("%d - %s", e.code, e.message) }

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Code32() int32 {
	return int32(e.code)
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.code == e.code
	}
	return false
}

// withMessage 信息组合error
type withMessage struct {
	cause error
	msg   string
}

func (w *withMessage) Error() string { return w.msg + "| " + w.cause.Error() }

func (w *withMessage) Cause() error { return w.cause }

func (w *withMessage) Unwrap() error { return w.cause }

// New 创建错误
func New(code int, message string) *Error {
	return &Error{
		code:    code,
		message: message,
	}
}

// WithMessage 追加错误信息
func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause: err,
		msg:   message,
	}
}

func Cause(err error) *Error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	if _, ok := err.(*Error); ok {
		return err.(*Error)
	}
	return New(unknownCode, err.Error())
}
