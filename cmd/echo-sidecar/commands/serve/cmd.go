// Package serve implements an Echo server.
package serve

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/agentio/sidecar"
	"github.com/agentio/sidecar/cmd/echo-sidecar/constants"
	"github.com/agentio/sidecar/cmd/echo-sidecar/genproto/echopb"
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
			mux.HandleFunc(constants.EchoGetProcedure, sidecar.HandleUnary(get))
			mux.HandleFunc(constants.EchoExpandProcedure, sidecar.HandleServerStreaming(expand))
			mux.HandleFunc(constants.EchoCollectProcedure, sidecar.HandleClientStreaming(collect))
			mux.HandleFunc(constants.EchoUpdateProcedure, sidecar.HandleBidiStreaming(update))
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

func get(ctx context.Context, req *sidecar.Request[echopb.EchoRequest]) (*sidecar.Response[echopb.EchoResponse], error) {
	return sidecar.NewResponse(&echopb.EchoResponse{
		Text: "Go echo get: " + req.Msg.Text,
	}), nil
}

func expand(ctx context.Context, req *sidecar.Request[echopb.EchoRequest], stream *sidecar.ServerStream[echopb.EchoResponse]) error {
	parts := strings.Split(req.Msg.Text, " ")
	for _, part := range parts {
		if err := stream.Send(&echopb.EchoResponse{Text: "Go echo expand: " + part}); err != nil {
			return err
		}
	}
	return nil
}

func collect(ctx context.Context, stream *sidecar.ClientStream[echopb.EchoRequest]) (*sidecar.Response[echopb.EchoResponse], error) {
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
	return sidecar.NewResponse(&echopb.EchoResponse{
		Text: "Go echo collect: " + strings.Join(parts, " "),
	}), nil
}

func update(ctx context.Context, stream *sidecar.BidiStream[echopb.EchoRequest, echopb.EchoResponse]) error {
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
