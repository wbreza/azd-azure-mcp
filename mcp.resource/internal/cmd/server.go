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
				"Azure Resources",
				"1.0.0",
				server.WithToolCapabilities(false),
				server.WithRecovery(),
				server.WithLogging(),
				server.WithInstructions("Supports tools for interacting with Azure subscriptions, resource groups and generic resources."),
			)

			// Subscription tools
			listSubscriptionsTool := mcp.NewTool(
				"list-subscriptions",
				mcp.WithDescription("Lists all Azure subscriptions accessible to the account"),
			)
			s.AddTool(listSubscriptionsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				azCmd := exec.Command("az", "account", "list")
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			// Resource group tools
			listResourceGroupsTool := mcp.NewTool(
				"list-resource-groups",
				mcp.WithDescription("Lists all Azure resource groups in a subscription"),
				mcp.WithString("subscriptionId",
					mcp.Required(),
					mcp.Description("The Azure subscription ID to list resource groups for"),
				),
			)
			s.AddTool(listResourceGroupsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				arg, ok := request.GetArguments()["subscriptionId"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: subscriptionId"), nil
				}
				subId, ok := arg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: subscriptionId, expected string"), nil
				}
				azCmd := exec.Command("az", "group", "list", "--subscription", subId)
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			createResourceGroupTool := mcp.NewTool(
				"create-resource-group",
				mcp.WithDescription("Creates a new Azure resource group in a subscription"),
				mcp.WithString("resourceGroupName",
					mcp.Required(),
					mcp.Description("The name of the resource group to create"),
				),
				mcp.WithString("location",
					mcp.Required(),
					mcp.Description("The Azure location for the resource group"),
				),
				mcp.WithString("subscriptionId",
					mcp.Required(),
					mcp.Description("The Azure subscription ID to create the resource group in"),
				),
			)
			s.AddTool(createResourceGroupTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["resourceGroupName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroupName"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroupName, expected string"), nil
				}
				locationArg, ok := request.GetArguments()["location"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: location"), nil
				}
				location, ok := locationArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: location, expected string"), nil
				}
				subIdArg, ok := request.GetArguments()["subscriptionId"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: subscriptionId"), nil
				}
				subId, ok := subIdArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: subscriptionId, expected string"), nil
				}
				azCmd := exec.Command("az", "group", "create", "--name", name, "--location", location, "--subscription", subId)
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			showResourceGroupTool := mcp.NewTool(
				"show-resource-group",
				mcp.WithDescription("Shows details of a specific Azure resource group"),
				mcp.WithString("resourceGroupName",
					mcp.Required(),
					mcp.Description("The name of the resource group to show"),
				),
				mcp.WithString("subscriptionId",
					mcp.Required(),
					mcp.Description("The subscription ID containing the resource group"),
				),
			)
			s.AddTool(showResourceGroupTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["resourceGroupName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroupName"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroupName, expected string"), nil
				}
				subIdArg, ok := request.GetArguments()["subscriptionId"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: subscriptionId"), nil
				}
				subId, ok := subIdArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: subscriptionId, expected string"), nil
				}
				azCmd := exec.Command("az", "group", "show", "--name", name, "--subscription", subId)
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			// Location tools
			listLocationsTool := mcp.NewTool(
				"list-locations",
				mcp.WithDescription("Lists all Azure locations available for the current account"),
			)
			s.AddTool(listLocationsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				azCmd := exec.Command("az", "account", "list-locations")
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
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
				_, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText("Default subscription set successfully."), nil
			})

			listResourcesTool := mcp.NewTool(
				"list-resources",
				mcp.WithDescription("Lists all resources in a specific Azure resource group or all resources if no group is specified"),
				mcp.WithString("resourceGroupName",
					mcp.Description("The name of the resource group to list resources for (optional)"),
				),
			)
			s.AddTool(listResourcesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				args := []string{"resource", "list"}
				if nameArg, ok := request.GetArguments()["resourceGroupName"]; ok {
					if name, ok := nameArg.(string); ok && name != "" {
						args = append(args, "--resource-group", name)
					}
				}
				azCmd := exec.Command("az", args...)
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			listResourcesByTypeTool := mcp.NewTool(
				"list-resources-by-type",
				mcp.WithDescription("Lists all resources of a specific type, optionally filtered by resource group and subscription"),
				mcp.WithString("resourceType",
					mcp.Required(),
					mcp.Description("The Azure resource type to filter by, e.g., 'Microsoft.Storage/storageAccounts'"),
				),
				mcp.WithString("resourceGroupName",
					mcp.Description("The name of the resource group to list resources for (optional)"),
				),
				mcp.WithString("subscriptionId",
					mcp.Description("The subscription ID containing the resource group (optional)"),
				),
			)
			s.AddTool(listResourcesByTypeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				args := []string{"resource", "list", "--resource-type"}
				typeArg, ok := request.GetArguments()["resourceType"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceType"), nil
				}
				resourceType, ok := typeArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceType, expected string"), nil
				}
				args = append(args, resourceType)
				if nameArg, ok := request.GetArguments()["resourceGroupName"]; ok {
					if name, ok := nameArg.(string); ok && name != "" {
						args = append(args, "--resource-group", name)
					}
				}
				if subIdArg, ok := request.GetArguments()["subscriptionId"]; ok {
					if subId, ok := subIdArg.(string); ok && subId != "" {
						args = append(args, "--subscription", subId)
					}
				}
				azCmd := exec.Command("az", args...)
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			fmt.Println("Starting Azure Metadata MCP server...")

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
