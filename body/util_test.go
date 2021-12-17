package body

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"testing"

	"github.com/chronos-tachyon/morehttp/internal/mockreader"
)

type TestOptions struct {
	ClosedMock *mockreader.MockReader
	ClosedBody Body

	EmptyMock *mockreader.MockReader
	EmptyBody Body

	ShortMock              *mockreader.MockReader
	ShortBody              Body
	ShortBodyUnknownLength bool

	NumLengthCalls uint
	NumReadCalls   uint
	NumReadAtCalls uint
	NumSeekCalls   uint
	NumCloseCalls  uint
}

func RunBodyTests(t *testing.T, o *TestOptions) {
	runClosedBodyTests(t, o)
	runEmptyBodyTests(t, o)
	runShortBodyTests(t, o)
}

func runClosedBodyTests(t *testing.T, o *TestOptions) {
	if o.ClosedBody == nil {
		return
	}

	o.ClosedMock.Mark("ClosedBody-Begin")
	runBodyAfterCloseTests(t, o.ClosedMock, o.ClosedBody, o)
	o.ClosedMock.Mark("ClosedBody-End")
}

func runEmptyBodyTests(t *testing.T, o *TestOptions) {
	if o.EmptyBody == nil {
		return
	}

	o.EmptyMock.Mark("EmptyBody-Begin")
	runEmptyBodyReadTests(t, o)
	runEmptyBodySeekTests(t, o)
	runEmptyBodyReadAtTests(t, o)
	runBodyCloseTests(t, o.EmptyMock, o.EmptyBody, o)
	runBodyAfterCloseTests(t, o.EmptyMock, o.EmptyBody, o)
	o.EmptyMock.Mark("EmptyBody-End")
}

func runShortBodyTests(t *testing.T, o *TestOptions) {
	if o.ShortBody == nil {
		return
	}

	o.ShortMock.Mark("ShortBody-Begin")
	runShortBodyReadTests(t, o)
	runShortBodySeekTests(t, o)
	runShortBodyReadAtTests(t, o)
	runBodyCloseTests(t, o.ShortMock, o.ShortBody, o)
	runBodyAfterCloseTests(t, o.ShortMock, o.ShortBody, o)
	o.ShortMock.Mark("ShortBody-End")
}

func runEmptyBodyReadTests(t *testing.T, o *TestOptions) {
	var (
		n   int
		n64 int64
		err error
	)
	p0 := []byte(nil)
	p1 := make([]byte, 1)

	o.EmptyMock.Mark("Read-Begin")

	// Test Length.

	o.NumLengthCalls++
	n64 = o.EmptyBody.Length()
	if expect := int64(0); n64 != expect {
		t.Errorf("Length #%d failed: expected %d, got %d", o.NumLengthCalls, expect, n64)
	}

	// Test zero-byte Read just before EOF.
	//
	// This verifies that EOF is only returned when the eof flag is set.
	//
	// NB: *MockReader is not consulted because the current
	//     file offset is GTE the known file length.

	o.NumReadCalls++
	n, err = o.EmptyBody.Read(p0)
	if expect := 0; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if err != nil {
		t.Errorf("Read #%d failed: expected <nil>, got %s", o.NumReadCalls, formatAny(err))
	}

	// Test one-byte Read at end of file.
	//
	// NB: *MockReader is not consulted because the current
	//     file offset is GTE the known file length.

	o.NumReadCalls++
	fillBytes('x', p1)
	n, err = o.EmptyBody.Read(p1)
	if expect := 0; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if !isEOF(err) {
		t.Errorf("Read #%d failed: expected io.EOF, got %s", o.NumReadCalls, formatAny(err))
	}

	// Test Length after EOF.

	o.NumLengthCalls++
	n64 = o.EmptyBody.Length()
	if expect := int64(0); expect != n64 {
		t.Errorf("Length #%d failed: expected %d, got %d", o.NumReadCalls, expect, n64)
	}

	// Test zero-byte Read after EOF.
	//
	// This verifies that the eof flag is respected.
	//
	// NB: *MockReader is not consulted because the current
	//     file offset is GTE the known file length.

	o.NumReadCalls++
	n, err = o.EmptyBody.Read(p0)
	if expect := 0; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if !isEOF(err) {
		t.Errorf("Read #%d failed: expected io.EOF, got %s", o.NumReadCalls, formatAny(err))
	}

	o.EmptyMock.Mark("Read-End")
}

