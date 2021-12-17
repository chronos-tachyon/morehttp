package mockreader

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"strconv"
	"sync"

	"github.com/chronos-tachyon/assert"
	"github.com/chronos-tachyon/bufferpool"
)

type Expectation struct {
	op        Op
	input0    []byte
	input1    int64
	input2    int
	input3    string
	output0   int64
	output1   fs.FileInfo
	output2   error
	hasOutput bool
}

func ExpectMark(str string) Expectation {
	return Expectation{
		op:        MarkOp,
		input3:    str,
		hasOutput: true,
	}
}

func ExpectStat(fi fs.FileInfo, err error) Expectation {
	return Expectation{
		op:        StatOp,
		output1:   fi,
		output2:   err,
		hasOutput: true,
	}
}

func ExpectRead(p []byte, n int, err error) Expectation {
	assert.Assertf(n >= 0, "%d >= 0", n)
	assert.Assertf(n <= len(p), "%d <= %d", n, len(p))
	return Expectation{
		op:        ReadOp,
		input0:    p,
		output0:   int64(n),
		output2:   err,
		hasOutput: true,
	}
}

func ExpectReadAt(p []byte, offset int64, n int, err error) Expectation {
	assert.Assertf(n >= 0, "%d >= 0", n)
	assert.Assertf(n <= len(p), "%d <= %d", n, len(p))
	return Expectation{
		op:        ReadAtOp,
		input0:    p,
		input1:    offset,
		output0:   int64(n),
		output2:   err,
		hasOutput: true,
	}
}

func ExpectSeek(offset int64, whence int, newOffset int64, err error) Expectation {
	return Expectation{
		op:        SeekOp,
		input1:    offset,
		input2:    whence,
		output0:   newOffset,
		output2:   err,
		hasOutput: true,
	}
}

func ExpectClose(err error) Expectation {
	return Expectation{
		op:        CloseOp,
		output2:   err,
		hasOutput: true,
	}
}

func (x Expectation) Matches(y Expectation) bool {
	return (x.op == y.op &&
		x.input1 == y.input1 &&
		x.input2 == y.input2 &&
		x.input3 == y.input3 &&
		len(x.input0) == len(y.input0))
}

func (x Expectation) GoStringTo(buf *bytes.Buffer) {
	var tail int

	buf.WriteString("Expectation{")
	buf.WriteString(x.op.GoString())
	buf.WriteString(" | ")

	switch x.op {
	case MarkOp:
		buf.WriteString(strconv.Quote(x.input3))
		tail = 0

	case StatOp:
		buf.WriteString("∅")
		tail = 3

	case ReadOp:
		formatBytesTo(buf, x.input0)
		tail = 2

	case ReadAtOp:
		formatBytesTo(buf, x.input0)
		buf.WriteString(", ")
		buf.WriteString(formatInt64(x.input1))
		tail = 2

	case SeekOp:
		buf.WriteString(formatInt64(x.input1))
		buf.WriteString(", ")
		buf.WriteString(formatInt(x.input2))
		tail = 2

	case CloseOp:
		buf.WriteString("∅")
		tail = 1
	}

	if x.hasOutput {
		buf.WriteString(" | ")
		switch tail {
		case 0:
			buf.WriteString("∅")
		case 1:
			buf.WriteString(formatAny(x.output2))
		case 2:
			buf.WriteString(formatInt64(x.output0))
			buf.WriteString(", ")
			buf.WriteString(formatAny(x.output2))
		case 3:
			buf.WriteString(formatAny(x.output1))
			buf.WriteString(", ")
			buf.WriteString(formatAny(x.output2))
		}
	}

	buf.WriteString("}")
}

