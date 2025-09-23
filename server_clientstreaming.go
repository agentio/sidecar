package sidecar

import (
	"io"
	"net/http"
)

type ClientStream[Req any] struct {
	reader io.ReadCloser
}

func (b *ClientStream[Req]) Receive() (*Req, error) {
	var request Req
	err := Receive(b.reader, &request)
	return &request, err
}

type clientStreaming[Req, Res any] func(stream *ClientStream[Req]) (*Response[Res], error)

func ClientStreaming[Req any, Res any](fn clientStreaming[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		response, err := fn(&ClientStream[Req]{reader: r.Body})
		if err != nil {
			return
		}
		err = Send(w, response.Msg)
		if err != nil {
			return
		}
		w.Header().Set("Trailer:Grpc-Status", "0")
	}
}
