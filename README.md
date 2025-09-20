# Sidecar RPC

Sidecar is a library for creating gRPC clients and servers in Go.

Sidecar arose from the realization that popular gRPC libraries like [grpc-go](https://github.com/grpc/grpc-go) and [connect-go](https://github.com/connectrpc/connect-go) are massively overloaded with features that aren't needed by gRPC clients and servers that run with sidecar proxies. In these environments, the sidecars themselves include all of the powerful networking features that applications need, and building these into applications adds needless complexity, bloat, and supply-chain risk.

Sidecar intentionally excludes capabilities that are better-handled by proxies, including but not limited to:
- Compression
- Transcoding
- Name Resolution
- Load Balancing
- Interceptors
- Retry
- Health Checking
- Observability

If you prefer to build these capabilities into your application, please use another gRPC library. However, if you are interested in creating lightweight gRPC services that delegate advanced networking to sidecar proxies, you are welcome to use Sidecar either directly or by copying its code into your application.

## No Generated Code

One of the tenets of Sidecar is to avoid generated code in client and server implementations. Instead, we rely on Go generics and slightly increased inline complexity in Sidecar servers and clients. For more detail, see the `echo-sidecar` example.

## Example

A complete example is in [cmd/echo-sidecar](/cmd/echo-sidecar), a command-line tool that uses Sidecar to build and call a gRPC server that implements a simple echo service. All four gRPC streaming modes are supported.

## License

Sidecar is released under the [Apache 2 license](/LICENSE).
