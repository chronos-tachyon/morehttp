package body

import (
	"io"
	"io/fs"
)

var closedSingleton Body = (*closedBody)(nil)

type closedBody struct{}

func (body *closedBody) Length() int64 {
	return 0
}

func (body *closedBody) Read(p []byte) (int, error) {
	return 0, fs.ErrClosed
}

func (body *closedBody) Close() error {
	return fs.ErrClosed
}

func (body *closedBody) Seek(offset int64, whence int) (int64, error) {
	return 0, fs.ErrClosed
}

func (body *closedBody) ReadAt(p []byte, offset int64) (int, error) {
	return 0, fs.ErrClosed
}

func (body *closedBody) WriteTo(w io.Writer) (int64, error) {
	return 0, fs.ErrClosed
}

func (body *closedBody) Copy() (Body, error) {
	return closedSingleton, nil
}

func (body *closedBody) Unwrap() io.Reader {
	return nil
}

var (
	_ Body        = (*closedBody)(nil)
	_ io.Seeker   = (*closedBody)(nil)
	_ io.ReaderAt = (*closedBody)(nil)
	_ io.WriterTo = (*closedBody)(nil)
)
