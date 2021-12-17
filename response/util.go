package response

import (
	"net/http"

	"github.com/chronos-tachyon/morehttp/body"
)

func copyStrings(src []string) []string {
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func copyHeaders(src http.Header) http.Header {
	srcLen := len(src)
	if srcLen < 16 {
		srcLen = 16
	}

	dst := make(http.Header, srcLen)
	for k, vlist := range src {
		dst[k] = copyStrings(vlist)
	}
	return dst
}

func copyBody(src body.Body) (body.Body, error) {
	var dst body.Body
	if src != nil {
		var err error
		dst, err = src.Copy()
		if err != nil {
			return nil, err
		}
	}
	return dst, nil
}
