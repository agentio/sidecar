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

	"google.golang.org/protobuf/types/known/anypb"
)

func TestEchoAnyPb(t *testing.T) {
	server, err := RunEchoServerAnyPb(9998, "")
	if err != nil {
		t.Fatalf("%s", err)
	}
	client := NewClient(ClientOptions{Address: "localhost:9998"})
	testGetAnyPb(t, client)
	testExpandAnyPb(t, client)
	testCollectAnyPb(t, client)
	testUpdateAnyPb(t, client)
	if err := server.Shutdown(t.Context()); err != nil {
		t.Fatalf("%s", err)
	}
}

func RunEchoServerAnyPb(port int, socket string) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc(EchoGetProcedure, HandleUnary(
		func(ctx context.Context, req *Request[anypb.Any]) (*Response[anypb.Any], error) {
			response := anypb.Any{Value: []byte("Go echo get: " + string(req.Msg.Value))}
			return NewResponse(&response), nil
		}))
	mux.HandleFunc(EchoExpandProcedure, HandleServerStreaming(
		func(ctx context.Context, req *Request[anypb.Any], stream *ServerStream[anypb.Any]) error {
			parts := strings.Split(string(req.Msg.Value), " ")
			for _, part := range parts {
				response := anypb.Any{Value: []byte("Go echo expand: " + part)}
				if err := stream.Send(&response); err != nil {
					return err
				}
			}
			return nil
		}))
	mux.HandleFunc(EchoCollectProcedure, HandleClientStreaming(
		func(ctx context.Context, stream *ClientStream[anypb.Any]) (*Response[anypb.Any], error) {
			parts := []string{}
			for {
				request, err := stream.Receive()
				if errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					return nil, err
				}
				parts = append(parts, string(request.Value))
			}
			response := anypb.Any{Value: []byte("Go echo collect: " + strings.Join(parts, " "))}
			return NewResponse(&response), nil
		}))
	mux.HandleFunc(EchoUpdateProcedure, HandleBidiStreaming(
		func(ctx context.Context, stream *BidiStream[anypb.Any, anypb.Any]) error {
			for {
				request, err := stream.Receive()
				if errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					return err
				}
				response := anypb.Any{Value: []byte("Go echo update: " + string(request.Value))}
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
	go server.Serve(listener)
	return server, nil
}

func testGetAnyPb(t *testing.T, client *Client) {
	request := anypb.Any{Value: []byte("hello")}
	response, err := CallUnary[anypb.Any, anypb.Any](
		t.Context(),
		client,
		EchoGetProcedure,
		NewRequest(&request),
	)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	if string(response.Msg.Value) != "Go echo get: hello" {
		t.Errorf("Invalid get response")
	}
}

func testExpandAnyPb(t *testing.T, client *Client) {
	request := anypb.Any{Value: []byte("hello hello hello")}
	stream, err := CallServerStream[anypb.Any, anypb.Any](
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
		if string(msg.Value) != "Go echo expand: hello" {
			t.Errorf("Invalid expand response")
		}
		count++
	}
	if count != 3 {
		t.Errorf("Incorrect number of expand responses (%d)", count)
	}
}

func testCollectAnyPb(t *testing.T, client *Client) {
	stream, err := CallClientStream[anypb.Any, anypb.Any](
		t.Context(),
		client,
		EchoCollectProcedure,
	)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	request := anypb.Any{Value: []byte("hello")}
	for range 3 {
		if err := stream.Send(&request); err != nil {
			t.Fatalf("error %s", err)
		}
	}
	msg, err := stream.CloseAndReceive()
	if string(msg.Value) != "Go echo collect: hello hello hello" {
		t.Errorf("Invalid collect response: %s", string(msg.Value))
	}
}

func testUpdateAnyPb(t *testing.T, client *Client) {
	stream, err := CallBidiStream[anypb.Any, anypb.Any](
		t.Context(),
		client,
		EchoUpdateProcedure,
	)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	request := anypb.Any{Value: []byte("hello")}
	for range 6 {
		if err := stream.Send(&request); err != nil {
			t.Fatalf("error %s", err)
		}
		msg, err := stream.Receive()
		if err != nil {
			t.Fatalf("error %s", err)
		}
		if string(msg.Value) != "Go echo update: hello" {
			t.Errorf("Invalid update response: %s", string(msg.Value))
		}
	}
	if err = stream.CloseRequest(); err != nil {
		t.Fatalf("error %s", err)
	}
}
