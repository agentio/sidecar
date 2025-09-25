package sidecar

import (
	"errors"
	"io"
	"net/http"
	"sync"
)

// ClientStreamForClient holds state for a client-streaming RPC call.
type ClientStreamForClient[Req, Res any] struct {
	Trailer http.Header

	client *http.Client
	req    *http.Request
	resp   *http.Response
	reader io.ReadCloser
	writer io.WriteCloser
	wg     sync.WaitGroup
}

// CallClientStream makes a client-streaming RPC call.
//
// The method argument should be the full path of the gRPC handler.
func CallClientStream[Req, Res any](client *Client, method string) (*ClientStreamForClient[Req, Res], error) {
	url := client.Host + method
	pr, pw := io.Pipe()
	stream := &ClientStreamForClient[Req, Res]{
		writer: pw,
	}
	stream.client = client.HttpClient
	var err error
	stream.req, err = http.NewRequest(http.MethodPost, url, io.NopCloser(pr))
	if err != nil {
		return nil, err
	}
	stream.req.Header = client.Header.Clone()
	stream.wg.Go(func() {
		// This will complete when the client closes and the server reply is sent.
		resp, err := stream.client.Do(stream.req)
		if err != nil {
			return
		}
		stream.reader = resp.Body
		stream.resp = resp
	})
	return stream, err
}

// Send sends a message to the client-streaming method.
func (b *ClientStreamForClient[Req, Res]) Send(msg *Req) error {
	err := Send(b.writer, msg)
	if err != nil {
		return err
	}
	return err
}

// CloseAndReceive closes the client connection and reads the response from the client-streaming method.
func (b *ClientStreamForClient[Req, Res]) CloseAndReceive() (*Res, error) {
	err := b.writer.Close()
	if err != nil {
		return nil, err
	}
	b.wg.Wait()
	var response Res
	err = Receive(b.reader, &response)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	_, err = io.ReadAll(b.reader)
	if err != nil {
		return nil, err
	}
	b.Trailer = b.resp.Trailer
	return &response, ErrorForTrailer(b.Trailer)
}
