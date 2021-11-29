package body

import (
	"io"
	"testing"
)

func TestClosedBody(t *testing.T) {
	b := &closedBody{}

	// Test Length.

	expectLength := int64(0)
	actualLength := b.Length()
	if expectLength != actualLength {
		t.Errorf("Length failed: expected %d, got %d", expectLength, actualLength)
	}

	// Test that all other calls return fs.ErrClosed.

	p := []byte(nil)
	n, err := b.Read(p)
	if n != 0 {
		t.Errorf("Read failed: expected %d, got %d", 0, n)
	}
	if !isErrClosed(err) {
		t.Errorf("Read failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	n, err = b.ReadAt(p, 0)
	if n != 0 {
		t.Errorf("ReadAt failed: expected %d, got %d", 0, n)
	}
	if !isErrClosed(err) {
		t.Errorf("ReadAt failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	n64, err := b.Seek(0, io.SeekStart)
	if n64 != -1 {
		t.Errorf("Seek failed: expected %d, got %d", -1, n64)
	}
	if !isErrClosed(err) {
		t.Errorf("Seek failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	n64, err = b.WriteTo(io.Discard)
	if n64 != 0 {
		t.Errorf("WriteTo failed: expected %d, got %d", 0, n64)
	}
	if !isErrClosed(err) {
		t.Errorf("WriteTo failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}

	err = b.Close()
	if !isErrClosed(err) {
		t.Errorf("Close failed: expected fs.ErrClosed, got %v of type %T", err, err)
	}
}