func runEmptyBodySeekTests(t *testing.T, o *TestOptions) {
	x, ok := o.EmptyBody.(io.Seeker)
	if !ok {
		return
	}

	var (
		n   int
		n64 int64
		err error
	)
	p0 := []byte(nil)

	o.EmptyMock.Mark("Seek-Begin")

	// Test that Seek clears the EOF state.

	o.NumSeekCalls++
	n64, err = x.Seek(0, io.SeekStart)
	if expect := int64(0); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if err != nil {
		t.Errorf("Seek #%d failed: expected <nil>, got %s", o.NumSeekCalls, formatAny(err))
	}

	o.NumReadCalls++
	n, err = o.EmptyBody.Read(p0)
	if expect := 0; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if err != nil {
		t.Errorf("Read #%d failed: expected <nil>, got %s", o.NumReadCalls, formatAny(err))
	}

	// Test that SeekCurrent works.

	o.NumSeekCalls++
	n64, err = x.Seek(0, io.SeekCurrent)
	if expect := int64(0); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if err != nil {
		t.Errorf("Seek #%d failed: expected <nil>, got %s", o.NumSeekCalls, formatAny(err))
	}

	// Test that SeekEnd works.

	o.NumSeekCalls++
	n64, err = x.Seek(0, io.SeekEnd)
	if expect := int64(0); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if err != nil {
		t.Errorf("Seek #%d failed: expected <nil>, got %s", o.NumSeekCalls, formatAny(err))
	}

	// Test that SeekStart past end of file snaps to end of file.

	o.NumSeekCalls++
	n64, err = x.Seek(2, io.SeekStart)
	if expect := int64(0); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if err != nil {
		t.Errorf("Seek #%d failed: expected <nil>, got %s", o.NumSeekCalls, formatAny(err))
	}

	// Test that SeekCurrent past end of file snaps to end of file.

	o.NumSeekCalls++
	n64, err = x.Seek(2, io.SeekCurrent)
	if expect := int64(0); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if err != nil {
		t.Errorf("Seek #%d failed: expected <nil>, got %s", o.NumSeekCalls, formatAny(err))
	}

	// Test that SeekEnd past end of file snaps to end of file.

	o.NumSeekCalls++
	n64, err = x.Seek(2, io.SeekEnd)
	if expect := int64(0); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if err != nil {
		t.Errorf("Seek #%d failed: expected <nil>, got %s", o.NumSeekCalls, formatAny(err))
	}

	// Test that a bogus whence value fails.

	o.NumSeekCalls++
	n64, err = x.Seek(0, 42)
	if expect := int64(-1); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if !isErrUnknownWhence(err, 42) {
		t.Errorf("Seek #%d failed: expected UnknownWhenceSeekError{%d}, got %s", o.NumSeekCalls, 42, formatAny(err))
	}

	// Test that SeekStart fails with negative values.

	o.NumSeekCalls++
	n64, err = x.Seek(-5, io.SeekStart)
	if expect := int64(-1); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if !isErrNegativeStart(err, -5) {
		t.Errorf("Seek #%d failed: expected NegativeStartOffsetSeekError{%d}, got %s", o.NumSeekCalls, -5, formatAny(err))
	}

	// Test that SeekEnd fails with negative values that land before start of file.

	o.NumSeekCalls++
	n64, err = x.Seek(-6, io.SeekCurrent)
	if expect := int64(-1); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if !isErrNegativeComputed(err, -6) {
		t.Errorf("Seek #%d failed: expected NegativeComputedOffsetSeekError{%d}, got %s", o.NumSeekCalls, -6, formatAny(err))
	}

	// Test that SeekEnd fails with negative values that land before start of file.

	o.NumSeekCalls++
	n64, err = x.Seek(-7, io.SeekEnd)
	if expect := int64(-1); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if !isErrNegativeComputed(err, -7) {
		t.Errorf("Seek #%d failed: expected NegativeComputedOffsetSeekError{%d}, got %s", o.NumSeekCalls, -7, formatAny(err))
	}

	o.EmptyMock.Mark("Seek-End")
}

