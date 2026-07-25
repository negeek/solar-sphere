// Package httpapi holds the HTTP response envelope, request decoding helper,
// and CORS middleware shared by every solar-sphere HTTP service.
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
)

// Response is the JSON envelope every solar-sphere HTTP endpoint responds
// with.
type Response struct {
	StatusCode int         `json:"statuscode"`
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

// JsonResponse writes a Response envelope to w with the given status code.
func JsonResponse(w http.ResponseWriter, success bool, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		StatusCode: statusCode,
		Success:    success,
		Message:    message,
		Data:       data,
	})
}

// Unmarshall decodes JSON from r into strct, which must be a pointer to a
// struct.
func Unmarshall(r io.Reader, strct interface{}) error {
	t := reflect.TypeOf(strct)
	if t == nil || t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return errors.New("strct must be a pointer to a struct")
	}
	return json.NewDecoder(r).Decode(strct)
}

// CORS is a permissive CORS middleware suitable for a public API with no
// cookie-based auth (access keys travel in the Authorization header, not
// cookies, so a wildcard origin does not expose credentials).
func CORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == http.MethodOptions {
			return
		}
		h.ServeHTTP(w, r)
	})
}
