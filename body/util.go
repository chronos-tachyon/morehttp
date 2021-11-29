package body

import (
	"errors"
	"io"
	"io/fs"
)

func isEOF(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	return errors.Is(err, io.EOF)
}

func isErrClosed(err error) bool {
	if err == nil {
		return false
	}
	if err == fs.ErrClosed {
		return true
	}
	return errors.Is(err, fs.ErrClosed)
}