func runEmptyBodyReadAtTests(t *testing.T, o *TestOptions) {
	x, ok := o.EmptyBody.(io.ReaderAt)
	if !ok {
		return
	}

	var (
		n   int
		err error
	)
	p0 := []byte(nil)
	p1 := make([]byte, 1)

	o.EmptyMock.Mark("ReadAt-Begin")

	// Test that zero-byte ReadAt succeeds.

	o.NumReadAtCalls++
	n, err = x.ReadAt(p0, 0)
	if expect := 0; n != expect {
		t.Errorf("ReadAt #%d failed: expected %d, got %d", o.NumReadAtCalls, expect, n)
	}
	if err != nil {
		t.Errorf("ReadAt #%d failed: expected <nil>, got %s", o.NumReadAtCalls, formatAny(err))
	}

	// Test that one-byte ReadAt returns EOF.

	o.NumReadAtCalls++
	fillBytes('x', p1)
	n, err = x.ReadAt(p1, 0)
	if expect := 0; n != expect {
		t.Errorf("ReadAt #%d failed: expected %d, got %d", o.NumReadAtCalls, expect, n)
	}
	if !isEOF(err) {
		t.Errorf("ReadAt #%d failed: expected io.EOF, got %s", o.NumReadAtCalls, formatAny(err))
	}
	if expect, actual := byte('x'), p1[0]; err == nil && expect != actual {
		t.Errorf("ReadAt #%d failed: expected %q, got %q", o.NumReadAtCalls, expect, actual)
	}

	o.EmptyMock.Mark("ReadAt-End")
}

func runShortBodyReadTests(t *testing.T, o *TestOptions) {
	var (
		n   int
		n64 int64
		err error
	)
	p0 := []byte(nil)
	p1 := make([]byte, 1)

	o.ShortMock.Mark("Read-Begin")

	// Test Length at start of file.

	o.NumLengthCalls++
	n64 = o.ShortBody.Length()
	expectLength := int64(4)
	if o.ShortBodyUnknownLength {
		expectLength = -1
	}
	if expect := expectLength; n64 != expect {
		t.Errorf("Length #%d failed: expected %d, got %d", o.NumLengthCalls, expect, n64)
	}

	// Test zero-byte Read at start of file.

	o.NumReadCalls++
	n, err = o.ShortBody.Read(p0)
	if expect := 0; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if err != nil {
		t.Errorf("Read #%d failed: expected <nil>, got %s", o.NumReadCalls, formatAny(err))
	}

	// Test four one-byte Reads, advancing from start of file to just
	// before end of file.

	o.NumReadCalls++
	fillBytes('x', p1)
	n, err = o.ShortBody.Read(p1)
	if expect := 1; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if err != nil {
		t.Errorf("Read #%d failed: expected <nil>, got %s", o.NumReadCalls, formatAny(err))
	}
	if expect, actual := byte('a'), p1[0]; err == nil && expect != actual {
		t.Errorf("Read #%d failed: expected %q, got %q", o.NumReadCalls, expect, actual)
	}

	o.NumReadCalls++
	fillBytes('x', p1)
	n, err = o.ShortBody.Read(p1)
	if expect := 1; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if err != nil {
		t.Errorf("Read #%d failed: expected <nil>, got %s", o.NumReadCalls, formatAny(err))
	}
	if expect, actual := byte('b'), p1[0]; err == nil && expect != actual {
		t.Errorf("Read #%d failed: expected %q, got %q", o.NumReadCalls, expect, actual)
	}

	o.NumReadCalls++
	fillBytes('x', p1)
	n, err = o.ShortBody.Read(p1)
	if expect := 1; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if err != nil {
		t.Errorf("Read #%d failed: expected <nil>, got %s", o.NumReadCalls, formatAny(err))
	}
	if expect, actual := byte('c'), p1[0]; err == nil && expect != actual {
		t.Errorf("Read #%d failed: expected %q, got %q", o.NumReadCalls, expect, actual)
	}

	o.NumReadCalls++
	fillBytes('x', p1)
	n, err = o.ShortBody.Read(p1)
	if expect := 1; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if err != nil {
		t.Errorf("Read #%d failed: expected <nil>, got %s", o.NumReadCalls, formatAny(err))
	}
	if expect, actual := byte('d'), p1[0]; err == nil && expect != actual {
		t.Errorf("Read #%d failed: expected %q, got %q", o.NumReadCalls, expect, actual)
	}

	// Test Length just before EOF.

	o.NumLengthCalls++
	n64 = o.ShortBody.Length()
	if expect := int64(0); n64 != expect {
		t.Errorf("Length #%d failed: expected %d, got %d", o.NumLengthCalls, expect, n64)
	}

	// Test zero-byte Read just before EOF.
	//
	// This verifies that EOF is only returned when the eof flag is set.
	//
	// NB: *MockReader is not consulted because the current
	//     file offset is GTE the known file length.

	o.NumReadCalls++
	n, err = o.ShortBody.Read(p0)
	if expect := 0; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if err != nil {
		t.Errorf("Read #%d failed: expected <nil>, got %s", o.NumReadCalls, formatAny(err))
	}

	// Test one-byte Read at end of file.
	//
	// NB: *MockReader is not consulted because the current
	//     file offset is GTE the known file length.

	o.NumReadCalls++
	fillBytes('x', p1)
	n, err = o.ShortBody.Read(p1)
	if expect := 0; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if !isEOF(err) {
		t.Errorf("Read #%d failed: expected io.EOF, got %s", o.NumReadCalls, formatAny(err))
	}

	// Test Length after EOF.

	o.NumLengthCalls++
	n64 = o.ShortBody.Length()
	if expect := int64(0); expect != n64 {
		t.Errorf("Length #%d failed: expected %d, got %d", o.NumReadCalls, expect, n64)
	}

	// Test zero-byte Read after EOF.
	//
	// This verifies that the eof flag is respected.
	//
	// NB: *MockReader is not consulted because the current
	//     file offset is GTE the known file length.

	o.NumReadCalls++
	n, err = o.ShortBody.Read(p0)
	if expect := 0; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if !isEOF(err) {
		t.Errorf("Read #%d failed: expected io.EOF, got %s", o.NumReadCalls, formatAny(err))
	}

	o.ShortMock.Mark("Read-End")
}

