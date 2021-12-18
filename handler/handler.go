package handler

import (
	"net/http"

	"github.com/chronos-tachyon/morehttp/response"
)

type Handler interface {
	Handle(r *http.Request) response.Response
}

type HandlerFunc func(*http.Request) response.Response

func (fn HandlerFunc) Handle(req *http.Request) response.Response {
	return fn(req)
}

var _ Handler = HandlerFunc(nil)
