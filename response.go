package sidecar

import "net/http"

// Response describes a response from a unary gRPC method.
type Response[T any] struct {
	Msg     *T
	Trailer http.Header
}
