package sidecar

import "net/http"

type unary[Req, Res any] func(request *Request[Req]) (*Response[Res], error)

func Unary[Req any, Res any](fn unary[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		var request Req
		err := Receive(r.Body, &request)
		if err != nil {
			return
		}
		response, err := fn(&Request[Req]{Msg: &request})
		if err != nil {
			w.Header().Set("Trailer:Grpc-Status", "11")
			return
		}
		err = Send(w, response.Msg)
		if err != nil {
			return
		}
		w.Header().Set("Trailer:Grpc-Status", "0")
	}
}
