package main

import (
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/gorilla/schema"
)

var schemaDecoder = schema.NewDecoder()

type Request struct {
	Req *http.Request
	W   http.ResponseWriter
}

func (r *Request) LoadData(data interface{}) {
	if err := r.Req.ParseForm(); err != nil {
		panic(err)
	}

	if err := schemaDecoder.Decode(data, r.Req.Form); err != nil {
		panic(err)
	}
}

func (r *Request) IsPOST() bool {
	return r.Req.Method == "POST"
}

func (r *Request) Path() string {
	return r.Req.URL.Path + "?" + r.Req.URL.RawQuery
}

// It returns a nil error always for easy of use inside the handlers.
// Example: return r.Redirect("/test")
func (r *Request) Redirect(path string) error {
	http.Redirect(r.W, r.Req, path, http.StatusFound)
	return nil
}

func (r *Request) NotFound(message string) error {
	http.Error(r.W, message, 404)
	return nil
}

func (r *Request) Forbidden(message string) error {
	http.Error(r.W, message, 403)
	return nil
}

func (r *Request) InternalServerError(message string) error {
	http.Error(r.W, message, 500)
	return nil
}

// All handlers in the app must implement this type
type Handler func(r *Request) error

// Serves a http request
func (fn Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Ask for chrome frame if we're in MSIE
	w.Header().Set("X-UA-Compatible", "chrome=1")

	// Create the request
	r := &Request{
		Req: req,
		W:   w,
	}

	// Defers the panic recovering
	defer func() {
		if rec := recover(); rec != nil {
			err := fmt.Errorf("panic recovered error: %+v", rec)
			r.InternalServerError(err.Error())
			log.Println(err)
		}
	}()

	if err := fn(r); err != nil {
		r.InternalServerError(err.Error())
		log.Println(err)
	}
}
