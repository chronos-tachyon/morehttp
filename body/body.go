// Package body represents the body of an HTTP request or response.
package body

import (
	"fmt"
	"io"
	"io/fs"
	"math"

	"github.com/chronos-tachyon/assert"
)

// Body represents a source of bytes.  The intended use case is for preparing
// HTTP request and response bodies before sending them.
//
// Unlike an io.ReadCloser, a Body can be copied using the Copy method.  This
// returns a second instance of Body which represents the same remaining bytes
// as this Body, but the two bodies have independent cursors and can both
// stream the same bytes, even if the underlying reader does not support such
// an operation.
//
// Body instances MAY implement some of the advanced I/O interfaces defined by
// Go, such as io.Seeker, io.ReaderAt, or io.WriterTo.
//
type Body interface {

	// Read operates as per the usual io.Reader contract, filling p with as
	// many bytes as it can and then returning the number of bytes filled.
	// It is allowed to return short reads without error.
	//
	Read(p []byte) (int, error)

	// Close closes the Body, releasing as many resources as possible.
	// Just like an *os.File, the owner of the Body must arrange for Close
	// to be called when the Body is no longer in use.
	//
	Close() error

	// Copy returns a copy of this Body.  The resulting copy represents the
	// same bytes at the same position, but the original and the copy
	// always have independent cursors.
	//
	// Note: this operation can be very expensive.
	//
	// If this Body implements any advanced I/O interfaces, then so will
	// the returned copy.  As the two instances have independent cursors,
	// calling Seek on one Body will not affect the other Body in any
	// meaningful way.
	//
	// For Body instances which attach to a Reader, calling Close on a Body
	// will cause that Body to detach itself from the Reader.  If the
	// Reader implements io.Closer, then the last detach calls Close on the
	// Reader, and the resulting error (if any) is returned to the caller
	// of Close on the last open Body.
	//
	Copy() (Body, error)

	// Unwrap will return the underlying io.Reader, or nil if no such open
	// Reader is currently attached to this Body.
	//
	// Beware that reading from the reader or seeking the cursor can have
	// unexpected visible effects, both on this Body and on any other Body
	// which shares this Body's Reader.
	//
	Unwrap() io.Reader
}

var _ io.ReadCloser = Body(nil)

// AlreadyClosed returns a Body which has already been closed.
func AlreadyClosed() Body {
	return closedSingleton
}

// Empty returns a zero-length Body which is open but already at EOF.
func Empty() Body {
	return &emptyBody{}
}

// Bytes returns a new Body which serves a slice of bytes.
func Bytes(data []byte) Body {
	if len(data) == 0 {
		return Empty()
	}
	return &bytesBody{data: data}
}

// Reader returns a new Body which serves the contents of a Reader.
//
// If the provided Reader also implements io.Closer, then the call to
// Body.Close() will be forwarded to Reader.Close() if there are no other Body
// instances still attached to the Reader.
//
// The provided implementation of Body may take advantage of more advanced
// interfaces provided by the Reader's concrete implementation, such as
// fs.File, io.ReaderAt, or io.Seeker.  Some of these MAY be reflected onto the
// returned Body.
//
func Reader(r io.Reader) (Body, error) {
	assert.NotNil(&r)

	f, fOK := r.(fs.File)
	s, sOK := r.(io.ReadSeeker)
	at, atOK := r.(io.ReaderAt)

	var length int64 = -1

	if fOK {
		fi, err := f.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to determine length of body via Stat: %w", err)
		}

		length = fi.Size()
		assert.Assertf(length >= 0, "FileInfo.Size must return %d >= 0", length)
	}

	if sOK && length < 0 {
		var err error
		length, err = s.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to determine length of body via Seek: %w", err)
		}
		assert.Assertf(length >= 0, "Seek must return %d >= 0", length)

		_, err = s.Seek(0, io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("failed to rewind to start of file via Seek: %w", err)
		}
	}

	if atOK {
		if length < 0 {
			length = math.MaxInt64
		}
		common := &readerAtCommon{
			r:      r,
			at:     at,
			length: length,
		}
		body := &readerAtBody{common: common}
		common.ref()
		return body, nil
	}

	if sOK {
		if length < 0 {
			length = math.MaxInt64
		}
		common := &seekerCommon{
			s:      s,
			length: length,
		}
		body := &seekerBody{common: common}
		common.ref()
		return body, nil
	}

	common := &bufferedCommon{
		r:      r,
		bodies: make(map[*bufferedBody]struct{}, 4),
	}
	body := &bufferedBody{common: common}
	common.ref(body)
	return body, nil
}
