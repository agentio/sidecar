package sidecar

import (
	"io"
	"net/http"
)

// BidiStream provides messaging to bidi streaming handlers.
type BidiStream[Req, Res any] struct {
	reader io.ReadCloser
	writer http.ResponseWriter
}

// Send sends a response message on a bidi stream.
func (b *BidiStream[Req, Res]) Send(msg *Res) error {
	return Send(b.writer, msg)
}

// Receive reads a request message from a bidi stream.
func (b *BidiStream[Req, Res]) Receive() (*Res, error) {
	var response Res
	err := Receive(b.reader, &response)
	return &response, err
}

// Bidi streaming handlers should be functions that implement this interface.
type BidiStreamingFunction[Req, Res any] func(stream *BidiStream[Req, Res]) error

// HandleBidiStreaming wraps a bidi streaming function in an HTTP handler.
func HandleBidiStreaming[Req any, Res any](fn BidiStreamingFunction[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
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
