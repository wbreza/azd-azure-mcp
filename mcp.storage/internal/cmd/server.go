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
				"Storage",
				"1.0.0",
				server.WithToolCapabilities(false),
				server.WithRecovery(),
				server.WithLogging(),
				server.WithInstructions("Supports tools interactions with Azure Storage accounts, containers and blobs."),
			)

			listAccountsTool := mcp.NewTool(
				"list-storage-accounts",
				mcp.WithDescription("Lists all storage accounts in the subscription"),
			)

			createAccountTool := mcp.NewTool(
				"create-storage-account",
				mcp.WithDescription("Creates a new azure storage account"),
				mcp.WithString("storageAccountName",
					mcp.Required(),
					mcp.Description("The name of the storage account to create"),
				),
				mcp.WithString("resourceGroupName",
					mcp.Required(),
					mcp.Description("The name of the resource group to create the storage account in"),
				),
				mcp.WithString("location",
					mcp.Required(),
					mcp.Description("The location to create the storage account in"),
				),
			)

			showAccountTool := mcp.NewTool(
				"show-storage-account",
				mcp.WithDescription("Shows details of a specific Azure storage account"),
				mcp.WithString("storageAccountName",
					mcp.Required(),
					mcp.Description("The name of the storage account to show details for"),
				),
				mcp.WithString("resourceGroupName",
					mcp.Required(),
					mcp.Description("The name of the resource group containing the storage account"),
				),
			)

			// Container management tools
			listContainersTool := mcp.NewTool(
				"list-containers",
				mcp.WithDescription("Lists all containers in a storage account"),
				mcp.WithString("storageAccountName",
					mcp.Required(),
					mcp.Description("The name of the storage account"),
				),
			)

			createContainerTool := mcp.NewTool(
				"create-container",
				mcp.WithDescription("Creates a new container in a storage account"),
				mcp.WithString("storageAccountName",
					mcp.Required(),
					mcp.Description("The name of the storage account"),
				),
				mcp.WithString("containerName",
					mcp.Required(),
					mcp.Description("The name of the container to create"),
				),
			)

			deleteContainerTool := mcp.NewTool(
				"delete-container",
				mcp.WithDescription("Deletes a container from a storage account"),
				mcp.WithString("storageAccountName",
					mcp.Required(),
					mcp.Description("The name of the storage account"),
				),
				mcp.WithString("containerName",
					mcp.Required(),
					mcp.Description("The name of the container to delete"),
				),
			)

			deleteAccountTool := mcp.NewTool(
				"delete-storage-account",
				mcp.WithDescription("Deletes a storage account from a resource group"),
				mcp.WithString("storageAccountName",
					mcp.Required(),
					mcp.Description("The name of the storage account to delete"),
				),
				mcp.WithString("resourceGroupName",
					mcp.Required(),
					mcp.Description("The name of the resource group containing the storage account"),
				),
			)

			listBlobsTool := mcp.NewTool(
				"list-blobs",
				mcp.WithDescription("Lists all blobs in a container within a storage account"),
				mcp.WithString("storageAccountName",
					mcp.Required(),
					mcp.Description("The name of the storage account"),
				),
				mcp.WithString("containerName",
					mcp.Required(),
					mcp.Description("The name of the container to list blobs for"),
				),
				mcp.WithString("resourceGroupName",
					mcp.Required(),
					mcp.Description("The name of the resource group containing the storage account"),
				),
			)

			s.AddTool(listAccountsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				azCmd := exec.Command("az", "storage", "account", "list")
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}

				return mcp.NewToolResultText(string(output)), nil
			})

			s.AddTool(createAccountTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["storageAccountName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: storageAccountName"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: storageAccountName, expected string"), nil
				}
				groupArg, ok := request.GetArguments()["resourceGroupName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroupName"), nil
				}
				group, ok := groupArg.(string)
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
				azCmd := exec.Command("az", "storage", "account", "create",
					"--name", name,
					"--resource-group", group,
					"--location", location,
				)
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			s.AddTool(showAccountTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["storageAccountName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: storageAccountName"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: storageAccountName, expected string"), nil
				}
				groupArg, ok := request.GetArguments()["resourceGroupName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroupName"), nil
				}
				group, ok := groupArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroupName, expected string"), nil
				}
				azCmd := exec.Command("az", "storage", "account", "show",
					"--name", name,
					"--resource-group", group,
				)
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			s.AddTool(listContainersTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["storageAccountName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: storageAccountName"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: storageAccountName, expected string"), nil
				}
				azCmd := exec.Command("az", "storage", "container", "list", "--account-name", name, "--auth-mode", "login")
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			s.AddTool(createContainerTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				accountArg, ok := request.GetArguments()["storageAccountName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: storageAccountName"), nil
				}
				account, ok := accountArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: storageAccountName, expected string"), nil
				}
				containerArg, ok := request.GetArguments()["containerName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: containerName"), nil
				}
				container, ok := containerArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: containerName, expected string"), nil
				}
				azCmd := exec.Command("az", "storage", "container", "create", "--account-name", account, "--name", container, "--auth-mode", "login")
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			s.AddTool(deleteContainerTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				accountArg, ok := request.GetArguments()["storageAccountName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: storageAccountName"), nil
				}
				account, ok := accountArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: storageAccountName, expected string"), nil
				}
				containerArg, ok := request.GetArguments()["containerName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: containerName"), nil
				}
				container, ok := containerArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: containerName, expected string"), nil
				}
				azCmd := exec.Command("az", "storage", "container", "delete", "--account-name", account, "--name", container, "--auth-mode", "login")
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			s.AddTool(deleteAccountTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				accountArg, ok := request.GetArguments()["storageAccountName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: storageAccountName"), nil
				}
				account, ok := accountArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: storageAccountName, expected string"), nil
				}
				groupArg, ok := request.GetArguments()["resourceGroupName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroupName"), nil
				}
				group, ok := groupArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroupName, expected string"), nil
				}
				azCmd := exec.Command("az", "storage", "account", "delete", "--name", account, "--resource-group", group, "--yes")
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			s.AddTool(listBlobsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				accountArg, ok := request.GetArguments()["storageAccountName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: storageAccountName"), nil
				}
				account, ok := accountArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: storageAccountName, expected string"), nil
				}
				containerArg, ok := request.GetArguments()["containerName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: containerName"), nil
				}
				container, ok := containerArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: containerName, expected string"), nil
				}
				groupArg, ok := request.GetArguments()["resourceGroupName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroupName"), nil
				}
				group, ok := groupArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroupName, expected string"), nil
				}
				azCmd := exec.Command("az", "storage", "blob", "list", "--account-name", account, "--container-name", container, "--resource-group", group, "--auth-mode", "login")
				output, err := azCmd.CombinedOutput()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				return mcp.NewToolResultText(string(output)), nil
			})

			fmt.Println("Starting Azure Storage MCP server...")

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