func runShortBodySeekTests(t *testing.T, o *TestOptions) {
	x, ok := o.ShortBody.(io.Seeker)
	if !ok {
		return
	}

	var (
		n   int
		n64 int64
		err error
	)
	p1 := make([]byte, 1)
	p4 := make([]byte, 4)

	o.ShortMock.Mark("Seek-Begin")

	// Test that Seek clears the EOF state.

	o.NumSeekCalls++
	n64, err = x.Seek(0, io.SeekStart)
	if expect := int64(0); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if err != nil {
		t.Errorf("Seek #%d failed: expected <nil>, got %s", o.NumSeekCalls, formatAny(err))
	}

	o.NumReadCalls++
	fillBytes('x', p1)
	n, err = o.ShortBody.Read(p1)
	if expect := 1; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if err != nil {
		t.Errorf("Read #%d failed: expected <nil>, got %s", o.NumReadCalls, formatAny(err))
	}
	if expect, actual := byte('a'), p1[0]; err == nil && expect != actual {
		t.Errorf("Read #%d failed: expected %q, got %q", o.NumReadCalls, expect, actual)
	}

	// Test that SeekCurrent works with positive values.

	o.NumSeekCalls++
	n64, err = x.Seek(2, io.SeekCurrent)
	if expect := int64(3); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if err != nil {
		t.Errorf("Seek #%d failed: expected <nil>, got %s", o.NumSeekCalls, formatAny(err))
	}

	// Test that SeekCurrent works with negative values.

	o.NumSeekCalls++
	n64, err = x.Seek(-2, io.SeekCurrent)
	if expect := int64(1); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if err != nil {
		t.Errorf("Seek #%d failed: expected <nil>, got %s", o.NumSeekCalls, formatAny(err))
	}

	// Test that SeekEnd works with negative values.

	o.NumSeekCalls++
	n64, err = x.Seek(-2, io.SeekEnd)
	if expect := int64(2); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if err != nil {
		t.Errorf("Seek #%d failed: expected <nil>, got %s", o.NumSeekCalls, formatAny(err))
	}

	// Test that a four-byte Read at offset 2 reads 2 bytes and hits EOF.

	o.NumReadCalls++
	fillBytes('x', p4)
	n, err = o.ShortBody.Read(p4)
	if expect := 2; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if !isEOF(err) {
		t.Errorf("Read #%d failed: expected io.EOF, got %s", o.NumReadCalls, formatAny(err))
	}
	if expect, actual := byte('c'), p4[0]; err == nil && expect != actual {
		t.Errorf("Read #%d failed: expected %q, got %q", o.NumReadCalls, expect, actual)
	}
	if expect, actual := byte('d'), p4[1]; err == nil && expect != actual {
		t.Errorf("Read #%d failed: expected %q, got %q", o.NumReadCalls, expect, actual)
	}
	if expect, actual := byte('x'), p4[2]; err == nil && expect != actual {
		t.Errorf("Read #%d failed: expected %q, got %q", o.NumReadCalls, expect, actual)
	}
	if expect, actual := byte('x'), p4[3]; err == nil && expect != actual {
		t.Errorf("Read #%d failed: expected %q, got %q", o.NumReadCalls, expect, actual)
	}

	// Test that a bogus whence value fails.

	o.NumSeekCalls++
	n64, err = x.Seek(0, 42)
	if expect := int64(-1); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if !isErrUnknownWhence(err, 42) {
		t.Errorf("Seek #%d failed: expected UnknownWhenceSeekError{%d}, got %s", o.NumSeekCalls, 42, formatAny(err))
	}

	// Test that SeekStart fails with negative values.

	o.NumSeekCalls++
	n64, err = x.Seek(-5, io.SeekStart)
	if expect := int64(-1); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if !isErrNegativeStart(err, -5) {
		t.Errorf("Seek #%d failed: expected NegativeStartOffsetSeekError{%d}, got %s", o.NumSeekCalls, -5, formatAny(err))
	}

	// Test that SeekEnd fails with negative values that land before start of file.

	o.NumSeekCalls++
	n64, err = x.Seek(-7, io.SeekEnd)
	if expect := int64(-1); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if !isErrNegativeComputed(err, -3) {
		t.Errorf("Seek #%d failed: expected NegativeComputedOffsetSeekError{%d}, got %s", o.NumSeekCalls, -3, formatAny(err))
	}

	// Test that SeekEnd succeeds with positive values but stops at end of file.

	o.NumSeekCalls++
	n64, err = x.Seek(42, io.SeekEnd)
	if expect := int64(4); n64 != expect {
		t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
	}
	if err != nil {
		t.Errorf("Seek #%d failed: expected <nil>, got %s", o.NumSeekCalls, formatAny(err))
	}

	o.ShortMock.Mark("Seek-End")
}

