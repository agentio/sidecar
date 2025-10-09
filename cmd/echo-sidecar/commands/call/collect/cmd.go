// Package collect implements calls to the collect method.
package collect

import (
	"fmt"
	"log"

	"github.com/agentio/sidecar"
	"github.com/agentio/sidecar/cmd/echo-sidecar/constants"
	"github.com/agentio/sidecar/cmd/echo-sidecar/genproto/echopb"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func Cmd() *cobra.Command {
	var message string
	var address string
	var n int
	var verbose bool
	cmd := &cobra.Command{
		Use:  "collect",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := sidecar.NewClient(address)
			stream, err := sidecar.CallClientStream[echopb.EchoRequest, echopb.EchoResponse](
				cmd.Context(),
				client,
				constants.EchoCollectProcedure,
			)
			if err != nil {
				return err
			}
			for range 3 {
				err = stream.Send(&echopb.EchoRequest{Text: message})
				if err != nil {
					log.Printf("Error writing to pipe: %v", err)
					return err
				}
			}
			response, err := stream.CloseAndReceive()
			if err != nil {
				return err
			}
			body, err := protojson.Marshal(response)
			if err != nil {
				return err
			}
			_, _ = cmd.OutOrStdout().Write(body)
			_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			if verbose {
				fmt.Println("Response Trailers:")
				for key, values := range stream.Trailer {
					fmt.Printf("  %s: %v\n", key, values)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&message, "message", "m", "hello", "message")
	cmd.Flags().StringVarP(&address, "address", "a", "unix:@echo", "address of the echo server to use")
	cmd.Flags().IntVarP(&n, "number", "n", 1, "number of times to call the method")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose")
	return cmd
}
