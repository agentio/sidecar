package sidecar

import (
	"io"
	"net/http"
	"sync"
)

// BidiStreamForClient holds state for a bidi-streaming RPC call.
type BidiStreamForClient[Req, Res any] struct {
	Trailer http.Header

	client *http.Client
	req    *http.Request
	resp   *http.Response
	reader io.ReadCloser
	writer io.WriteCloser
	wg     sync.WaitGroup
}

// CallBidiStream makes a bidi-streaming RPC call.
//
// The method argument should be the full path of the gRPC handler.
func CallBidiStream[Req, Res any](client *Client, method string) (*BidiStreamForClient[Req, Res], error) {
	url := client.Host + method
	pr, pw := io.Pipe()
	stream := &BidiStreamForClient[Req, Res]{
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
		// This will complete when the server sends its first reply.
		resp, err := stream.client.Do(stream.req)
		if err != nil {
			return
		}
		stream.reader = resp.Body
		stream.resp = resp
	})
	return stream, err
}

// Send sends a message to the bidi-streaming method.
func (b *BidiStreamForClient[Req, Res]) Send(msg *Req) error {
	return Send(b.writer, msg)
}

// CloseRequest closes the request-sending connection to the bidi-streaming method.
func (b *BidiStreamForClient[Req, Res]) CloseRequest() error {
	err := b.writer.Close() // Close the writer when done streaming
	b.wg.Wait()
	return err
}

// Receive reads a message from the bidi-streaming method.
func (b *BidiStreamForClient[Req, Res]) Receive() (*Res, error) {
	b.wg.Wait() // wait for reader to be set
	var response Res
	err := Receive(b.reader, &response)
	return &response, err
}

// CloseResponse closes the connection to the bidi-streaming method.
func (b *BidiStreamForClient[Req, Res]) CloseResponse() error {
	_, err := io.ReadAll(b.reader)
	if err != nil {
		return err
	}
	b.Trailer = b.resp.Trailer
	return ErrorForTrailer(b.Trailer)
}
