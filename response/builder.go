package response

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/chronos-tachyon/assert"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	"github.com/chronos-tachyon/morehttp/body"
)

func NewBuilder() *Builder {
	return &Builder{}
}

type Builder struct {
	gen  PageGenerator
	code int
	hdrs http.Header
	body body.Body
	err  error
}

// PageGenerator returns the associated PageGenerator instance, or
// DefaultPageGenerator if not set.
func (builder *Builder) PageGenerator() PageGenerator {
	if builder.gen == nil {
		return DefaultPageGenerator
	}
	return builder.gen
}

// Status returns the associated HTTP status code, or 0 if not set.
func (builder *Builder) Status() int {
	return builder.code
}

// Headers returns the associated HTTP response headers.
//
// The caller is allowed to mutate the returned map and its contents.
//
func (builder *Builder) Headers() http.Header {
	if builder.hdrs == nil {
		builder.hdrs = make(http.Header, 16)
	}
	return builder.hdrs
}

// Body returns the associated HTTP response body.
//
// The caller must not read from or modify the returned Body.  Call Body.Copy()
// and operate on the copy when performing such actions.
//
func (builder *Builder) Body() body.Body {
	return builder.body
}

// Err returns the associated Go error, or nil if not set.
func (builder *Builder) Err() error {
	return builder.err
}

// WithPageGenerator associates the given PageGenerator instance with this
// Builder.  This affects future calls to RedirectPage or ErrorPage.
//
// The given value MUST NOT be nil.
//
func (builder *Builder) WithPageGenerator(gen PageGenerator) *Builder {
	assert.NotNil(&gen)
	builder.gen = gen
	return builder
}

// WithStatus associates the given HTTP status code with this Builder.
//
// The given value MUST lie between 200 and 999 inclusive.
//
func (builder *Builder) WithStatus(code int) *Builder {
	assert.Assertf(code >= 200, "code %03d >= 200", code)
	assert.Assertf(code <= 999, "code %03d <= 999", code)
	builder.code = code
	return builder
}

// WithHeaders associates the given HTTP response headers with this Builder.
//
// This method makes a deep copy of the given map and its contents, so that no
// references to it are retained.
//
func (builder *Builder) WithHeaders(hdrs http.Header) *Builder {
	builder.hdrs = copyHeaders(hdrs)
	return builder
}

// WithHeader adds the given header to the HTTP response.
func (builder *Builder) WithHeader(name string, value string, appendToExisting bool) *Builder {
	assert.Assert(name != "", "header name must not be empty")
	hdrs := builder.Headers()
	if appendToExisting {
		hdrs.Add(name, value)
	} else {
		hdrs.Set(name, value)
	}
	return builder
}

// WithoutHeader removes the given header from the HTTP response.
func (builder *Builder) WithoutHeader(name string) *Builder {
	assert.Assert(name != "", "header name must not be empty")
	hdrs := builder.Headers()
	hdrs.Del(name)
	return builder
}

// WithContentType sets the Content-Type header.
func (builder *Builder) WithContentType(str string) *Builder {
	mediaType, params, err := mime.ParseMediaType(str)
	if err != nil {
		panic(fmt.Errorf("failed to parse Content-Type header: %q: %w", str, err))
	}

	str = mime.FormatMediaType(mediaType, params)
	if str == "" {
		panic(fmt.Errorf("failed to format MIME media type: %q %+v", mediaType, params))
	}

	hdrs := builder.Headers()
	hdrs.Set("Content-Type", str)
	return builder
}

// WithContentLanguage sets the Content-Language header.
func (builder *Builder) WithContentLanguage(str string) *Builder {
	str = strings.TrimSpace(str)
	hdrs := builder.Headers()
	hdrs.Set("Content-Language", str)
	return builder
}

// WithContentEncoding sets the Content-Encoding header.
func (builder *Builder) WithContentEncoding(str string) *Builder {
	str = strings.TrimSpace(str)
	hdrs := builder.Headers()
	hdrs.Set("Content-Encoding", str)
	return builder
}

// WithETag sets the ETag header.
func (builder *Builder) WithETag(tag string, isStrong bool) *Builder {
	var str string
	if !isStrong {
		str = `W/`
	}
	str = str + `"` + tag + `"`
	hdrs := builder.Headers()
	hdrs.Set("Etag", str)
	return builder
}

// WithDigest adds the given Digest header.
func (builder *Builder) WithDigest(algo string, sum []byte) *Builder {
	assert.Assert(algo != "", "algo must not be empty")
	assert.Assert(len(sum) > 0, "sum must not be empty")
	encodedSum := base64.StdEncoding.EncodeToString(sum)
	hdrs := builder.Headers()
	hdrs.Add("Digest", algo+"="+encodedSum)
	return builder
}

// WithBody associates the given Body with this Builder, taking ownership of it.
//
// The given value MUST NOT be nil.
//
func (builder *Builder) WithBody(b body.Body) *Builder {
	assert.NotNil(&b)
	builder.body = b
	return builder
}

