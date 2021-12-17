package response

import (
	"net/http"

	"github.com/chronos-tachyon/morehttp/body"
)

func copyStrings(src []string) []string {
	if len(src) == 0 {
		return nil
	}

	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func copyHeaders(src http.Header) http.Header {
	if len(src) == 0 {
		return nil
	}

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
	if src == nil {
		return nil, nil
	}

	return src.Copy()
}
