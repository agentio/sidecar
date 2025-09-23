package sidecar

import (
	"io"
	"net/http"
)

// ClientStream provides messaging to client streaming handlers.
type ClientStream[Req any] struct {
	reader io.ReadCloser
}

// Receive reads a request message from a client stream.
func (b *ClientStream[Req]) Receive() (*Req, error) {
	var request Req
	err := Receive(b.reader, &request)
	return &request, err
}

// Client streaming handlers should be functions that implement this interface.
type ClientStreamingFunction[Req, Res any] func(stream *ClientStream[Req]) (*Response[Res], error)

// HandleClientStreaming wraps a client streaming function in an HTTP handler.
func HandleClientStreaming[Req any, Res any](fn ClientStreamingFunction[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		response, err := fn(&ClientStream[Req]{reader: r.Body})
		if err != nil {
			goto end
		}
		err = Send(w, response.Msg)
	end:
		WriteTrailer(w, err)
	}
}
