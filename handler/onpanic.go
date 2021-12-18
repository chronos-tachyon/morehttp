package handler

import (
	"fmt"
)

type PanicError struct {
	Value interface{}
}

func (err PanicError) Error() string {
	return fmt.Sprintf("panic called with value of type %T: %+v", err.Value, err.Value)
}

var _ error = PanicError{}

func DefaultOnPanic(err error) {}

var OnPanic func(error) = DefaultOnPanic
