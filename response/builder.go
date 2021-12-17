package response

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/chronos-tachyon/assert"

	"github.com/chronos-tachyon/morehttp/body"
)

func Build() *Builder {
	return &Builder{
		gen:  genSingleton,
		code: 0,
		hdrs: make(http.Header, 16),
		body: nil,
		err:  nil,
	}
}

type Builder struct {
	gen  PageGenerator
	code int
	hdrs http.Header
	body body.Body
	err  error
}

func (builder *Builder) WithPageGenerator(gen PageGenerator) *Builder {
	if gen == nil {
		gen = genSingleton
	}
	builder.gen = gen
	return builder
}

func (builder *Builder) Status() int {
	return builder.code
}

func (builder *Builder) Headers() http.Header {
	if builder.hdrs == nil {
		builder.hdrs = make(http.Header, 16)
	}
	return builder.hdrs
}

func (builder *Builder) Body() body.Body {
	return builder.body
}

func (builder *Builder) Err() error {
	return builder.err
}

func (builder *Builder) WithStatus(code int) *Builder {
	assert.Assertf(code >= 200, "code %03d >= 200", code)
	assert.Assertf(code <= 999, "code %03d <= 999", code)
	builder.code = code
	return builder
}

func (builder *Builder) WithHeaders(hdrs http.Header) *Builder {
	builder.hdrs = copyHeaders(hdrs)
	return builder
}

func (builder *Builder) WithHeader(headerName, headerValue string, appendToExisting bool) *Builder {
	assert.Assert(headerName != "", "empty header name")
	headerName = http.CanonicalHeaderKey(headerName)
	hdrs := builder.Headers()
	x := hdrs[headerName]
	if x != nil && appendToExisting {
		x = append(x, headerValue)
	} else {
		x = make([]string, 1)
		x[0] = headerValue
	}
	hdrs[headerName] = x
	return builder
}

func (builder *Builder) WithoutHeader(headerName string) *Builder {
	assert.Assert(headerName != "", "empty header name")
	headerName = http.CanonicalHeaderKey(headerName)
	hdrs := builder.Headers()
	delete(hdrs, headerName)
	return builder
}

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

func (builder *Builder) WithContentLanguage(str string) *Builder {
	str = strings.TrimSpace(str)
	hdrs := builder.Headers()
	if str != "" {
		hdrs.Set("Content-Language", str)
	} else {
		hdrs.Del("Content-Language")
	}
	return builder
}

func (builder *Builder) WithContentEncoding(str string) *Builder {
	str = strings.TrimSpace(str)
	hdrs := builder.Headers()
	if str != "" {
		hdrs.Set("Content-Encoding", str)
	} else {
		hdrs.Del("Content-Encoding")
	}
	return builder
}

func (builder *Builder) WithDigest(algo string, sum []byte) *Builder {
	assert.Assert(algo != "", "algo must not be empty")
	assert.Assert(len(sum) > 0, "sum must not be empty")
	hdrs := builder.Headers()
	hdrs.Add("Digest", algo+"="+base64.StdEncoding.EncodeToString(sum))
	return builder
}

func (builder *Builder) WithBody(b body.Body) *Builder {
	assert.NotNil(&b)
	builder.body = b
	hdrs := builder.Headers()
	if length := b.Length(); length >= 0 {
		hdrs.Set("Content-Length", strconv.FormatInt(length, 10))
	} else {
		hdrs.Del("Content-Length")
	}
	return builder
}

func (builder *Builder) RedirectPage(code int, location string) *Builder {
	assert.Assertf(code >= 300, "code %03d >= 300", code)
	assert.Assertf(code <= 399, "code %03d <= 399", code)
	h, b := builder.gen.GenerateRedirectPage(code, location)
	builder.code = code
	builder.hdrs = h
	builder.body = b
	builder.err = nil
	return builder
}

func (builder *Builder) ErrorPage(code int, err error) *Builder {
	assert.Assertf(code >= 200, "code %03d >= 200", code)
	assert.Assertf(code <= 999, "code %03d <= 999", code)
	assert.NotNil(&err)
	h, b := builder.gen.GenerateErrorPage(code, err)
	builder.code = code
	builder.hdrs = h
	builder.body = b
	builder.err = err
	return builder
}

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

func (builder *Builder) Build() *Response {
	assert.Assert(builder.code != 0, "must specify status code")
	assert.NotNil(&builder.body)
	return &Response{
		code: builder.code,
		hdrs: builder.hdrs,
		body: builder.body,
		err:  builder.err,
	}
}
