package cmd

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

func runAzCommandWithResult(cmd *exec.Cmd) *mcp.CallToolResult {
	output, err := cmd.CombinedOutput()
	result := string(output)
	if err != nil {
		result = result + "\n[error] " + err.Error()
	}
	return mcp.NewToolResultText(result)
}

func newServerCommand() *cobra.Command {
	serverGroup := &cobra.Command{
		Use: "server",
	}

	startCmd := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := server.NewMCPServer(
				"KeyVault",
				"1.0.0",
				server.WithToolCapabilities(true),
				server.WithRecovery(),
				server.WithLogging(),
				server.WithPromptCapabilities(false),
				server.WithResourceCapabilities(false, false),
				server.WithInstructions("Supports tools for interacting with Azure Key Vaults and secrets."),
			)

			// list-keyvaults
			listKeyVaultsTool := mcp.NewTool(
				"list-keyvaults",
				mcp.WithDescription("Lists all Key Vaults in a subscription or resource group"),
				mcp.WithString("resourceGroupName",
					mcp.Description("The name of the resource group to filter by (optional)"),
				),
			)
			s.AddTool(listKeyVaultsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				args := []string{"keyvault", "list"}
				if rg, ok := request.GetArguments()["resourceGroupName"]; ok {
					if rgStr, ok := rg.(string); ok && rgStr != "" {
						args = append(args, "--resource-group", rgStr)
					}
				}
				azCmd := exec.Command("az", args...)
				return runAzCommandWithResult(azCmd), nil
			})

			// create-keyvault
			createKeyVaultTool := mcp.NewTool(
				"create-keyvault",
				mcp.WithDescription("Creates a new Azure Key Vault"),
				mcp.WithString("name",
					mcp.Required(),
					mcp.Description("The name of the Key Vault to create"),
				),
				mcp.WithString("resourceGroupName",
					mcp.Required(),
					mcp.Description("The name of the resource group to create the Key Vault in"),
				),
				mcp.WithString("location",
					mcp.Required(),
					mcp.Description("The Azure region for the Key Vault"),
				),
			)
			s.AddTool(createKeyVaultTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["name"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: name"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: name, expected string"), nil
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
				azCmd := exec.Command("az", "keyvault", "create", "--name", name, "--resource-group", group, "--location", location)
				return runAzCommandWithResult(azCmd), nil
			})

			// delete-keyvault
			deleteKeyVaultTool := mcp.NewTool(
				"delete-keyvault",
				mcp.WithDescription("Deletes a Key Vault from a resource group"),
				mcp.WithString("name",
					mcp.Required(),
					mcp.Description("The name of the Key Vault to delete"),
				),
				mcp.WithString("resourceGroupName",
					mcp.Required(),
					mcp.Description("The name of the resource group containing the Key Vault"),
				),
			)
			s.AddTool(deleteKeyVaultTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["name"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: name"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: name, expected string"), nil
				}
				groupArg, ok := request.GetArguments()["resourceGroupName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroupName"), nil
				}
				group, ok := groupArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroupName, expected string"), nil
				}
				azCmd := exec.Command("az", "keyvault", "delete", "--name", name, "--resource-group", group)
				return runAzCommandWithResult(azCmd), nil
			})

			// list-secrets
			listSecretsTool := mcp.NewTool(
				"list-secrets",
				mcp.WithDescription("Lists all secrets in a Key Vault"),
				mcp.WithString("vaultName",
					mcp.Required(),
					mcp.Description("The name of the Key Vault to list secrets for"),
				),
			)
			s.AddTool(listSecretsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				vaultArg, ok := request.GetArguments()["vaultName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: vaultName"), nil
				}
				vault, ok := vaultArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: vaultName, expected string"), nil
				}
				azCmd := exec.Command("az", "keyvault", "secret", "list", "--vault-name", vault)
				return runAzCommandWithResult(azCmd), nil
			})

			// show-keyvault
			showKeyVaultTool := mcp.NewTool(
				"show-keyvault",
				mcp.WithDescription("Shows details of a specific Azure Key Vault"),
				mcp.WithString("name",
					mcp.Required(),
					mcp.Description("The name of the Key Vault to show details for"),
				),
				mcp.WithString("resourceGroupName",
					mcp.Required(),
					mcp.Description("The name of the resource group containing the Key Vault"),
				),
			)
			s.AddTool(showKeyVaultTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["name"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: name"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: name, expected string"), nil
				}
				groupArg, ok := request.GetArguments()["resourceGroupName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroupName"), nil
				}
				group, ok := groupArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroupName, expected string"), nil
				}
				azCmd := exec.Command("az", "keyvault", "show", "--name", name, "--resource-group", group)
				return runAzCommandWithResult(azCmd), nil
			})

			// show-keyvault-secret
			showKeyVaultSecretTool := mcp.NewTool(
				"show-keyvault-secret",
				mcp.WithDescription("Shows details of a specific secret in a Key Vault"),
				mcp.WithString("vaultName",
					mcp.Required(),
					mcp.Description("The name of the Key Vault containing the secret"),
				),
				mcp.WithString("secretName",
					mcp.Required(),
					mcp.Description("The name of the secret to show"),
				),
			)
			s.AddTool(showKeyVaultSecretTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				vaultArg, ok := request.GetArguments()["vaultName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: vaultName"), nil
				}
				vault, ok := vaultArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: vaultName, expected string"), nil
				}
				secretArg, ok := request.GetArguments()["secretName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: secretName"), nil
				}
				secret, ok := secretArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: secretName, expected string"), nil
				}
				azCmd := exec.Command("az", "keyvault", "secret", "show", "--vault-name", vault, "--name", secret)
				return runAzCommandWithResult(azCmd), nil
			})

			// set-keyvault-secret
			setKeyVaultSecretTool := mcp.NewTool(
				"set-keyvault-secret",
				mcp.WithDescription("Sets a secret in a Key Vault. Creates or updates the secret value."),
				mcp.WithString("vaultName",
					mcp.Required(),
					mcp.Description("The name of the Key Vault to set the secret in"),
				),
				mcp.WithString("secretName",
					mcp.Required(),
					mcp.Description("The name of the secret to set"),
				),
				mcp.WithString("value",
					mcp.Required(),
					mcp.Description("The value to set for the secret"),
				),
			)
			s.AddTool(setKeyVaultSecretTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				vaultArg, ok := request.GetArguments()["vaultName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: vaultName"), nil
				}
				vault, ok := vaultArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: vaultName, expected string"), nil
				}
				secretArg, ok := request.GetArguments()["secretName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: secretName"), nil
				}
				secret, ok := secretArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: secretName, expected string"), nil
				}
				valueArg, ok := request.GetArguments()["value"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: value"), nil
				}
				value, ok := valueArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: value, expected string"), nil
				}
				azCmd := exec.Command("az", "keyvault", "secret", "set", "--vault-name", vault, "--name", secret, "--value", value)
				return runAzCommandWithResult(azCmd), nil
			})

			// delete-keyvault-secret
			deleteKeyVaultSecretTool := mcp.NewTool(
				"delete-keyvault-secret",
				mcp.WithDescription("Deletes a secret from a Key Vault"),
				mcp.WithString("vaultName",
					mcp.Required(),
					mcp.Description("The name of the Key Vault containing the secret"),
				),
				mcp.WithString("secretName",
					mcp.Required(),
					mcp.Description("The name of the secret to delete"),
				),
			)
			s.AddTool(deleteKeyVaultSecretTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				vaultArg, ok := request.GetArguments()["vaultName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: vaultName"), nil
				}
				vault, ok := vaultArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: vaultName, expected string"), nil
				}
				secretArg, ok := request.GetArguments()["secretName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: secretName"), nil
				}
				secret, ok := secretArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: secretName, expected string"), nil
				}
				azCmd := exec.Command("az", "keyvault", "secret", "delete", "--vault-name", vault, "--name", secret)
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
