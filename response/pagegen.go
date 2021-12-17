package response

import (
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/chronos-tachyon/morehttp/body"
)

type PageGenerator interface {
	GenerateRedirectPage(code int, location string) (http.Header, body.Body)
	GenerateErrorPage(code int, err error) (http.Header, body.Body)
}

type defaultPageGenerator struct{}

func (gen *defaultPageGenerator) GenerateRedirectPage(code int, location string) (http.Header, body.Body) {
	const CRLF = "\r\n"
	text := `<!DOCTYPE html>` + CRLF + `<p>Redirecting to <a href="%s">%s</a>...</p>` + CRLF
	text = fmt.Sprintf(text, html.EscapeString(location), html.EscapeString(location))
	raw := []byte(text)

	headers := make(http.Header, 16)
	headers.Set("Location", location)
	headers.Set("Content-Type", "text/html; charset=utf-8")
	headers.Set("Content-Length", strconv.Itoa(len(raw)))
	headers.Set("Cache-Control", "max-age=86400, must-revalidate")

	return headers, body.FromBytes(raw)
}

func (gen *defaultPageGenerator) GenerateErrorPage(code int, err error) (http.Header, body.Body) {
	statusText := http.StatusText(code)
	statusLine := statusText + "\r\n"
	raw := []byte(statusLine)

	headers := make(http.Header, 16)
	headers.Set("Content-Type", "text/plain; charset=utf-8")
	headers.Set("Content-Length", strconv.Itoa(len(raw)))
	headers.Set("Cache-Control", "no-cache")

	return headers, body.FromBytes(raw)
}

var genSingleton PageGenerator = (*defaultPageGenerator)(nil)
