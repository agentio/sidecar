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
			mux.HandleFunc(service.EchoGetProcedure, handleGet)
			mux.HandleFunc(service.EchoExpandProcedure, handleExpand)
			mux.HandleFunc(service.EchoCollectProcedure, handleCollect)
			mux.HandleFunc(service.EchoUpdateProcedure, handleUpdate)
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

func handleGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/grpc")
	var request echopb.EchoRequest
	err := sidecar.Receive(r.Body, &request)
	if err != nil {
		return
	}
	err = sidecar.Send(w, &echopb.EchoResponse{Text: "Go echo get: " + request.Text})
	if err != nil {
		return
	}
	w.Header().Set("Trailer:Grpc-Status", "0")
}

func handleExpand(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/grpc")
	var request echopb.EchoRequest
	err := sidecar.Receive(r.Body, &request)
	if err != nil {
		return
	}
	parts := strings.Split(request.Text, " ")
	for _, part := range parts {
		err := sidecar.Send(w, &echopb.EchoResponse{Text: "Go echo expand: " + part})
		if err != nil {
			return
		}
	}
	w.Header().Set("Trailer:Grpc-Status", "0")
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/grpc")
	for {
		var request echopb.EchoRequest
		err := sidecar.Receive(r.Body, &request)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return
		}
		err = sidecar.Send(w, &echopb.EchoResponse{Text: "Go echo update: " + request.Text})
		if err != nil {
			return
		}
	}
	w.Header().Set("Trailer:Grpc-Status", "0")
}

func handleCollect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/grpc")
	parts := []string{}
	for {
		var request echopb.EchoRequest
		err := sidecar.Receive(r.Body, &request)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return
		}
		parts = append(parts, request.Text)
	}
	err := sidecar.Send(w, &echopb.EchoResponse{Text: "Go echo collect: " + strings.Join(parts, " ")})
	if err != nil {
		return
	}
	w.Header().Set("Trailer:Grpc-Status", "0")
}
