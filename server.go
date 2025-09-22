package sidecar

import "net/http"

// NewServer creates an http.Server instance that is configured for h2c communication.
//
// With appropriate handlers, this can be used to run gRPC services.
func NewServer(handler http.Handler) *http.Server {
	// Configure protocols for h2c-only support (HTTP/2 cleartext)
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true) // Enable h2c (HTTP/2 cleartext)
	protocols.SetHTTP1(false)           // Explicitly disable HTTP/1.1
	protocols.SetHTTP2(false)           // Explicitly disable encrypted HTTP/2 (HTTPS)
	return &http.Server{
		Handler:   handler,
		Protocols: protocols,
	}
}
