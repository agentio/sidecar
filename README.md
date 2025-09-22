# Sidecar RPC

Sidecar is a library for creating gRPC clients and servers in Go.

Sidecar arose from the realization that popular gRPC libraries like [grpc-go](https://github.com/grpc/grpc-go) and [connect-go](https://github.com/connectrpc/connect-go) are loaded with capabilities that aren't needed by gRPC applications that use sidecar proxies. When applications use sidecars, the sidecars provide these capabilities along with assurance that they are implemented and configured correctly. Redundantly including them in networking support libraries adds needless complexity, bloat, and supply-chain risk.

Some of the capabilities that Sidecar intentionally omits include:
- Compression
- Transcoding
- Name Resolution
- Load Balancing
- Interceptors
- Retry
- Health Checking
- Observability

If you need these capabilities built into your application, then another gRPC library is probably a better fit. But if you are building gRPC services that delegate advanced networking to sidecar proxies, Sidecar can help you make your services lean and maintainable.

## Built on the Go standard library

With Sidecar, gRPC applications build directly on the HTTP2 support in the Go standard library. Beginning with Go 1.25, this includes [SetUnencryptedHTTP2](https://pkg.go.dev/net/http#Protocols.SetUnencryptedHTTP2), which clients and servers can use to create unencrypted HTTP2 connections (also called "HTTP/2 cleartext" or h2c). It also includes connection sharing and reuse, which allows HTTP2 connections to be automatically reused, and all configuration options in the Go standard library are directly available to Sidecar-based applications.

## No Generated Code

Apart from protocol buffer serialization, Sidecar does not use code generation. Instead, Go generics are used to call gRPC methods with appropriate types. This slightly increases inline complexity, but adds development and build-time simplicity. For example, here is a call to a unary gRPC method:
```go
// Create a reusable client.
client := sidecar.NewClient(address)
// Use the client to make a unary rpc call.
response, err := sidecar.CallUnary[echopb.EchoRequest, echopb.EchoResponse](
	client,
	"/echo.v1.Echo/Get",
	&echopb.EchoRequest{Text: message},
)
```

## Bring Your Own Serialization

Along with protobuf-encoded messages, Sidecar allows messages to be sent and received as raw bytes. Here's an example with protobuf encoding (replace that with your own favorite encoding):
```go
// Create a reusable client.
client := sidecar.NewClient(address)
// Marshal a request message.
b, _ := proto.Marshal(&echopb.EchoRequest{Text: message})
// Use the client to make a unary rpc call.
response, err := sidecar.CallUnary[[]byte, []byte](
	client,
	"/echo.v1.Echo/Get",
	&b,
)
// Unmarshal the response
var message echopb.EchoResponse
err = proto.Unmarshal(*(response.Msg), &message)
```

## Example

This repo includes [echo-sidecar](/cmd/echo-sidecar), a command-line tool that uses Sidecar to build and call a gRPC server that implements a simple echo service. All four gRPC streaming modes are supported.

## License

Sidecar is released under the [Apache 2 license](/LICENSE).
