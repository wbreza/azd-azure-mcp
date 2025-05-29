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
				server.WithToolCapabilities(true),
				server.WithRecovery(),
				server.WithLogging(),
				server.WithPromptCapabilities(false),
				server.WithResourceCapabilities(false, false),
				server.WithInstructions("Supports tools for interacting with Azure subscriptions, resource groups and generic resources."),
			)

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
				return runAzCommandWithResult(azCmd), nil
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
				return runAzCommandWithResult(azCmd), nil
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
				return runAzCommandWithResult(azCmd), nil
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
				return runAzCommandWithResult(azCmd), nil
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
				return runAzCommandWithResult(azCmd), nil
			})

			// Delete resource group tool

			deleteResourceGroupTool := mcp.NewTool(
				"delete-resource-group",
				mcp.WithDescription("Deletes a specific Azure resource group."),
				mcp.WithString("resourceGroupName",
					mcp.Required(),
					mcp.Description("The name of the resource group to delete"),
				),
				mcp.WithString("subscriptionId",
					mcp.Required(),
					mcp.Description("The subscription ID containing the resource group"),
				),
			)
			s.AddTool(deleteResourceGroupTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				azCmd := exec.Command("az", "group", "delete", "--name", name, "--subscription", subId, "--yes")
				return runAzCommandWithResult(azCmd), nil
			})

			// Exists resource group tool

			existsResourceGroupTool := mcp.NewTool(
				"exists-resource-group",
				mcp.WithDescription("Checks if a specific Azure resource group exists."),
				mcp.WithString("resourceGroupName",
					mcp.Required(),
					mcp.Description("The name of the resource group to check"),
				),
				mcp.WithString("subscriptionId",
					mcp.Required(),
					mcp.Description("The subscription ID containing the resource group"),
				),
			)
			s.AddTool(existsResourceGroupTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				azCmd := exec.Command("az", "group", "exists", "--name", name, "--subscription", subId)
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