func runShortBodyReadAtTests(t *testing.T, o *TestOptions) {
	x, ok := o.ShortBody.(io.ReaderAt)
	if !ok {
		return
	}

	var (
		n   int
		err error
	)
	p0 := []byte(nil)
	p1 := make([]byte, 1)
	p4 := make([]byte, 4)

	o.ShortMock.Mark("ReadAt-Begin")

	// Test that one-byte ReadAt at start of file succeeds.

	o.NumReadAtCalls++
	fillBytes('x', p1)
	n, err = x.ReadAt(p1, 0)
	if expect := 1; n != expect {
		t.Errorf("ReadAt #%d failed: expected %d, got %d", o.NumReadAtCalls, expect, n)
	}
	if err != nil {
		t.Errorf("ReadAt #%d failed: expected <nil>, got %s", o.NumReadAtCalls, formatAny(err))
	}
	if expect, actual := byte('a'), p1[0]; err == nil && expect != actual {
		t.Errorf("ReadAt #%d failed: expected %q, got %q", o.NumReadAtCalls, expect, actual)
	}

	// Test that zero-byte ReadAt at end of file succeeds.

	o.NumReadAtCalls++
	n, err = x.ReadAt(p0, 4)
	if expect := 0; n != expect {
		t.Errorf("ReadAt #%d failed: expected %d, got %d", o.NumReadAtCalls, expect, n)
	}
	if err != nil {
		t.Errorf("ReadAt #%d failed: expected <nil>, got %s", o.NumReadAtCalls, formatAny(err))
	}

	// Test that one-byte ReadAt at end of file fails with EOF.

	o.NumReadAtCalls++
	fillBytes('x', p1)
	n, err = x.ReadAt(p1, 4)
	if expect := 0; n != expect {
		t.Errorf("ReadAt #%d failed: expected %d, got %d", o.NumReadAtCalls, expect, n)
	}
	if !isEOF(err) {
		t.Errorf("ReadAt #%d failed: expected io.EOF, got %s", o.NumReadAtCalls, formatAny(err))
	}

	// Test that ReadAt reaching end of file partially succeeds with EOF.

	o.NumReadAtCalls++
	fillBytes('x', p4)
	n, err = x.ReadAt(p4, 2)
	if expect := 2; n != expect {
		t.Errorf("ReadAt #%d failed: expected %d, got %d", o.NumReadAtCalls, expect, n)
	}
	if !isEOF(err) {
		t.Errorf("ReadAt #%d failed: expected io.EOF, got %s", o.NumReadAtCalls, formatAny(err))
	}
	if expect, actual := byte('c'), p4[0]; err == nil && expect != actual {
		t.Errorf("ReadAt #%d failed: expected %q, got %q", o.NumReadAtCalls, expect, actual)
	}
	if expect, actual := byte('d'), p4[1]; err == nil && expect != actual {
		t.Errorf("ReadAt #%d failed: expected %q, got %q", o.NumReadAtCalls, expect, actual)
	}
	if expect, actual := byte('x'), p4[2]; err == nil && expect != actual {
		t.Errorf("ReadAt #%d failed: expected %q, got %q", o.NumReadAtCalls, expect, actual)
	}
	if expect, actual := byte('x'), p4[3]; err == nil && expect != actual {
		t.Errorf("ReadAt #%d failed: expected %q, got %q", o.NumReadAtCalls, expect, actual)
	}

	o.ShortMock.Mark("ReadAt-End")
}

