package sidecar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestEchoBytes(t *testing.T) {
	server, err := RunEchoServerBytes(9999, "")
	if err != nil {
		t.Fatalf("%s", err)
	}
	client := NewClient(ClientOptions{Address: "localhost:9999"})
	testGetBytes(t, client)
	testExpandBytes(t, client)
	testCollectBytes(t, client)
	testUpdateBytes(t, client)
	if err := server.Shutdown(t.Context()); err != nil {
		t.Fatalf("%s", err)
	}
}

func RunEchoServerBytes(port int, socket string) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc(EchoGetProcedure, HandleUnary(
		func(ctx context.Context, req *Request[[]byte]) (*Response[[]byte], error) {
			response := []byte("Go echo get: " + string(*req.Msg))
			return NewResponse(&response), nil
		}))
	mux.HandleFunc(EchoExpandProcedure, HandleServerStreaming(
		func(ctx context.Context, req *Request[[]byte], stream *ServerStream[[]byte]) error {
			parts := strings.Split(string(*req.Msg), " ")
			for _, part := range parts {
				response := []byte("Go echo expand: " + part)
				if err := stream.Send(&response); err != nil {
					return err
				}
			}
			return nil
		}))
	mux.HandleFunc(EchoCollectProcedure, HandleClientStreaming(
		func(ctx context.Context, stream *ClientStream[[]byte]) (*Response[[]byte], error) {
			parts := []string{}
			for {
				request, err := stream.Receive()
				if errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					return nil, err
				}
				parts = append(parts, string(*request))
			}
			response := []byte("Go echo collect: " + strings.Join(parts, " "))
			return NewResponse(&response), nil
		}))
	mux.HandleFunc(EchoUpdateProcedure, HandleBidiStreaming(
		func(ctx context.Context, stream *BidiStream[[]byte, []byte]) error {
			for {
				request, err := stream.Receive()
				if errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					return err
				}
				response := []byte("Go echo update: " + string(*request))
				err = stream.Send(&response)
				if err != nil {
					return err
				}
			}
			return nil
		}))
	server := NewServer(mux)
	var err error
	var listener net.Listener
	if port == 0 {
		listener, err = net.Listen("unix", socket)
	} else {
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	}
	if err != nil {
		return nil, err
	}
	go func() { _ = server.Serve(listener) }()
	return server, nil
}

func testGetBytes(t *testing.T, client *Client) {
	request := []byte("hello")
	response, err := CallUnary[[]byte, []byte](
		t.Context(),
		client,
		EchoGetProcedure,
		NewRequest(&request),
	)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	if string(*response.Msg) != "Go echo get: hello" {
		t.Errorf("Invalid get response")
	}
}

func testExpandBytes(t *testing.T, client *Client) {
	request := []byte("hello hello hello")
	stream, err := CallServerStream[[]byte, []byte](
		t.Context(),
		client,
		EchoExpandProcedure,
		NewRequest(&request),
	)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	count := 0
	for {
		msg, err := stream.Receive()
		if err != nil {
			break
		}
		if string(*msg) != "Go echo expand: hello" {
			t.Errorf("Invalid expand response")
		}
		count++
	}
	if count != 3 {
		t.Errorf("Incorrect number of expand responses (%d)", count)
	}
}

func testCollectBytes(t *testing.T, client *Client) {
	stream, err := CallClientStream[[]byte, []byte](
		t.Context(),
		client,
		EchoCollectProcedure,
	)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	request := []byte("hello")
	for range 3 {
		if err := stream.Send(&request); err != nil {
			t.Fatalf("error %s", err)
		}
	}
	msg, err := stream.CloseAndReceive()
	if err != nil {
		t.Fatalf("error %s", err)
	}
	if string(*msg) != "Go echo collect: hello hello hello" {
		t.Errorf("Invalid collect response: %s", string(*msg))
	}
}

func testUpdateBytes(t *testing.T, client *Client) {
	stream, err := CallBidiStream[[]byte, []byte](
		t.Context(),
		client,
		EchoUpdateProcedure,
	)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	request := []byte("hello")
	for range 6 {
		if err := stream.Send(&request); err != nil {
			t.Fatalf("error %s", err)
		}
		msg, err := stream.Receive()
		if err != nil {
			t.Fatalf("error %s", err)
		}
		if string(*msg) != "Go echo update: hello" {
			t.Errorf("Invalid update response: %s", string(*msg))
		}
	}
	if err = stream.CloseRequest(); err != nil {
		t.Fatalf("error %s", err)
	}
}
