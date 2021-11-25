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
	closed bool
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
		closed: body.closed,
	}
	return dupe, nil
}

func (body *bytesBody) Read(p []byte) (int, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	avail := len(body.data) - body.offset

	var err error
	n := len(p)
	if n > avail {
		n = avail
		err = io.EOF
	}

	i := body.offset
	j := i + n
	copy(p[0:n], body.data[i:j])
	body.offset = j
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

	avail := bodyLen - offset

	var err error
	n := len(p)
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

	i := body.offset
	j := len(body.data)
	avail := (j - i)
	n, err := w.Write(body.data[i:j])
	assert.Assertf(n >= 0, "Write must return %d >= 0", n)
	assert.Assertf(n <= avail, "Write must return %d <= %d", n, avail)
	body.offset += int(n)
	return int64(n), err
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
