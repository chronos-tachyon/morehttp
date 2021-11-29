package body

import (
	"io"
	"testing"
)

func TestBytesBody(t *testing.T) {
	b := &bytesBody{data: []byte{'a', 'b', 'c', 'd'}}

	// Test Length at start of file.

	if expect, actual := int64(4), b.Length(); expect != actual {
		t.Errorf("Length #1 failed: expected %d, got %d", expect, actual)
	}

	// Test zero-byte Read at start of file.

	p0 := []byte(nil)
	n, err := b.Read(p0)
	if n != 0 {
		t.Errorf("Read #1 failed: expected %d, got %d", 0, n)
	}
	if err != nil {
		t.Errorf("Read #1 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test four one-byte Reads, advancing from start of file to
	// just before end of file.

	p1 := make([]byte, 1, 1)
	n, err = b.Read(p1)
	if n != 1 {
		t.Errorf("Read #2 failed: expected %d, got %d", 1, n)
	}
	if err != nil {
		t.Errorf("Read #2 failed: expected <nil>, got %v of type %T", err, err)
	}
	if expect, actual := byte('a'), p1[0]; err == nil && expect != actual {
		t.Errorf("Read #2 failed: expected %q, got %q", expect, actual)
	}

	n, err = b.Read(p1)
	if n != 1 {
		t.Errorf("Read #3 failed: expected %d, got %d", 1, n)
	}
	if err != nil {
		t.Errorf("Read #3 failed: expected <nil>, got %v of type %T", err, err)
	}
	if expect, actual := byte('b'), p1[0]; err == nil && expect != actual {
		t.Errorf("Read #3 failed: expected %q, got %q", expect, actual)
	}

	n, err = b.Read(p1)
	if n != 1 {
		t.Errorf("Read #4 failed: expected %d, got %d", 1, n)
	}
	if err != nil {
		t.Errorf("Read #4 failed: expected <nil>, got %v of type %T", err, err)
	}
	if expect, actual := byte('c'), p1[0]; err == nil && expect != actual {
		t.Errorf("Read #4 failed: expected %q, got %q", expect, actual)
	}

	n, err = b.Read(p1)
	if n != 1 {
		t.Errorf("Read #5 failed: expected %d, got %d", 1, n)
	}
	if err != nil {
		t.Errorf("Read #5 failed: expected <nil>, got %v of type %T", err, err)
	}
	if expect, actual := byte('d'), p1[0]; err == nil && expect != actual {
		t.Errorf("Read #5 failed: expected %q, got %q", expect, actual)
	}

	// Test Length just before EOF.

	if expect, actual := int64(0), b.Length(); expect != actual {
		t.Errorf("Length #2 failed: expected %d, got %d", expect, actual)
	}

	// Test zero-byte Read just before EOF.
	//
	// This verifies that EOF is only returned when the eof flag is set.

	n, err = b.Read(p0)
	if n != 0 {
		t.Errorf("Read #6 failed: expected %d, got %d", 0, n)
	}
	if err != nil {
		t.Errorf("Read #6 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test one-byte Read at EOF.

	n, err = b.Read(p1)
	if n != 0 {
		t.Errorf("Read #7 failed: expected %d, got %d", 0, n)
	}
	if !isEOF(err) {
		t.Errorf("Read #7 failed: expected io.EOF, got %v of type %T", err, err)
	}

	// Test Length after EOF.

	if expect, actual := int64(0), b.Length(); expect != actual {
		t.Errorf("Length #3 failed: expected %d, got %d", expect, actual)
	}

	// Test zero-byte Read after EOF.
	//
	// This verifies that the eof flag is respected.

	n, err = b.Read(p0)
	if n != 0 {
		t.Errorf("Read #8 failed: expected %d, got %d", 0, n)
	}
	if !isEOF(err) {
		t.Errorf("Read #8 failed: expected io.EOF, got %v of type %T", err, err)
	}

	// Test zero-byte ReadAt(start of file).
	//
	// This and subsequent ReadAt tests verify that the eof flag does not
	// affect ReadAt.

	n, err = b.ReadAt(p0, 0)
	if n != 0 {
		t.Errorf("ReadAt #1 failed: expected %d, got %d", 0, n)
	}
	if err != nil {
		t.Errorf("ReadAt #1 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test one-byte ReadAt(start of file).

	n, err = b.ReadAt(p1, 0)
	if n != 1 {
		t.Errorf("ReadAt #2 failed: expected %d, got %d", 1, n)
	}
	if err != nil {
		t.Errorf("ReadAt #2 failed: expected <nil>, got %v of type %T", err, err)
	}
	if err == nil && p1[0] != 'a' {
		t.Errorf("ReadAt #2 failed: expected %q, got %q", 'a', p1[0])
	}

	// Test zero-byte ReadAt(end of file).

	n, err = b.ReadAt(p0, 4)
	if n != 0 {
		t.Errorf("ReadAt #3 failed: expected %d, got %d", 0, n)
	}
	if err != nil {
		t.Errorf("ReadAt #3 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test one-byte ReadAt(end of file).

	n, err = b.ReadAt(p1, 4)
	if n != 0 {
		t.Errorf("ReadAt #4 failed: expected %d, got %d", 0, n)
	}
	if !isEOF(err) {
		t.Errorf("ReadAt #4 failed: expected io.EOF, got %v of type %T", err, err)
	}

	// Test Seek(beyond EOF)

	n64, err := b.Seek(1000, io.SeekStart)
	if n64 != 4 {
		t.Errorf("Seek #1 failed: expected %d, got %d", 4, n64)
	}
	if err != nil {
		t.Errorf("Seek #1 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test zero-byte Read just before EOF.
	//
	// This verifies that Seek clears the eof flag.

	n, err = b.Read(p0)
	if n != 0 {
		t.Errorf("Read #9 failed: expected %d, got %d", 0, n)
	}
	if err != nil {
		t.Errorf("Read #9 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test Seek(start of file).

	n64, err = b.Seek(0, io.SeekStart)
	if n64 != 0 {
		t.Errorf("Seek #2 failed: expected %d, got %d", 0, n64)
	}
	if err != nil {
		t.Errorf("Seek #2 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test WriteTo at start of file.

	n64, err = b.WriteTo(io.Discard)
	if n64 != 4 {
		t.Errorf("WriteTo #1 failed: expected %d, got %d", 4, n64)
	}
	if err != nil {
		t.Errorf("WriteTo #1 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test WriteTo after EOF.

	n64, err = b.WriteTo(io.Discard)
	if n64 != 0 {
		t.Errorf("WriteTo #2 failed: expected %d, got %d", 0, n64)
	}
	if err != nil {
		t.Errorf("WriteTo #2 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test Close.

	err = b.Close()
	if err != nil {
		t.Errorf("Close #1 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test that all methods except Length return fs.ErrClosed now.

	if expect, actual := int64(0), b.Length(); expect != actual {
		t.Errorf("Length #4 failed: expected %d, got %d", expect, actual)
	}

	n, err = b.Read(p0)
	if n != 0 {
		t.Errorf("Read #10 failed: expected %d, got %d", 0, n)
	}
	if !isErrClosed(err) {
		t.Errorf("Read #10 failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	n64, err = b.Seek(0, io.SeekStart)
	if n64 != 0 {
		t.Errorf("Seek #3 failed: expected %d, got %d", 0, n64)
	}
	if !isErrClosed(err) {
		t.Errorf("Seek #3 failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	n, err = b.ReadAt(p0, 0)
	if n != 0 {
		t.Errorf("ReadAt #5 failed: expected %d, got %d", 0, n)
	}
	if !isErrClosed(err) {
		t.Errorf("ReadAt #5 failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	n64, err = b.WriteTo(io.Discard)
	if n64 != 0 {
		t.Errorf("WriteTo #3 failed: expected %d, got %d", 0, n64)
	}
	if !isErrClosed(err) {
		t.Errorf("WriteTo #3 failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	err = b.Close()
	if !isErrClosed(err) {
		t.Errorf("Close #2 failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}
}
