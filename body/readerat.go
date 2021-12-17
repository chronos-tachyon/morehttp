package body

import (
	"fmt"
	"io"
	"io/fs"
	"sync"

	"github.com/chronos-tachyon/assert"
)

type readerAtCommon struct {
	mu     sync.Mutex
	r      io.Reader
	at     io.ReaderAt
	length int64
	refcnt int32
}

func (common *readerAtCommon) ref() {
	common.mu.Lock()
	common.refcnt++
	common.mu.Unlock()
}

func (common *readerAtCommon) unref() error {
	common.mu.Lock()
	defer common.mu.Unlock()

	common.refcnt--

	if common.refcnt > 0 {
		return nil
	}

	var err error
	if c, cOK := common.r.(io.Closer); cOK {
		err = c.Close()
	}

	common.r = nil
	common.at = nil
	common.length = 0
	common.refcnt = 0

	return err
}

func (common *readerAtCommon) Length() int64 {
	common.mu.Lock()
	length := common.length
	common.mu.Unlock()
	return length
}

func (common *readerAtCommon) readAt(p []byte, offset int64) (int, error) {
	x := len(p)
	n, err := common.at.ReadAt(p, offset)
	assert.Assertf(n >= 0, "ReadAt must return %d >= 0", n)
	assert.Assertf(n <= x, "ReadAt must return %d <= %d", n, x)
	return n, err
}

type readerAtBody struct {
	mu     sync.Mutex
	common *readerAtCommon
	err    error
	offset int64
	closed bool
}

func (body *readerAtBody) Length() int64 {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0
	}

	length := body.common.Length()
	offset := body.offset
	if offset >= length {
		return 0
	}
	return (length - offset)
}

func (body *readerAtBody) Read(p []byte) (int, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	if body.err != nil {
		return 0, body.err
	}

	common := body.common
	common.mu.Lock()
	defer common.mu.Unlock()

	offset := body.offset
	length := common.length
	if offset > length {
		offset = length
	}

	avail := (length - offset)
	x := int64(len(p))
	eof := false
	if x > avail {
		x = avail
		eof = true
	}

	var n int
	var err error

	if avail > 0 {
		n, err = common.readAt(p[0:x], offset)
	}
	if eof && err == nil {
		err = io.EOF
	}

	offset += int64(n)
	body.offset = offset

	if err != nil {
		body.err = err

		length = offset
		if length < common.length {
			common.length = length
		}
	}

	return n, err
}

func (body *readerAtBody) Close() error {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return fs.ErrClosed
	}

	common := body.common
	body.common = nil
	body.err = nil
	body.offset = 0
	body.closed = true
	return common.unref()
}

func (body *readerAtBody) Seek(offset int64, whence int) (int64, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return -1, fs.ErrClosed
	}

	common := body.common
	length := common.Length()

	switch whence {
	case io.SeekStart:
		if offset < 0 {
			return -1, NegativeStartOffsetSeekError{offset}
		}
	case io.SeekCurrent:
		offset += body.offset
	case io.SeekEnd:
		offset += length
	default:
		return -1, UnknownWhenceSeekError{whence}
	}

	if offset < 0 {
		return -1, NegativeComputedOffsetSeekError{offset}
	}

	if offset > length {
		offset = length
	}

	body.err = nil
	body.offset = offset
	return offset, nil
}

func (body *readerAtBody) ReadAt(p []byte, offset int64) (int, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	if offset < 0 {
		return 0, fmt.Errorf("ReadAt error: offset %d is negative", offset)
	}

	common := body.common
	common.mu.Lock()
	defer common.mu.Unlock()

	length := common.length
	if offset > length {
		offset = length
	}

	avail := (length - offset)
	x := int64(len(p))
	eof := false
	if x > avail {
		x = avail
		eof = true
	}

	var n int
	var err error
	if avail > 0 {
		n, err = common.readAt(p[0:x], offset)
	}
	if eof && err == nil {
		err = io.EOF
	}

	return n, err
}

func (body *readerAtBody) Copy() (Body, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return closedSingleton, nil
	}

	dupe := &readerAtBody{
		common: body.common,
		err:    body.err,
		offset: body.offset,
		closed: body.closed,
	}
	dupe.common.ref()
	return dupe, nil
}

func (body *readerAtBody) Unwrap() io.Reader {
	var r io.Reader
	body.mu.Lock()
	if !body.closed {
		common := body.common
		common.mu.Lock()
		r = common.r
		common.mu.Unlock()
	}
	body.mu.Unlock()
	return r
}

var (
	_ Body        = (*readerAtBody)(nil)
	_ io.Seeker   = (*readerAtBody)(nil)
	_ io.ReaderAt = (*readerAtBody)(nil)
)
