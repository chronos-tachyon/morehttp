package body

import (
	"fmt"
	"io"
	"io/fs"
	"sync"

	"github.com/chronos-tachyon/assert"
)

type bytesBody struct {
	mu     sync.Mutex
	data   []byte
	offset int
	eof    bool
	closed bool
}

func (body *bytesBody) Length() int64 {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0
	}

	return int64(len(body.data) - body.offset)
}

func (body *bytesBody) Read(p []byte) (int, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	if body.eof {
		return 0, io.EOF
	}

	n := len(p)
	avail := len(body.data) - body.offset
	eof := false
	var err error
	if n > avail {
		n = avail
		eof = true
		err = io.EOF
	}

	i := body.offset
	j := i + n
	copy(p[0:n], body.data[i:j])
	body.offset = j
	body.eof = eof
	return n, err
}

func (body *bytesBody) Close() error {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return fs.ErrClosed
	}

	body.data = nil
	body.offset = 0
	body.eof = true
	body.closed = true
	return nil
}

func (body *bytesBody) Seek(offset int64, whence int) (int64, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	bodyLen := int64(len(body.data))

	switch whence {
	case io.SeekStart:
		if offset < 0 {
			return 0, fmt.Errorf("Seek error: whence is SeekStart but offset %d is negative", offset)
		}

	case io.SeekCurrent:
		offset += int64(body.offset)

	case io.SeekEnd:
		offset += bodyLen

	default:
		return 0, fmt.Errorf("Seek error: unknown whence value %d", whence)
	}

	if offset < 0 {
		return 0, fmt.Errorf("Seek error: computed offset %d is negative", offset)
	}

	if offset > bodyLen {
		offset = bodyLen
	}

	body.offset = int(offset)
	body.eof = false
	return offset, nil
}

func (body *bytesBody) ReadAt(p []byte, offset int64) (int, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	if offset < 0 {
		return 0, fmt.Errorf("ReadAt error: offset %d is negative", offset)
	}

	bodyLen := int64(len(body.data))

	if offset > bodyLen {
		offset = bodyLen
	}

	var err error
	n := len(p)
	avail := bodyLen - offset
	if int64(n) > avail {
		n = int(avail)
		err = io.EOF
	}

	i := int(offset)
	j := i + n
	copy(p[0:n], body.data[i:j])
	return n, err
}

func (body *bytesBody) WriteTo(w io.Writer) (int64, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	if body.eof {
		return 0, nil
	}

	i := body.offset
	j := len(body.data)
	avail := (j - i)
	n, err := w.Write(body.data[i:j])
	assert.Assertf(n >= 0, "Write must return %d >= 0", n)
	assert.Assertf(n <= avail, "Write must return %d <= %d", n, avail)
	body.offset += int(n)
	body.eof = true
	return int64(n), err
}

func (body *bytesBody) Copy() (Body, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return closedSingleton, nil
	}

	dupe := &bytesBody{
		data:   body.data,
		offset: body.offset,
		eof:    body.eof,
		closed: body.closed,
	}
	return dupe, nil
}

func (body *bytesBody) Unwrap() io.Reader {
	return nil
}

var (
	_ Body        = (*bytesBody)(nil)
	_ io.Seeker   = (*bytesBody)(nil)
	_ io.ReaderAt = (*bytesBody)(nil)
	_ io.WriterTo = (*bytesBody)(nil)
)
