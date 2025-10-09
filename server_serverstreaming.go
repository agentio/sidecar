package sidecar

import (
	"context"
	"net/http"
)

// ServerStream provides messaging to server streaming handlers.
type ServerStream[Res any] struct {
	writer http.ResponseWriter
}

// Send sends a response message on a server stream.
func (b *ServerStream[Res]) Send(msg *Res) error {
	return Send(b.writer, msg)
}

// Server streaming handlers should be functions that implement this interface.
type ServerStreamingFunction[Req, Res any] func(ctx context.Context, request *Request[Req], stream *ServerStream[Res]) error

// HandleServerStreaming wraps a server streaming function in an HTTP handler.
func HandleServerStreaming[Req any, Res any](fn ServerStreamingFunction[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		var request Req
		err := Receive(r.Body, &request)
		if err != nil {
			goto end
		}
		err = fn(r.Context(), &Request[Req]{Msg: &request}, &ServerStream[Res]{writer: w})
	end:
		WriteTrailer(w, err)
	}
}
