package sidecar

import "net/http"

// Response describes a response from a unary gRPC method.
type Response[T any] struct {
	Msg     *T
	Trailer http.Header
}

// NewResponse creates a response from a message.
func NewResponse[T any](msg *T) *Response[T] {
	return &Response[T]{Msg: msg}
}
