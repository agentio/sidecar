// Package get implements calls to the get method.
package get

import (
	"fmt"
	"time"

	"github.com/agentio/sidecar"
	"github.com/agentio/sidecar/cmd/echo-sidecar/constants"
	"github.com/agentio/sidecar/cmd/echo-sidecar/genproto/echopb"
	"github.com/agentio/sidecar/cmd/echo-sidecar/track"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func Cmd() *cobra.Command {
	var message string
	var address string
	var n int
	var verbose bool
	cmd := &cobra.Command{
		Use:  "get",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := sidecar.NewClient(address)
			defer track.Measure(time.Now(), "get", n, cmd.OutOrStdout())
			for j := 0; j < n; j++ {
				response, err := sidecar.CallUnary[echopb.EchoRequest, echopb.EchoResponse](
					cmd.Context(),
					client,
					constants.EchoGetProcedure,
					sidecar.NewRequest(&echopb.EchoRequest{Text: message}),
				)
				if err != nil {
					return err
				}
				if n == 1 {
					body, err := protojson.Marshal(response.Msg)
					if err != nil {
						return err
					}
					_, _ = cmd.OutOrStdout().Write(body)
					_, _ = cmd.OutOrStdout().Write([]byte("\n"))
					if verbose {
						fmt.Println("Response Trailers:")
						for key, values := range response.Trailer {
							fmt.Printf("  %s: %v\n", key, values)
						}
					}
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
