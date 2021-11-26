package body

import (
	"fmt"
	"io"
	"io/fs"
	"sync"
	"sync/atomic"

	"github.com/chronos-tachyon/assert"
)

type seekerCommon struct {
	mu     sync.Mutex
	s      io.ReadSeeker
	length int64
	refcnt int32
}

func (common *seekerCommon) ref() {
	atomic.AddInt32(&common.refcnt, 1)
}

func (common *seekerCommon) unref() error {
	refcnt := atomic.AddInt32(&common.refcnt, -1)
	if refcnt > 0 {
		return nil
	}
	if c, ok := common.s.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (common *seekerCommon) readAt(p []byte, offset int64) (int, error) {
	common.mu.Lock()
	defer common.mu.Unlock()

	_, err := common.s.Seek(offset, io.SeekStart)
	if err != nil {
		return 0, err
	}

	x := len(p)
	n, err := common.s.Read(p)
	assert.Assertf(n >= 0, "Read must return %d >= 0", n)
	assert.Assertf(n <= x, "Read must return %d <= %d", n, x)
	return n, err
}

type seekerBody struct {
	mu     sync.Mutex
	common *seekerCommon
	offset int64
	closed bool
}

func (body *seekerBody) Length() int64 {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0
	}

	return (body.common.length - body.offset)
}

func (body *seekerBody) Read(p []byte) (int, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	x := int64(len(p))
	eof := false
	avail := body.common.length - body.offset
	if x > avail {
		x = avail
		eof = true
	}

	n, err := body.common.readAt(p[0:x], body.offset)
	body.offset += int64(n)
	if eof && err == nil {
		err = io.EOF
	}
	return n, err
}

func (body *seekerBody) Close() error {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return fs.ErrClosed
	}

	err := body.common.unref()
	body.common = nil
	body.offset = 0
	body.closed = true
	return err
}

func (body *seekerBody) Seek(offset int64, whence int) (int64, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	bodyLen := body.common.length

	switch whence {
	case io.SeekStart:
		if offset < 0 {
			return 0, fmt.Errorf("Seek error: whence is SeekStart but offset %d is negative", offset)
		}

	case io.SeekCurrent:
		offset += body.offset

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

	body.offset = offset
	return offset, nil
}

func (body *seekerBody) ReadAt(p []byte, offset int64) (int, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	if offset < 0 {
		return 0, fmt.Errorf("ReadAt error: offset %d is negative", offset)
	}

	bodyLen := body.common.length

	if offset > bodyLen {
		offset = bodyLen
	}

	x := int64(len(p))
	eof := false
	avail := bodyLen - offset
	if x > avail {
		x = avail
		eof = true
	}

	n, err := body.common.readAt(p[0:x], offset)
	if eof && err == nil {
		err = io.EOF
	}
	return n, err
}

func (body *seekerBody) Copy() (Body, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return closedSingleton, nil
	}

	dupe := &seekerBody{
		common: body.common,
		offset: body.offset,
		closed: body.closed,
	}
	dupe.common.ref()
	return dupe, nil
}

func (body *seekerBody) Unwrap() io.Reader {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return nil
	}

	return body.common.s
}

var (
	_ Body        = (*seekerBody)(nil)
	_ io.Seeker   = (*seekerBody)(nil)
	_ io.ReaderAt = (*seekerBody)(nil)
)
