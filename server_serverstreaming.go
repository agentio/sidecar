package sidecar

import "net/http"

type ServerStream[Res any] struct {
	writer http.ResponseWriter
}

func (b *ServerStream[Res]) Send(msg *Res) error {
	return Send(b.writer, msg)
}

type serverStreaming[Req, Res any] func(request *Request[Req], stream *ServerStream[Res]) error

func ServerStreaming[Req any, Res any](fn serverStreaming[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		var request Req
		err := Receive(r.Body, &request)
		if err != nil {
			return
		}
		err = fn(&Request[Req]{Msg: &request}, &ServerStream[Res]{writer: w})
		if err != nil {
			w.Header().Set("Trailer:Grpc-Status", "11")
			return
		}
		w.Header().Set("Trailer:Grpc-Status", "0")
	}
}
