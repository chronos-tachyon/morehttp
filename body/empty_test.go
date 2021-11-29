package body

import (
	"io"
	"testing"
)

func TestEmptyBody(t *testing.T) {
	b := &emptyBody{}

	// Test Length before EOF.

	if expect, actual := int64(0), b.Length(); expect != actual {
		t.Errorf("Length failed: expected %d, got %d", expect, actual)
	}

	// Test zero-length Read before EOF.

	p0 := []byte(nil)
	n, err := b.Read(p0)
	if n != 0 {
		t.Errorf("Read #1 failed: expected %d, got %d", 0, n)
	}
	if err != nil {
		t.Errorf("Read #1 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test one-byte read that hits EOF.

	p1 := make([]byte, 1, 1)
	n, err = b.Read(p1)
	if n != 0 {
		t.Errorf("Read #2 failed: expected %d, got %d", 0, n)
	}
	if !isEOF(err) {
		t.Errorf("Read #2 failed: expected io.EOF, got %v of type %T", err, err)
	}

	// Test Length after EOF.

	if expect, actual := int64(0), b.Length(); expect != actual {
		t.Errorf("Length failed: expected %d, got %d", expect, actual)
	}

	// Test zero-byte read after EOF.

	n, err = b.Read(p0)
	if n != 0 {
		t.Errorf("Read #1 failed: expected %d, got %d", 0, n)
	}
	if !isEOF(err) {
		t.Errorf("Read #1 failed: expected io.EOF, got %v of type %T", err, err)
	}

	// Test zero-byte ReadAt.

	n, err = b.ReadAt(p0, 0)
	if n != 0 {
		t.Errorf("ReadAt #1 failed: expected %d, got %d", 0, n)
	}
	if err != nil {
		t.Errorf("ReadAt #1 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test one-byte ReadAt.

	n, err = b.ReadAt(p1, 0)
	if n != 0 {
		t.Errorf("ReadAt #2 failed: expected %d, got %d", 0, n)
	}
	if !isEOF(err) {
		t.Errorf("ReadAt #2 failed: expected io.EOF, got %v of type %T", err, err)
	}

	// Test that Seek resets the eof flag.

	n64, err := b.Seek(1000, io.SeekStart)
	if n64 != 0 {
		t.Errorf("Seek #1 failed: expected %d, got %d", 0, n64)
	}
	if err != nil {
		t.Errorf("Seek #1 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test WriteTo.

	n64, err = b.WriteTo(io.Discard)
	if n64 != 0 {
		t.Errorf("WriteTo #1 failed: expected %d, got %d", 0, n64)
	}
	if !isEOF(err) {
		t.Errorf("WriteTo #1 failed: expected io.EOF, got %v of type %T", err, err)
	}

	// Test Close.

	err = b.Close()
	if err != nil {
		t.Errorf("Close #1 failed: expected <nil>, got %v of type %T", err, err)
	}

	// Test that all methods except Length return fs.ErrClosed now.

	n64, err = b.Seek(1000, io.SeekStart)
	if n64 != -1 {
		t.Errorf("Seek #2 failed: expected %d, got %d", -1, n64)
	}
	if !isErrClosed(err) {
		t.Errorf("Seek #2 failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	n, err = b.Read(p0)
	if n != 0 {
		t.Errorf("Read #3 failed: expected %d, got %d", 0, n)
	}
	if !isErrClosed(err) {
		t.Errorf("Read #3 failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	n, err = b.ReadAt(p0, 0)
	if n != 0 {
		t.Errorf("ReadAt #3 failed: expected %d, got %d", 0, n)
	}
	if !isErrClosed(err) {
		t.Errorf("ReadAt #3 failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	n64, err = b.WriteTo(io.Discard)
	if n64 != 0 {
		t.Errorf("WriteTo #2 failed: expected %d, got %d", 0, n64)
	}
	if !isErrClosed(err) {
		t.Errorf("WriteTo #2 failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	err = b.Close()
	if !isErrClosed(err) {
		t.Errorf("Close #2 failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}
}
