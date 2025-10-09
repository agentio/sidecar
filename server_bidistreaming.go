package sidecar

import (
	"context"
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
func (b *BidiStream[Req, Res]) Receive() (*Req, error) {
	var request Req
	err := Receive(b.reader, &request)
	return &request, err
}

// Bidi streaming handlers should be functions that implement this interface.
type BidiStreamingFunction[Req, Res any] func(ctx context.Context, stream *BidiStream[Req, Res]) error

// HandleBidiStreaming wraps a bidi streaming function in an HTTP handler.
func HandleBidiStreaming[Req any, Res any](fn BidiStreamingFunction[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		err := fn(r.Context(), &BidiStream[Req, Res]{reader: r.Body, writer: w})
		WriteTrailer(w, err)
	}
}
