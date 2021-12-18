// Package body represents the body of an HTTP request or response.
package body

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"

	"github.com/chronos-tachyon/assert"
	"github.com/chronos-tachyon/bufferpool"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// Body represents a source of bytes, extending the io.ReadCloser interface.
// The intended use case is for preparing HTTP request and response bodies
// before sending them.
//
// Unlike an io.ReadCloser, the length of a Body can be retrieved cheaply using
// the BytesRemaining() method, assuming that it is actually known.
//
// Also unlike an io.ReadCloser, a Body can be duplicated using the Copy()
// method.  This returns a second instance of Body that represents the same
// remaining bytes as this Body, but the two bodies have independent cursors
// and can both stream the same bytes without interfering with each other.
// This is true *even if* they share an underlying Reader that does not support
// such an operation.
//
// Body implementations MAY also provide some of the advanced I/O interfaces
// defined by Go, such as io.Seeker, io.ReaderAt, or io.WriterTo.
//
type Body interface {

	// BytesRemaining returns the exact number of bytes remaining in the
	// Body, or -1 if the number of bytes remaining is unknown.
	BytesRemaining() int64

	// Read operates as per the usual io.Reader contract, filling p with as
	// many bytes as it can and then returning the number of bytes filled.
	// It is allowed to return short reads without error.
	//
	Read(p []byte) (int, error)

	// Close operates as per the usual io.Closer contract, closing the Body
	// and releasing as many resources as possible.  As usual, the owner of
	// the Body must arrange for Close() to be called when the Body is no
	// longer in use.
	//
	// Calling Close() on the Body does not necessarily call Close() on the
	// underlying Reader in a 1-to-1 fashion: Body implementations use
	// reference counting during Copy(), and they use that to Close() the
	// Reader only when the last copy is being closed.  As such, this
	// method may return nil even when the eventual call to Close() on the
	// backing Reader will return non-nil.
	//
	// Note: even if the Body implementation in use is known to not rely on
	// a backing Reader, Close() should still be called to free backing
	// memory.
	//
	Close() error

	// Copy returns a copy of this Body.  The resulting copy represents the
	// same bytes at the same position, but the original and the copy
	// always have independent cursors.
	//
	// Note: this operation can be very cheap or very expensive, depending
	// on the implementation.
	//
	// If this Body implements any advanced I/O interfaces, then so will
	// the returned copy.  As the two instances have independent cursors,
	// calling Seek on one Body will not affect the other Body in any
	// meaningful way.
	//
	// For Body instances which attach to a Reader, calling Close() on a
	// Body will cause that Body to detach itself from the Reader using
	// reference counting.  If the Reader implements io.Closer, then the
	// Body calls Close() on the Reader once the reference count reaches
	// zero, and the resulting error (if any) is returned by the Close()
	// which caused the reference counter to reach zero.
	//
	Copy() (Body, error)

	// Unwrap will return the backing io.Reader, or nil if no such open
	// Reader is currently attached to this Body.
	//
	// Beware that reading from the backing Reader directly or seeking its
	// cursor can have unexpected visible effects, both on this Body and on
	// any other copies of it.
	//
	Unwrap() io.Reader
}

var _ io.ReadCloser = Body(nil)

// AlreadyClosed returns a Body which has already been closed.
//
// Current and future implementations make these promises:
//
// - The returned Body implementation will provide io.ReaderAt, io.Seeker, and
//   io.WriterTo.
//
func AlreadyClosed() Body {
	return closedSingleton
}

// Empty returns a zero-length Body which is open but already at EOF.
//
// Current and future implementations make these promises:
//
// - The returned Body implementation will provide io.ReaderAt, io.Seeker, and
//   io.WriterTo.
//
func Empty() Body {
	return &emptyBody{}
}

// FromBytes returns a new Body which serves a slice of bytes.
//
// Note: the returned Body holds a reference to the provided slice.  The slice
// must not be modified for the lifetime of the Body.
//
// Current and future implementations make these promises:
//
// - The returned Body implementation will provide io.ReaderAt, io.Seeker, and
//   io.WriterTo.
//
func FromBytes(data []byte) Body {
	if len(data) == 0 {
		return Empty()
	}
	return &bytesBody{data: data}
}

// FromString returns a new Body which serves bytes from a string.
//
// Current and future implementations make these promises:
//
// - The returned Body implementation will provide io.ReaderAt, io.Seeker, and
//   io.WriterTo.
//
func FromString(data string) Body {
	if len(data) == 0 {
		return Empty()
	}
	raw := []byte(data)
	return &bytesBody{data: raw}
}

