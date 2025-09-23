package sidecar

import (
	"net/http"
)

// Unary handlers should be functions that implement this interface.
type UnaryFunction[Req, Res any] func(request *Request[Req]) (*Response[Res], error)

// HandleUnary wraps a unary function in an HTTP handler.
func HandleUnary[Req any, Res any](fn UnaryFunction[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		var request Req
		var response *Response[Res]
		err := Receive(r.Body, &request)
		if err != nil {
			goto end
		}
		response, err = fn(&Request[Req]{Msg: &request})
		if err != nil {
			goto end
		}
		err = Send(w, response.Msg)
	end:
		WriteTrailer(w, err)
	}
}
