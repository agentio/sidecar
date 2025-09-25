package sidecar

import (
	"errors"
	"io"
	"net/http"
)

// CallUnary makes a unary RPC call.
//
// The method argument should be the full path of the gRPC handler.
func CallUnary[Req, Res any](client *Client, method string, request *Request[Req]) (*Response[Res], error) {
	buf, err := serialize(request.Msg)
	if err != nil {
		return nil, err
	}
	url := client.Host + method
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header = client.Header.Clone()
	req.Header.Set("Content-Type", "application/grpc")
	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var response Res
	err = Receive(resp.Body, &response)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Response[Res]{
		Msg:     &response,
		Trailer: resp.Trailer,
	}, ErrorForTrailer(resp.Trailer)
}
