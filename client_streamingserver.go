package sidecar

import (
	"context"
	"io"
	"net/http"
)

// ServerStreamForClient holds state for a server-streaming RPC call.
type ServerStreamForClient[Req, Res any] struct {
	Trailer http.Header

	resp   *http.Response
	reader io.ReadCloser
}

// CallServerStream makes a server-streaming RPC call.
//
// The method argument should be the full path of the gRPC handler.
func CallServerStream[Req, Res any](ctx context.Context, client *Client, method string, request *Request[Req]) (*ServerStreamForClient[Req, Res], error) {
	buf, err := serialize(request.Msg)
	if err != nil {
		return nil, err
	}
	url := client.Host + method
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header = client.Header.Clone()
	resp, err := client.HttpClient.Do(req)
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
	if err != nil {
		return err
	}
	b.Trailer = b.resp.Trailer
	return ErrorForTrailer(b.Trailer)
}
