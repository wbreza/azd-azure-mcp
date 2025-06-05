package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of the Azure Service Bus MCP Server",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Azure Service Bus MCP Server v0.0.1")
		},
	}
}
