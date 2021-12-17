package response

import (
	"fmt"
	"io"
	"net/http"

	"github.com/chronos-tachyon/morehttp/body"
)

// Response represents an HTTP response.
type Response struct {
	code int
	hdrs http.Header
	body body.Body
	err  error
}

// Status returns the HTTP status code of the response.
//
// The returned value lies between 200 and 999 inclusive.
//
func (resp *Response) Status() int {
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

// Copy returns a copy of this Response.
func (resp *Response) Copy() (*Response, error) {
	body2, err := copyBody(resp.body)
	if err != nil {
		return nil, err
	}

	out := &Response{
		code: resp.code,
		hdrs: resp.hdrs,
		body: body2,
		err:  resp.err,
	}
	return out, nil
}

// Serve serves the Response via the given ResponseWriter, consuming its Body.
func (resp *Response) Serve(w http.ResponseWriter) error {
	h := w.Header()
	for k, vlist := range resp.hdrs {
		h[k] = vlist
	}

	w.WriteHeader(resp.code)

	_, err := io.Copy(w, resp.body)

	err2 := resp.body.Close()
	if err == nil {
		err = err2
	}

	return err
}
