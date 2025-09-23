// Package serve implements an Echo server.
package serve

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/agentio/sidecar"
	"github.com/agentio/sidecar/cmd/echo-sidecar/genproto/echopb"
	"github.com/agentio/sidecar/cmd/echo-sidecar/service"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var port int
	var socket string
	var verbose bool
	cmd := &cobra.Command{
		Use:  "serve",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mux := http.NewServeMux()
			mux.HandleFunc(service.EchoGetProcedure, UnaryHandler(get))
			mux.HandleFunc(service.EchoExpandProcedure, ServerStreamingHandler(expand))
			mux.HandleFunc(service.EchoCollectProcedure, ClientStreamingHandler(collect))
			mux.HandleFunc(service.EchoUpdateProcedure, BidiStreamingHandler(update))
			server := sidecar.NewServer(mux)
			var err error
			var listener net.Listener
			if port == 0 {
				listener, err = net.Listen("unix", socket)
			} else {
				listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
			}
			if err != nil {
				return err
			}
			return server.Serve(listener)
		},
	}
	cmd.Flags().IntVarP(&port, "port", "p", 0, "server port")
	cmd.Flags().StringVarP(&socket, "socket", "s", "@echo", "server socket")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose")
	return cmd
}

/////////////////////////////////////////////////////////

// Request describes a request to a unary gRPC method.
type Request[T any] struct {
	Msg     *T
	Trailer http.Header
}

// Response describes a response from a unary gRPC method.
type Response[T any] struct {
	Msg     *T
	Trailer http.Header
}

type Unary[Req, Res any] func(request *Request[Req]) (*Response[Res], error)

func UnaryHandler[Req any, Res any](fn Unary[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		var request Req
		err := sidecar.Receive(r.Body, &request)
		if err != nil {
			return
		}
		response, err := fn(&Request[Req]{Msg: &request})
		if err != nil {
			w.Header().Set("Trailer:Grpc-Status", "11")
			return
		}
		err = sidecar.Send(w, response.Msg)
		if err != nil {
			return
		}
		w.Header().Set("Trailer:Grpc-Status", "0")
	}
}

func get(req *Request[echopb.EchoRequest]) (*Response[echopb.EchoResponse], error) {
	return &Response[echopb.EchoResponse]{
		Msg: &echopb.EchoResponse{
			Text: "Go echo get: " + req.Msg.Text,
		},
	}, nil
}

/////////////////////////////////////////////////////////

type ServerStreamForServer[Req, Res any] struct {
	writer http.ResponseWriter
}

func (b *ServerStreamForServer[Req, Res]) Send(msg *Res) error {
	return sidecar.Send(b.writer, msg)
}

type ServerStreaming[Req, Res any] func(request *Request[Req], stream *ServerStreamForServer[Req, Res]) error

func ServerStreamingHandler[Req any, Res any](fn ServerStreaming[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		var request Req
		err := sidecar.Receive(r.Body, &request)
		if err != nil {
			return
		}
		err = fn(&Request[Req]{Msg: &request}, &ServerStreamForServer[Req, Res]{writer: w})
		if err != nil {
			w.Header().Set("Trailer:Grpc-Status", "11")
			return
		}
		w.Header().Set("Trailer:Grpc-Status", "0")
	}
}

func expand(req *Request[echopb.EchoRequest], stream *ServerStreamForServer[echopb.EchoRequest, echopb.EchoResponse]) error {
	parts := strings.Split(req.Msg.Text, " ")
	for _, part := range parts {
		err := stream.Send(&echopb.EchoResponse{Text: "Go echo expand: " + part})
		if err != nil {
			return err
		}
	}
	return nil
}

/////////////////////////////////////////////////////////

type BidiStreamForServer[Req, Res any] struct {
	reader io.ReadCloser
	writer http.ResponseWriter
}

func (b *BidiStreamForServer[Req, Res]) Send(msg *Res) error {
	return sidecar.Send(b.writer, msg)
}

func (b *BidiStreamForServer[Req, Res]) Receive() (*Res, error) {
	var response Res
	err := sidecar.Receive(b.reader, &response)
	return &response, err
}

func update(stream *BidiStreamForServer[echopb.EchoRequest, echopb.EchoResponse]) error {
	for {
		request, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		err = stream.Send(&echopb.EchoResponse{Text: "Go echo update: " + request.Text})
		if err != nil {
			return err
		}
	}
	return nil
}

type BidiStreaming[Req, Res any] func(stream *BidiStreamForServer[Req, Res]) error

func BidiStreamingHandler[Req any, Res any](fn BidiStreaming[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		err := fn(&BidiStreamForServer[Req, Res]{reader: r.Body, writer: w})
		if err != nil {
			w.Header().Set("Trailer:Grpc-Status", "11")
			return
		}
		w.Header().Set("Trailer:Grpc-Status", "0")
	}
}

/////////////////////////////////////////////////////////

type ClientStreamForServer[Req, Res any] struct {
	reader io.ReadCloser
}

func (b *ClientStreamForServer[Req, Res]) Receive() (*Res, error) {
	var response Res
	err := sidecar.Receive(b.reader, &response)
	return &response, err
}

type ClientStreaming[Req, Res any] func(stream *ClientStreamForServer[Req, Res]) (*Response[Res], error)

func ClientStreamingHandler[Req any, Res any](fn ClientStreaming[Req, Res]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/grpc")
		response, err := fn(&ClientStreamForServer[Req, Res]{reader: r.Body})
		if err != nil {
			return
		}
		err = sidecar.Send(w, response.Msg)
		if err != nil {
			return
		}
		w.Header().Set("Trailer:Grpc-Status", "0")
	}
}

func collect(stream *ClientStreamForServer[echopb.EchoRequest, echopb.EchoResponse]) (*Response[echopb.EchoResponse], error) {
	parts := []string{}
	for {
		request, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		parts = append(parts, request.Text)
	}
	return &Response[echopb.EchoResponse]{Msg: &echopb.EchoResponse{Text: "Go echo collect: " + strings.Join(parts, " ")}}, nil
}
