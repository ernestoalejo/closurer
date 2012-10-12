package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/schema"
)

var schemaDecoder = schema.NewDecoder()
var errHandler Handler = nil

type Request struct {
	Req *http.Request
	W   http.ResponseWriter
}

// Load the request data using gorilla schema into a struct
func (r *Request) LoadData(data interface{}) error {
	if err := r.Req.ParseForm(); err != nil {
		return fmt.Errorf("error parsing the request form: %s", err)
	}

	if err := schemaDecoder.Decode(data, r.Req.Form); err != nil {
		e, ok := err.(schema.MultiError)
		if ok {
			// Delete the invalid path errors
			for k, v := range e {
				if strings.Contains(v.Error(), "schema: invalid path") {
					delete(e, k)
				}
			}

			// Return directly if there are no other kind of errors
			if len(e) == 0 {
				return nil
			}
		}

		// Not a MultiError, log it
		if err != nil {
			return fmt.Errorf("error decoding the schema: %s", err)
		}
	}

	return nil
}

func (r *Request) LoadJsonData(data interface{}) error {
	if err := json.NewDecoder(r.Req.Body).Decode(data); err != nil {
		return fmt.Errorf("error decoding the json: %s", err)
	}

	return nil
}

func (r *Request) EmitJson(data interface{}) error {
	if err := json.NewEncoder(r.W).Encode(data); err != nil {
		return fmt.Errorf("error encoding the json: %s", err)
	}

	return nil
}

func (r *Request) IsPOST() bool {
	return r.Req.Method == "POST"
}

func (r *Request) Path() string {
	u := r.Req.URL.Path
	query := r.Req.URL.RawQuery
	if len(query) > 0 {
		u += query
	}
	return u
}

// It returns a nil error always for easy of use inside the handlers.
// Example: return r.Redirect("/foo")
func (r *Request) Redirect(path string) error {
	http.Redirect(r.W, r.Req, path, http.StatusFound)
	return nil
}

// It returns a nil error always for easy of use inside the handlers.
// Example: return r.RedirectPermanently("/foo")
func (r *Request) RedirectPermanently(path string) error {
	http.Redirect(r.W, r.Req, path, http.StatusMovedPermanently)
	return nil
}

func (r *Request) ExecuteTemplate(names []string, data interface{}) error {
	return RawExecuteTemplate(r.W, names, data)
}

// You shouldn't use this function, but directly return
// an error from the handler.
func (r *Request) internalServerError(message string) {
	if errHandler != nil {
		r.W.WriteHeader(http.StatusInternalServerError)
		err := errHandler(r)
		if err == nil {
			return
		}
		message += err.Error()
	}

	http.Error(r.W, message, http.StatusInternalServerError)
}

func (r *Request) JsonResponse(data interface{}) error {
	if err := json.NewEncoder(r.W).Encode(data); err != nil {
		return fmt.Errorf("cannot serialize the response: %s", err)
	}
	return nil
}

func (r *Request) LogError(err error) {
	log.Printf("ERROR: %s\n", err)
}

func SetErrorHandler(f Handler) {
	errHandler = f
}
