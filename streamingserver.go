package sidecar

import (
	"io"
	"net/http"
)

// ServerStreamForClient holds state for a server-streaming RPC call.
type ServerStreamForClient[Req, Res any] struct {
	Trailer http.Header

	client *http.Client
	req    *http.Request
	resp   *http.Response
	reader io.ReadCloser
}

// CallServerStream makes a server-streaming RPC call.
//
// The method argument should be the full path of the gRPC handler.
func CallServerStream[Req, Res any](client *Client, method string, request *Request[Req]) (*ServerStreamForClient[Req, Res], error) {
	buf, err := serialize(request.Msg)
	url := client.host + method
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/grpc")
	resp, err := client.httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	return &ServerStreamForClient[Req, Res]{
		reader: resp.Body,
		resp:   resp,
	}, err
}

// Receive reads a message from the server-streaming method.
func (b *ServerStreamForClient[Req, Res]) Receive() (*Res, error) {
	var response Res
	err := Receive(b.reader, &response)
	return &response, err
}

// CloseResponse closes the connection to the server-streaming method.
func (b *ServerStreamForClient[Req, Res]) CloseResponse() error {
	_, err := io.ReadAll(b.reader)
	b.Trailer = b.resp.Trailer
	return err
}