// FromJSON returns a new Body which serves a JSON payload.
//
// Current and future implementations make these promises:
//
// - The returned Body implementation will provide io.ReaderAt, io.Seeker, and
//   io.WriterTo.
//
func FromJSON(v interface{}) Body {
	return fromJSONImpl(v, false)
}

// FromPrettyJSON returns a new Body which serves a JSON payload with indents and a terminal newline.
//
// Current and future implementations make these promises:
//
// - The returned Body implementation will provide io.ReaderAt, io.Seeker, and
//   io.WriterTo.
//
func FromPrettyJSON(v interface{}) Body {
	return fromJSONImpl(v, true)
}

func fromJSONImpl(v interface{}, isPretty bool) Body {
	prefix := ""
	indent := ""
	if isPretty {
		indent = "  "
	}

	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	e := json.NewEncoder(buf)
	e.SetEscapeHTML(false)
	e.SetIndent(prefix, indent)

	err := e.Encode(v)
	if err != nil {
		panic(err)
	}

	src := buf.Bytes()
	var dst []byte
	if isPretty {
		numLF := bytes.Count(src, []byte{'\n'})
		dst = make([]byte, 0, len(src)+numLF)
		for _, ch := range src {
			if ch == '\n' {
				dst = append(dst, '\r', '\n')
			} else {
				dst = append(dst, ch)
			}
		}
	} else {
		copy(dst, src)
	}

	return FromBytes(dst)
}

// FromProto returns a new Body which serves a binary protobuf payload.
//
// Current and future implementations make these promises:
//
// - The returned Body implementation will provide io.ReaderAt, io.Seeker, and
//   io.WriterTo.
//
func FromProto(msg proto.Message) Body {
	assert.NotNil(&msg)

	raw, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return FromBytes(raw)
}

// FromProtoText returns a new Body which serves a text protobuf payload.
//
// The MarshalOptions argument MAY be nil, in which case sensible defaults are
// used.
//
// Current and future implementations make these promises:
//
// - The returned Body implementation will provide io.ReaderAt, io.Seeker, and
//   io.WriterTo.
//
func FromProtoText(msg proto.Message, o *prototext.MarshalOptions) Body {
	assert.NotNil(&msg)

	if o == nil {
		o = &prototext.MarshalOptions{
			Multiline: true,
			Indent:    "  ",
		}
	}

	raw, err := o.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return FromBytes(raw)
}

// FromReader returns a new Body which serves bytes from a Reader.
//
// If the provided Reader also implements io.Closer, then the call to
// Body.Close() will be forwarded to Reader.Close() if there are no other Body
// instances still attached to the Reader.
//
// The provided implementation of Body may take advantage of more advanced
// interfaces provided by the Reader's concrete implementation, such as
// fs.File, io.ReaderAt, or io.Seeker.
//
// Current and future implementations make these promises:
//
// - If the Reader implements io.ReaderAt plus io.Seeker, or io.ReaderAt plus
//   fs.File, then the returned implementation of Body will support io.ReaderAt
//   and io.Seeker.
//
// - If the Reader implements io.Seeker but not io.ReaderAt, then the returned
//   implementation of Body will support io.Seeker and io.ReaderAt, the latter
//   through emulation using io.ReadSeeker.
//
func FromReader(r io.Reader) (Body, error) {
	assert.NotNil(&r)

	if b, bOK := r.(Body); bOK {
		return b, nil
	}

	return FromReaderAndLength(r, -1)
}

// FromReaderAndLength returns a new Body which serves bytes from a Reader.
//
// See FromReader for details.
//
func FromReaderAndLength(r io.Reader, length int64) (Body, error) {
	assert.NotNil(&r)

	if length < 0 {
		length = -1
	}

	f, fOK := r.(fs.File)
	s, sOK := r.(io.ReadSeeker)
	at, atOK := r.(io.ReaderAt)

	if fOK && length < 0 {
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

	if atOK && length >= 0 {
		common := &readerAtCommon{
			r:      r,
			at:     at,
			length: length,
		}
		body := &readerAtBody{common: common}
		common.ref()
		return body, nil
	}

	if sOK && length >= 0 {
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
		length: length,
	}
	body := &bufferedBody{common: common}
	common.ref(body)
	return body, nil
}
