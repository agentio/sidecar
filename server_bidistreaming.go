package sidecar

import (
	"io"
	"net/http"
)

type BidiStream[Req, Res any] struct {
	reader io.ReadCloser
	writer http.ResponseWriter
}

func (b *BidiStream[Req, Res]) Send(msg *Res) error {
	return Send(b.writer, msg)
}

func (b *BidiStream[Req, Res]) Receive() (*Res, error) {
	var response Res
	err := Receive(b.reader, &response)
	return &response, err
}

type bidiStreaming[Req, Res any] func(stream *BidiStream[Req, Res]) error

func BidiStreaming[Req any, Res any](fn bidiStreaming[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		err := fn(&BidiStream[Req, Res]{reader: r.Body, writer: w})
		if err != nil {
			w.Header().Set("Trailer:Grpc-Status", "11")
			return
		}
		w.Header().Set("Trailer:Grpc-Status", "0")
	}
}
