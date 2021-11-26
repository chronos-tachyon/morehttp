package body

import (
	"fmt"
	"io"
	"io/fs"
	"sync/atomic"
)

type emptyBody struct {
	closed uint32
}

func (body *emptyBody) isClosed() bool {
	return atomic.LoadUint32(&body.closed) != 0
}

func (body *emptyBody) Length() int64 {
	return 0
}

func (body *emptyBody) Read(p []byte) (int, error) {
	if body.isClosed() {
		return 0, fs.ErrClosed
	}

	if len(p) > 0 {
		return 0, io.EOF
	}

	return 0, nil
}

func (body *emptyBody) Close() error {
	closed := atomic.AddUint32(&body.closed, 1)
	if closed > 1 {
		return fs.ErrClosed
	}
	return nil
}

func (body *emptyBody) Seek(offset int64, whence int) (int64, error) {
	if body.isClosed() {
		return 0, fs.ErrClosed
	}

	switch whence {
	case io.SeekStart:
		if offset < 0 {
			return 0, fmt.Errorf("Seek error: whence is SeekStart but offset %d is negative", offset)
		}

	case io.SeekCurrent:
		// pass

	case io.SeekEnd:
		// pass

	default:
		return 0, fmt.Errorf("Seek error: unknown whence value %d", whence)
	}

	if offset < 0 {
		return 0, fmt.Errorf("Seek error: computed offset %d is negative", offset)
	}

	return 0, nil
}

func (body *emptyBody) ReadAt(p []byte, offset int64) (int, error) {
	if body.isClosed() {
		return 0, fs.ErrClosed
	}

	if len(p) > 0 {
		return 0, io.EOF
	}

	return 0, fs.ErrClosed
}

func (body *emptyBody) WriteTo(w io.Writer) (int64, error) {
	if body.isClosed() {
		return 0, fs.ErrClosed
	}

	return 0, io.EOF
}

func (body *emptyBody) Copy() (Body, error) {
	if body.isClosed() {
		return closedSingleton, nil
	}

	dupe := &emptyBody{}
	return dupe, nil
}

func (body *emptyBody) Unwrap() io.Reader {
	return nil
}

var (
	_ Body        = (*emptyBody)(nil)
	_ io.Seeker   = (*emptyBody)(nil)
	_ io.ReaderAt = (*emptyBody)(nil)
	_ io.WriterTo = (*emptyBody)(nil)
)
