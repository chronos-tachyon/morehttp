package body

import (
	"fmt"
)

type UnknownWhenceSeekError struct {
	Whence int
}

func (err UnknownWhenceSeekError) GoString() string {
	return fmt.Sprintf("UnknownWhenceSeekError{%d}", err.Whence)
}

func (err UnknownWhenceSeekError) Error() string {
	return fmt.Sprintf("Seek error: unknown whence value %d", err.Whence)
}

var _ error = UnknownWhenceSeekError{}

type NegativeStartOffsetSeekError struct {
	Offset int64
}

func (err NegativeStartOffsetSeekError) GoString() string {
	return fmt.Sprintf("NegativeStartOffsetSeekError{%d}", err.Offset)
}

func (err NegativeStartOffsetSeekError) Error() string {
	return fmt.Sprintf("Seek error: whence is io.SeekStart but offset %d is negative", err.Offset)
}

var _ error = NegativeStartOffsetSeekError{}

type NegativeComputedOffsetSeekError struct {
	Offset int64
}

func (err NegativeComputedOffsetSeekError) GoString() string {
	return fmt.Sprintf("NegativeComputedOffsetSeekError{%d}", err.Offset)
}

func (err NegativeComputedOffsetSeekError) Error() string {
	return fmt.Sprintf("Seek error: computed offset %d is negative", err.Offset)
}

var _ error = NegativeComputedOffsetSeekError{}
