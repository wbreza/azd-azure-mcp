package cmd

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

func newServerCommand() *cobra.Command {
	serverGroup := &cobra.Command{
		Use: "server",
	}

	startCmd := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := server.NewMCPServer(
				"Azure Account",
				"1.0.0",
				server.WithToolCapabilities(false),
				server.WithRecovery(),
				server.WithLogging(),
				server.WithInstructions("Supports tools for interacting with Azure accounts, subscriptions and locations."),
			)

			// Subscription tools
			listSubscriptionsTool := mcp.NewTool(
				"list-subscriptions",
				mcp.WithDescription("Lists all Azure subscriptions accessible to the account"),
			)
			s.AddTool(listSubscriptionsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				azCmd := exec.Command("az", "account", "list")
				return runAzCommandWithResult(azCmd), nil
			})

			// Location tools
			listLocationsTool := mcp.NewTool(
				"list-locations",
				mcp.WithDescription("Lists all Azure locations available for the current account"),
			)
			s.AddTool(listLocationsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				azCmd := exec.Command("az", "account", "list-locations")
				return runAzCommandWithResult(azCmd), nil
			})

			setDefaultSubscriptionTool := mcp.NewTool(
				"set-default-subscription",
				mcp.WithDescription("Sets the specified Azure subscription as the default for subsequent Azure CLI operations"),
				mcp.WithString("subscriptionId",
					mcp.Required(),
					mcp.Description("The Azure subscription ID to set as default"),
				),
			)
			s.AddTool(setDefaultSubscriptionTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				subIdArg, ok := request.GetArguments()["subscriptionId"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: subscriptionId"), nil
				}
				subId, ok := subIdArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: subscriptionId, expected string"), nil
				}
				azCmd := exec.Command("az", "account", "set", "--subscription", subId)
				return runAzCommandWithResult(azCmd), nil
			})

			showAccountTool := mcp.NewTool(
				"show-account",
				mcp.WithDescription("Shows details of the current Azure subscription/account."),
			)
			s.AddTool(showAccountTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				azCmd := exec.Command("az", "account", "show")
				return runAzCommandWithResult(azCmd), nil
			})

			showUserTool := mcp.NewTool(
				"show-user",
				mcp.WithDescription("Shows information for the current logged in Azure AD user."),
			)
			s.AddTool(showUserTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				azCmd := exec.Command("az", "ad", "signed-in-user", "show")
				return runAzCommandWithResult(azCmd), nil
			})

			// Start the server
			if err := server.ServeStdio(s); err != nil {
				fmt.Printf("Server error: %v\n", err)
			}

			return nil
		},
	}

	serverGroup.AddCommand(startCmd)

	return serverGroup
}

func runAzCommandWithResult(cmd *exec.Cmd) *mcp.CallToolResult {
	output, err := cmd.CombinedOutput()
	result := string(output)
	if err != nil {
		result = result + "\n[error] " + err.Error()
	}
	return mcp.NewToolResultText(result)
}