// WithJSON converts the given value to a JSON Body, then associates it with
// this Builder.
//
// This method also adds the header "Content-Type: application/json".
//
func (builder *Builder) WithJSON(v interface{}) *Builder {
	builder.body = body.FromJSON(v)
	hdrs := builder.Headers()
	hdrs.Set("Content-Type", "application/json")
	return builder
}

// WithPrettyJSON converts the given value to a human-formatted JSON Body, then
// associates it with this Builder.
//
// This method also adds the header "Content-Type: application/json".
//
func (builder *Builder) WithPrettyJSON(v interface{}) *Builder {
	builder.body = body.FromPrettyJSON(v)
	hdrs := builder.Headers()
	hdrs.Set("Content-Type", "application/json")
	return builder
}

// WithProto converts the given value to a binary protobuf Body, then
// associates it with this Builder.
//
// This method also adds the header "Content-Type: application/vnd.google.protobuf".
//
func (builder *Builder) WithProto(msg proto.Message) *Builder {
	builder.body = body.FromProto(msg)
	hdrs := builder.Headers()
	hdrs.Set("Content-Type", "application/vnd.google.protobuf")
	return builder
}

// WithProtoText converts the given value to a text protobuf Body, then
// associates it with this Builder.
//
// This method also adds the header "Content-Type: text/plain; charset=utf-8".
//
func (builder *Builder) WithProtoText(msg proto.Message, o *prototext.MarshalOptions) *Builder {
	builder.body = body.FromProtoText(msg, o)
	hdrs := builder.Headers()
	hdrs.Set("Content-Type", "text/plain; charset=utf-8")
	return builder
}

// WithError associates the given Go error with this Builder.
//
// The given value MAY be nil.
//
func (builder *Builder) WithError(err error) *Builder {
	builder.err = err
	return builder
}

// RedirectPage populates this Builder with a generated redirect response.
func (builder *Builder) RedirectPage(code int, location string) *Builder {
	location = strings.TrimSpace(location)

	assert.Assertf(code >= 300, "code %03d >= 300", code)
	assert.Assertf(code <= 399, "code %03d <= 399", code)
	assert.Assert(location != "", "location must not be empty")

	gen := builder.PageGenerator()
	h, b := gen.GenerateRedirectPage(code, location)

	builder.code = code
	builder.hdrs = h
	builder.body = b
	builder.err = nil
	return builder
}

// ErrorPage populates this Builder with a generated error response.
func (builder *Builder) ErrorPage(code int, err error) *Builder {
	assert.Assertf(code >= 200, "code %03d >= 200", code)
	assert.Assertf(code <= 999, "code %03d <= 999", code)
	assert.NotNil(&err)

	gen := builder.PageGenerator()
	h, b := gen.GenerateErrorPage(code, err)

	builder.code = code
	builder.hdrs = h
	builder.body = b
	builder.err = err
	return builder
}

// Copy creates a copy of this Builder, duplicating the Body as needed.
func (builder *Builder) Copy() (*Builder, error) {
	body2, err := copyBody(builder.body)
	if err != nil {
		return nil, err
	}

	hdrs2 := copyHeaders(builder.hdrs)

	out := &Builder{
		gen:  builder.gen,
		code: builder.code,
		hdrs: hdrs2,
		body: body2,
		err:  builder.err,
	}
	return out, nil
}

// Build returns a new Response with the properties specified on this Builder.
//
// A Body MUST already have been specified.  The Response takes ownership of
// the Body.
//
// If the Content-Type header has not been specified, then the Content-Type
// header is automatically populated as "application/octet-stream".
//
// If the Content-Length header has not been specified AND the Body has a
// non-negative Length(), then the Content-Length header is automatically
// populated from Length().
//
// After calling this method, the Builder is reset to an empty state and is
// ready to build another Response.  Only the PageGenerator is retained.
//
func (builder *Builder) Build() *Response {
	assert.Assert(builder.body != nil, "must specify body")

	code := builder.code
	hdrs := builder.hdrs
	body := builder.body
	err := builder.err

	builder.code = 0
	builder.hdrs = nil
	builder.body = nil
	builder.err = nil

	if code == 0 {
		code = http.StatusOK
	}

	if hdrs == nil {
		hdrs = make(http.Header, 16)
	}

	contentType := http.CanonicalHeaderKey("Content-Type")
	if _, found := hdrs[contentType]; !found {
		v := make([]string, 1)
		v[0] = "application/octet-stream"
		hdrs[contentType] = v
	}

	bodyLen := body.Length()
	if bodyLen >= 0 {
		contentLength := http.CanonicalHeaderKey("Content-Length")
		if _, found := hdrs[contentLength]; !found {
			v := make([]string, 1)
			v[0] = strconv.FormatInt(bodyLen, 10)
			hdrs[contentLength] = v
		}
	}

	return &Response{
		code: code,
		hdrs: hdrs,
		body: body,
		err:  err,
	}
}