func (x Expectation) StringTo(buf *bytes.Buffer) {
	var tail int

	switch x.op {
	case MarkOp:
		buf.WriteString("Mark(")
		buf.WriteString(strconv.Quote(x.input3))
		buf.WriteString(")")
		tail = 0

	case StatOp:
		buf.WriteString("Stat()")
		tail = 3

	case ReadOp:
		buf.WriteString("Read([")
		buf.WriteString(formatInt(len(x.input0)))
		buf.WriteString(" bytes])")
		tail = 2

	case ReadAtOp:
		buf.WriteString("ReadAt([")
		buf.WriteString(formatInt(len(x.input0)))
		buf.WriteString(" bytes], ")
		buf.WriteString(formatInt64(x.input1))
		buf.WriteString(")")
		tail = 2

	case SeekOp:
		buf.WriteString("Seek(")
		buf.WriteString(formatInt64(x.input1))
		buf.WriteString(", ")
		buf.WriteString(whenceString(x.input2))
		buf.WriteString(")")
		tail = 2

	case CloseOp:
		buf.WriteString("Close()")
		tail = 1

	default:
		buf.WriteString("Nothing")
	}

	if x.hasOutput {
		switch tail {
		case 0:
			buf.WriteString(" => ()")
		case 1:
			buf.WriteString(" => (")
			buf.WriteString(formatAny(x.output2))
			buf.WriteString(")")
		case 2:
			buf.WriteString(" => (")
			buf.WriteString(formatInt64(x.output0))
			buf.WriteString(", ")
			buf.WriteString(formatAny(x.output2))
			buf.WriteString(")")
		case 3:
			buf.WriteString(" => (")
			buf.WriteString(formatAny(x.output1))
			buf.WriteString(", ")
			buf.WriteString(formatAny(x.output2))
			buf.WriteString(")")
		}
	}
}

func (x Expectation) GoString() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	x.GoStringTo(buf)
	return buf.String()
}

func (x Expectation) String() string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	x.StringTo(buf)
	return buf.String()
}

var (
	_ fmt.GoStringer = Expectation{}
	_ fmt.Stringer   = Expectation{}
)

func New(list ...Expectation) *MockReader {
	return &MockReader{expect: list}
}

type MockReader struct {
	mu     sync.Mutex
	expect []Expectation
	index  uint
}

func (r *MockReader) next() (uint, Expectation) {
	if r == nil {
		return 0, Expectation{}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	index := r.index
	length := uint(len(r.expect))
	if index >= length {
		return length, Expectation{}
	}

	r.index++
	return index, r.expect[index]
}

func (r *MockReader) Mark(str string) {
	if r == nil {
		return
	}

	index, expect := r.next()
	actual := Expectation{op: MarkOp, input3: str}

	if !expect.Matches(actual) {
		panic(ExpectationFailedError{index, expect, actual})
	}

	debug("mock index %d: matched %+v", index, expect)
}

func (r *MockReader) Stat() (fs.FileInfo, error) {
	index, expect := r.next()
	actual := Expectation{op: StatOp}

	if !expect.Matches(actual) {
		panic(ExpectationFailedError{index, expect, actual})
	}

	debug("mock index %d: matched %+v", index, expect)
	fi := expect.output1
	err := expect.output2
	return fi, err
}

func (r *MockReader) Read(p []byte) (int, error) {
	index, expect := r.next()
	actual := Expectation{op: ReadOp, input0: p}

	if !expect.Matches(actual) {
		panic(ExpectationFailedError{index, expect, actual})
	}

	debug("mock index %d: matched %+v", index, expect)
	n := int(expect.output0)
	err := expect.output2
	if n > 0 {
		copy(p[:n], expect.input0[:n])
	}
	return n, err
}

func (r *MockReader) ReadAt(p []byte, offset int64) (int, error) {
	index, expect := r.next()
	actual := Expectation{op: ReadAtOp, input0: p, input1: offset}

	if !expect.Matches(actual) {
		panic(ExpectationFailedError{index, expect, actual})
	}

	debug("mock index %d: matched %+v", index, expect)
	n := int(expect.output0)
	err := expect.output2
	if n > 0 {
		copy(p[:n], expect.input0[:n])
	}
	return n, err
}

func (r *MockReader) Seek(offset int64, whence int) (int64, error) {
	index, expect := r.next()
	actual := Expectation{op: SeekOp, input1: offset, input2: whence}

	if !expect.Matches(actual) {
		panic(ExpectationFailedError{index, expect, actual})
	}

	debug("mock index %d: matched %+v", index, expect)
	n := expect.output0
	err := expect.output2
	return n, err
}

func (r *MockReader) Close() error {
	index, expect := r.next()
	actual := Expectation{op: CloseOp}

	if !expect.Matches(actual) {
		panic(ExpectationFailedError{index, expect, actual})
	}

	debug("mock index %d: matched %+v", index, expect)
	err := expect.output2
	return err
}

var (
	_ fs.File     = (*MockReader)(nil)
	_ io.Reader   = (*MockReader)(nil)
	_ io.ReaderAt = (*MockReader)(nil)
	_ io.Seeker   = (*MockReader)(nil)
	_ io.Closer   = (*MockReader)(nil)
)
