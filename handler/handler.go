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

type Adaptor struct {
	Inner Handler
}

func (a Adaptor) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp := a.Inner.Handle(req)
	err := resp.Serve(w)
	if err != nil {
		OnError(err)
	}
}

var _ http.Handler = Adaptor{}

func DefaultOnError(err error) {
	panic(err)
}

var OnError func(error) = DefaultOnError
