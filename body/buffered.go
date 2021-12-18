package body

import (
	"io"
	"io/fs"
	"sync"

	"github.com/chronos-tachyon/assert"
)

const blockSize = 65536

type bufferedCommon struct {
	mu     sync.Mutex
	r      io.Reader
	err    error
	bodies map[*bufferedBody]struct{}
	bytes  []byte
	length int64
	start  int64
	refcnt int32
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

	if len(common.bodies) > 0 {
		if body.offset == common.start {
			common.refcnt--
			if common.refcnt <= 0 {
				common.updateStart()
			}
		}
		return nil
	}

	r := common.r

	common.r = nil
	common.err = nil
	common.bodies = nil
	common.bytes = nil
	common.length = 0
	common.start = 0

	var err error
	if c, cOK := r.(io.Closer); cOK {
		err = c.Close()
	}
	return err
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
	var beginBufferNew int64
	var refcnt int32
	for body := range common.bodies {
		if refcnt <= 0 || body.offset < beginBufferNew {
			beginBufferNew = body.offset
			refcnt = 0
		}
		if body.offset == beginBufferNew {
			refcnt++
		}
	}

	beginBufferOld := common.start
	assert.Assertf(beginBufferNew >= beginBufferOld, "new start %d >= old start %d", beginBufferNew, beginBufferOld)
	assert.Assertf(refcnt >= 1, "new refcnt %d >= %d", refcnt, 1)

	i := beginBufferNew - beginBufferOld
	common.bytes = common.bytes[i:]
	common.start = beginBufferNew
	common.refcnt = refcnt
}

type bufferedBody struct {
	mu     sync.Mutex
	common *bufferedCommon
	offset int64
	eof    bool
	closed bool
}

func (body *bufferedBody) BytesRemaining() int64 {
	body.mu.Lock()
	defer body.mu.Unlock()

	if body.closed {
		return 0
	}

	if body.eof {
		return 0
	}

	common := body.common
	common.mu.Lock()
	defer common.mu.Unlock()

	if common.length < 0 {
		return -1
	}

	beginOffset := body.offset
	endOffset := common.length
	if beginOffset > endOffset {
		beginOffset = endOffset
	}
	return (endOffset - beginOffset)
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

	if body.eof {
		return 0, common.err
	}

	beginBuffer := common.start
	endBuffer := beginBuffer + int64(len(common.bytes))
	beginOffset := body.offset
	endOffset := beginOffset + int64(len(p))

	assert.Assertf(beginOffset >= beginBuffer, "%d >= %d", beginOffset, beginBuffer)

	if common.err == nil && endOffset > endBuffer {
		numToRead := (endOffset - endBuffer)
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
				common.length = endBuffer
				common.err = err
				break
			}
		}

		if eof && common.err == nil {
			common.length = endBuffer
			common.err = io.EOF
		}
	}

	var err error
	if endOffset > endBuffer {
		endOffset = endBuffer
		err = common.err
		body.eof = true
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
	body.eof = false
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
		eof:    body.eof,
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
