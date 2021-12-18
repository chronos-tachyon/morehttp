package body

import (
	"io"
	"io/fs"
	"sync"
)

type emptyBody struct {
	mu     sync.Mutex
	eof    bool
	closed bool
}

func (body *emptyBody) BytesRemaining() int64 {
	return 0
}

func (body *emptyBody) Read(p []byte) (int, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	if body.eof || len(p) > 0 {
		body.eof = true
		return 0, io.EOF
	}

	return 0, nil
}

func (body *emptyBody) Close() error {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return fs.ErrClosed
	}

	body.closed = true
	return nil
}

func (body *emptyBody) Seek(offset int64, whence int) (int64, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return -1, fs.ErrClosed
	}

	switch whence {
	case io.SeekStart:
		if offset < 0 {
			return -1, NegativeStartOffsetSeekError{offset}
		}
	case io.SeekCurrent:
		// pass
	case io.SeekEnd:
		// pass
	default:
		return -1, UnknownWhenceSeekError{whence}
	}

	if offset < 0 {
		return -1, NegativeComputedOffsetSeekError{offset}
	}

	body.eof = false
	return 0, nil
}

func (body *emptyBody) ReadAt(p []byte, offset int64) (int, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	if len(p) > 0 {
		return 0, io.EOF
	}

	return 0, nil
}

func (body *emptyBody) WriteTo(w io.Writer) (int64, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	return 0, io.EOF
}

func (body *emptyBody) Copy() (Body, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return closedSingleton, nil
	}

	dupe := &emptyBody{eof: body.eof}
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
