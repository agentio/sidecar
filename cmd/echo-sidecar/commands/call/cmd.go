// Package call implements calls to the Echo service.
package call

import (
	"github.com/agentio/sidecar/cmd/echo-sidecar/commands/call/collect"
	"github.com/agentio/sidecar/cmd/echo-sidecar/commands/call/expand"
	"github.com/agentio/sidecar/cmd/echo-sidecar/commands/call/get"
	"github.com/agentio/sidecar/cmd/echo-sidecar/commands/call/update"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "call",
	}
	cmd.AddCommand(get.Cmd())
	cmd.AddCommand(expand.Cmd())
	cmd.AddCommand(collect.Cmd())
	cmd.AddCommand(update.Cmd())
	return cmd
}
