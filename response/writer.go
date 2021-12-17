package response

import (
	"bufio"
	"io"
	"net"
	"net/http"

	"github.com/chronos-tachyon/assert"
)

type Writer interface {
	http.ResponseWriter
	http.Pusher
	MaybeWriteHeader(int)
	Status() int
	BytesWritten() int64
	SawError() bool
	Unwrap() http.ResponseWriter
}

func NewWriter(w http.ResponseWriter, r *http.Request) Writer {
	assert.NotNil(&w)
	assert.NotNil(&r)

	isHEAD := (r.Method == http.MethodHead)

	if _, ok := w.(fancy); ok {
		return &fancyWriter{basicWriter{next: w, isHEAD: isHEAD}}
	}

	if _, ok := w.(flush); ok {
		return &flushWriter{basicWriter{next: w, isHEAD: isHEAD}}
	}

	return &basicWriter{next: w, isHEAD: isHEAD}
}

type basicWriter struct {
	next     http.ResponseWriter
	bytes    int64
	status   int
	isHEAD   bool
	sawError bool
}

func (w *basicWriter) Header() http.Header {
	return w.next.Header()
}

func (w *basicWriter) writeHeaderImpl(status int) {
	w.status = status
	w.next.WriteHeader(status)
}

func (w *basicWriter) MaybeWriteHeader(status int) {
	assert.Assertf(status >= 100, "%d >= %d", status, 100)
	assert.Assertf(status <= 999, "%d <= %d", status, 999)
	if w.status != 0 {
		return
	}
	w.writeHeaderImpl(status)
}

func (w *basicWriter) WriteHeader(status int) {
	assert.Assertf(status >= 100, "%d >= %d", status, 100)
	assert.Assertf(status <= 999, "%d <= %d", status, 999)
	if w.status != 0 {
		w.sawError = true
		assert.Raisef("multiple calls to WriteHeader: previous status was %d, new status is %d", w.status, status)
		return
	}
	w.writeHeaderImpl(status)
}

func (w *basicWriter) Write(p []byte) (int, error) {
	w.MaybeWriteHeader(http.StatusOK)

	if w.isHEAD || w.status == http.StatusNoContent {
		return len(p), nil
	}

	n, err := w.next.Write(p)
	assert.Assertf(n >= 0, "%d >= 0", n)
	w.bytes += int64(n)
	if err != nil {
		w.sawError = true
	}
	return n, err
}

func (w *basicWriter) Push(url string, opts *http.PushOptions) error {
	if x, ok := w.next.(http.Pusher); ok {
		return x.Push(url, opts)
	}
	return http.ErrNotSupported
}

func (w *basicWriter) Status() int {
	return w.status
}

func (w *basicWriter) BytesWritten() int64 {
	return w.bytes
}

func (w *basicWriter) SawError() bool {
	return w.sawError
}

func (w *basicWriter) Unwrap() http.ResponseWriter {
	return w.next
}

type flushWriter struct {
	basicWriter
}

func (w *flushWriter) Flush() {
	w.MaybeWriteHeader(http.StatusOK)

	w.next.(http.Flusher).Flush()
}

type fancyWriter struct {
	basicWriter
}

func (w *fancyWriter) CloseNotify() <-chan bool {
	return w.next.(http.CloseNotifier).CloseNotify()
}

func (w *fancyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.next.(http.Hijacker).Hijack()
}

func (w *fancyWriter) ReadFrom(r io.Reader) (int64, error) {
	w.MaybeWriteHeader(http.StatusOK)

	n, err := w.next.(io.ReaderFrom).ReadFrom(r)
	assert.Assertf(n >= 0, "%d >= 0", n)
	w.bytes += n
	return n, err
}

func (w *fancyWriter) Flush() {
	w.MaybeWriteHeader(http.StatusOK)

	w.next.(http.Flusher).Flush()
}

type flush interface {
	http.ResponseWriter
	http.Flusher
}

type fancy interface {
	http.ResponseWriter
	http.CloseNotifier
	http.Hijacker
	http.Flusher
	io.ReaderFrom
}

var (
	_ Writer = (*basicWriter)(nil)
	_ Writer = (*flushWriter)(nil)
	_ flush  = (*flushWriter)(nil)
	_ Writer = (*fancyWriter)(nil)
	_ fancy  = (*fancyWriter)(nil)
)
