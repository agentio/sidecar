package sidecar

import "net/http"

// Request describes a request to a unary gRPC method.
type Request[T any] struct {
	Msg     *T
	Trailer http.Header
}

// NewRequest creates a request from a message.
func NewRequest[T any](msg *T) *Request[T] {
	return &Request[T]{Msg: msg}
}
