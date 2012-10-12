package app

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// All handlers in the app must implement this type
type Handler func(r *Request) error

// Serves a http request
func (fn Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-UA-Compatible", "chrome=1")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	r := &Request{Req: req, W: w}

	defer func() {
		if rec := recover(); rec != nil {
			err := fmt.Errorf("panic recovered error: %+v\n%s", rec, debug.Stack())
			r.LogError(err)
			r.internalServerError(err.Error())
		}
	}()

	if err := fn(r); err != nil {
		r.LogError(err)
		r.internalServerError(err.Error())
	}
}