func runBodyCloseTests(t *testing.T, m *mockreader.MockReader, b Body, o *TestOptions) {
	m.Mark("Close-Begin")

	o.NumCloseCalls++
	err := b.Close()
	if err != nil {
		t.Errorf("Close #%d failed: expected <nil>, got %s", o.NumCloseCalls, formatAny(err))
	}

	m.Mark("Close-End")
}

func runBodyAfterCloseTests(t *testing.T, m *mockreader.MockReader, b Body, o *TestOptions) {
	var (
		n   int
		n64 int64
		err error
	)
	p0 := []byte(nil)

	m.Mark("AfterClose-Begin")

	// Test Length after close.

	o.NumLengthCalls++
	n64 = b.Length()
	if expect := int64(0); n64 != expect {
		t.Errorf("Length #%d failed: expected %d, got %d", o.NumLengthCalls, expect, n64)
	}

	// Test that all other calls return fs.ErrClosed.

	o.NumReadCalls++
	n, err = b.Read(p0)
	if expect := 0; n != expect {
		t.Errorf("Read #%d failed: expected %d, got %d", o.NumReadCalls, expect, n)
	}
	if !isErrClosed(err) {
		t.Errorf("Read #%d failed: expected fs.ErrClosed, got %s", o.NumReadCalls, formatAny(err))
	}

	if x, ok := b.(io.ReaderAt); ok {
		o.NumReadAtCalls++
		n, err = x.ReadAt(p0, 0)
		if expect := 0; n != expect {
			t.Errorf("ReadAt #%d failed: expected %d, got %d", o.NumReadAtCalls, expect, n)
		}
		if !isErrClosed(err) {
			t.Errorf("ReadAt #%d failed: expected fs.ErrClosed, got %s", o.NumReadAtCalls, formatAny(err))
		}
	}

	if x, ok := b.(io.Seeker); ok {
		o.NumSeekCalls++
		n64, err = x.Seek(0, io.SeekStart)
		if expect := int64(-1); n64 != expect {
			t.Errorf("Seek #%d failed: expected %d, got %d", o.NumSeekCalls, expect, n64)
		}
		if !isErrClosed(err) {
			t.Errorf("Seek #%d failed: expected fs.ErrClosed, got %s", o.NumSeekCalls, formatAny(err))
		}
	}

	o.NumCloseCalls++
	err = b.Close()
	if !isErrClosed(err) {
		t.Errorf("Close #%d failed: expected fs.ErrClosed, got %s", o.NumCloseCalls, formatAny(err))
	}

	m.Mark("AfterClose-End")
}

func fillBytes(ch byte, p []byte) {
	for index := range p {
		p[index] = ch
	}
}

func formatAny(x interface{}) string {
	if x == nil {
		return "<nil>"
	}
	if x == io.EOF {
		return "io.EOF"
	}
	return fmt.Sprintf("%T[%+v]", x, x)
}

func isEOF(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	return errors.Is(err, io.EOF)
}

func isErrClosed(err error) bool {
	if err == nil {
		return false
	}
	if err == fs.ErrClosed {
		return true
	}
	return errors.Is(err, fs.ErrClosed)
}

func isErrUnknownWhence(err error, whence int) bool {
	var xerr UnknownWhenceSeekError
	if errors.As(err, &xerr) {
		return xerr.Whence == whence
	}
	return false
}

func isErrNegativeStart(err error, offset int64) bool {
	var xerr NegativeStartOffsetSeekError
	if errors.As(err, &xerr) {
		return xerr.Offset == offset
	}
	return false
}

func isErrNegativeComputed(err error, offset int64) bool {
	var xerr NegativeComputedOffsetSeekError
	if errors.As(err, &xerr) {
		return xerr.Offset == offset
	}
	return false
}
