package body

import (
	"io"
	"io/fs"
	"sync"

	"github.com/chronos-tachyon/assert"
)

const blockSize = 65536

type bufferedCommon struct {
	mu      sync.Mutex
	r       io.Reader
	readErr error
	bodies  map[*bufferedBody]struct{}
	bytes   []byte
	length  int64
	start   int64
	refcnt  int32
}

func (common *bufferedCommon) ref(body *bufferedBody) {
	common.mu.Lock()
	defer common.mu.Unlock()

	common.bodies[body] = struct{}{}
	if body.offset == common.start {
		common.refcnt++
	}
}

func (common *bufferedCommon) unref(body *bufferedBody) error {
	common.mu.Lock()
	defer common.mu.Unlock()

	delete(common.bodies, body)

	if len(common.bodies) <= 0 {
		r := common.r
		common.r = nil
		common.readErr = nil
		common.bodies = nil
		common.bytes = nil
		common.length = 0
		common.start = 0
		if c, ok := r.(io.Closer); ok {
			return c.Close()
		}
		return nil
	}

	if body.offset == common.start {
		common.refcnt--
		if common.refcnt <= 0 {
			common.updateStart()
		}
	}
	return nil
}

func (common *bufferedCommon) advance(oldOffset, newOffset int64) {
	beginBuffer := common.start
	endBuffer := beginBuffer + int64(len(common.bytes))

	assert.Assertf(oldOffset >= beginBuffer, "%d >= %d", oldOffset, beginBuffer)
	assert.Assertf(newOffset >= oldOffset, "%d >= %d", newOffset, oldOffset)
	assert.Assertf(endBuffer >= newOffset, "%d >= %d", endBuffer, newOffset)

	if oldOffset == beginBuffer && newOffset != beginBuffer {
		common.refcnt--
		if common.refcnt <= 0 {
			common.updateStart()
		}
	}
}

func (common *bufferedCommon) updateStart() {
	var newStart int64
	var newRefcnt int32
	for body := range common.bodies {
		if newRefcnt <= 0 || body.offset < newStart {
			newStart = body.offset
			newRefcnt = 0
		}
		if body.offset == newStart {
			newRefcnt++
		}
	}

	assert.Assertf(newStart >= common.start, "new start %d >= old start %d", newStart, common.start)

	i := newStart - common.start
	common.bytes = common.bytes[i:]
	common.start = newStart
	common.refcnt = newRefcnt
}

type bufferedBody struct {
	mu     sync.Mutex
	common *bufferedCommon
	offset int64
	closed bool
}

func (body *bufferedBody) Length() int64 {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0
	}

	common := body.common
	common.mu.Lock()
	defer common.mu.Unlock()

	beginOffset := body.offset

	if common.length >= 0 && beginOffset >= common.length {
		return 0
	}

	if common.length >= 0 {
		return (common.length - beginOffset)
	}

	if common.readErr == nil {
		return -1
	}

	beginBuffer := common.start
	endBuffer := beginBuffer + int64(len(common.bytes))
	assert.Assertf(beginOffset >= beginBuffer, "%d >= %d", beginOffset, beginBuffer)

	if beginOffset >= endBuffer {
		return 0
	}

	return (endBuffer - beginOffset)
}

func (body *bufferedBody) Read(p []byte) (int, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0, fs.ErrClosed
	}

	common := body.common
	common.mu.Lock()
	defer common.mu.Unlock()

	beginBuffer := common.start
	endBuffer := beginBuffer + int64(len(common.bytes))
	beginOffset := body.offset
	endOffset := beginOffset + int64(len(p))

	assert.Assertf(beginOffset >= beginBuffer, "%d >= %d", beginOffset, beginBuffer)

	if common.readErr == nil && endOffset > endBuffer {
		numToRead := int64(len(p))
		numToRead = int64(uint64(numToRead+blockSize-1) &^ uint64(blockSize-1))

		eof := false
		if common.length >= 0 {
			assert.Assertf(endBuffer <= common.length, "%d <= %d", endBuffer, common.length)
			remain := (common.length - endBuffer)
			if numToRead >= remain {
				numToRead = remain
				eof = true
			}
		}

		bufLen := (endBuffer - beginBuffer)
		bufCap := bufLen
		bufCap += numToRead
		bufNew := make([]byte, bufCap)
		copy(bufNew[0:bufLen], common.bytes[0:bufLen])
		common.bytes = bufNew[0:bufLen]

		for numToRead > 0 {
			i := bufLen
			x := numToRead
			j := i + x

			n, err := common.r.Read(bufNew[i:j])

			n64 := int64(n)
			assert.Assertf(n64 >= 0, "Read must return %d >= 0", n64)
			assert.Assertf(n64 <= x, "Read must return %d <= %d", n64, x)

			bufLen += n64
			numToRead -= n64

			common.bytes = bufNew[0:bufLen]
			endBuffer = beginBuffer + bufLen

			if err != nil {
				common.readErr = err
				break
			}
		}

		if eof && common.readErr == nil {
			common.readErr = io.EOF
		}
	}

	var err error
	if endOffset > endBuffer {
		endOffset = endBuffer
		err = common.readErr
	}

	i := int(beginOffset - beginBuffer)
	j := int(endOffset - beginBuffer)
	n := (j - i)
	copy(p[0:n], common.bytes[i:j])
	body.offset = endOffset
	common.advance(beginOffset, endOffset)
	return n, err
}

func (body *bufferedBody) Close() error {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return fs.ErrClosed
	}

	err := body.common.unref(body)
	body.common = nil
	body.offset = 0
	body.closed = true
	return err
}

func (body *bufferedBody) Copy() (Body, error) {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return closedSingleton, nil
	}

	dupe := &bufferedBody{
		common: body.common,
		offset: body.offset,
		closed: body.closed,
	}
	dupe.common.ref(dupe)
	return dupe, nil
}

func (body *bufferedBody) Unwrap() io.Reader {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return nil
	}

	common := body.common
	common.mu.Lock()
	r := common.r
	common.mu.Unlock()

	return r
}

var (
	_ Body = (*bufferedBody)(nil)
)
