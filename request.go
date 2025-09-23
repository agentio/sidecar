package sidecar

import "net/http"

// Request describes a request to a unary gRPC method.
type Request[T any] struct {
	Msg     *T
	Trailer http.Header
}
