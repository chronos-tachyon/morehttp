package response

import (
	"fmt"
	"io"
	"net/http"

	"github.com/chronos-tachyon/morehttp/body"
)

type Response struct {
	code int
	hdrs http.Header
	body body.Body
	err  error
}

// StatusCode returns the HTTP status code of the response.
//
// The returned value lies between 200 and 999 inclusive.
//
func (resp *Response) StatusCode() int {
	return resp.code
}

// Headers returns the HTTP headers of the response.
//
// The caller MUST NOT modify the returned map or its contents.
//
func (resp *Response) Headers() http.Header {
	return resp.hdrs
}

// Body returns this Response's Body.
//
// The caller must not read from or modify the returned Body.  Call Body.Copy()
// and operate on the copy when performing such actions.
//
func (resp *Response) Body() body.Body {
	return resp.body
}

// Err returns the Go error which provoked this Response, if any.
func (resp *Response) Err() error {
	return resp.err
}

// String returns a programmer-friendly string description.
func (resp *Response) String() string {
	if bodyLen := resp.body.Length(); bodyLen >= 0 {
		return fmt.Sprintf("[HTTP %03d - %d bytes]", resp.code, bodyLen)
	}
	return fmt.Sprintf("[HTTP %03d - unknown length]", resp.code)
}

// Serve serves the response via the given ResponseWriter.
func (resp *Response) Serve(w http.ResponseWriter) (int64, error) {
	h := w.Header()
	for k, vlist := range resp.hdrs {
		h[k] = vlist
	}
	w.WriteHeader(resp.code)
	n, err := io.Copy(w, resp.body)
	err2 := resp.body.Close()
	if err == nil {
		err = err2
	}
	return n, err
}
