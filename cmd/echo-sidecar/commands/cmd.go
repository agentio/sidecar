// Package commands contains command implementations.
package commands

import (
	"github.com/agentio/sidecar/cmd/echo-sidecar/commands/call"
	"github.com/agentio/sidecar/cmd/echo-sidecar/commands/serve"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "echo-sidecar",
	}

	cmd.AddCommand(call.Cmd())
	cmd.AddCommand(serve.Cmd())
	return cmd
}
