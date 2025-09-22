package sidecar

import (
	"io"
	"net/http"
)

// UnaryResponse describes a response from a unary gRPC method.
type UnaryResponse[T any] struct {
	Msg     *T
	Trailer http.Header
}

// CallUnary makes a unary RPC call.
//
// The method argument should be the full path of the gRPC handler.
func CallUnary[Req, Res any](client *Client, method string, request *Req) (*UnaryResponse[Res], error) {
	buf, err := serialize(request)
	if err != nil {
		return nil, err
	}
	url := client.host + method
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/grpc")
	client.addHeaders(req)
	resp, err := client.httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var response Res
	err = Receive(resp.Body, &response)
	if err != nil {
		return nil, err
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &UnaryResponse[Res]{
		Msg:     &response,
		Trailer: resp.Trailer,
	}, nil
}
