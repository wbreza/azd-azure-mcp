package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
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
				server.WithToolCapabilities(true),
				server.WithRecovery(),
				server.WithLogging(),
				server.WithPromptCapabilities(false),
				server.WithResourceCapabilities(false, false),
				server.WithInstructions("Supports tools interactions with Azure Storage accounts, containers and blobs."),
				server.WithHooks(&server.Hooks{
					OnError: []server.OnErrorHookFunc{
						func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
							fmt.Printf("Error in method %s with ID %v: %s\n", method, id, err.Error())
						},
					},
				}),
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
			)

			uploadBlobTool := mcp.NewTool(
				"upload-blob",
				mcp.WithDescription("Uploads a blob to a container in a storage account."),
				mcp.WithString("storageAccountName", mcp.Required(), mcp.Description("The name of the storage account.")),
				mcp.WithString("containerName", mcp.Required(), mcp.Description("The name of the container.")),
				mcp.WithString("filePath", mcp.Required(), mcp.Description("The local file path to upload.")),
				mcp.WithString("blobName", mcp.Required(), mcp.Description("The name of the blob to create in the container.")),
			)

			downloadBlobTool := mcp.NewTool(
				"download-blob",
				mcp.WithDescription("Downloads a blob from a container in a storage account."),
				mcp.WithString("storageAccountName", mcp.Required(), mcp.Description("The name of the storage account.")),
				mcp.WithString("containerName", mcp.Required(), mcp.Description("The name of the container.")),
				mcp.WithString("blobName", mcp.Required(), mcp.Description("The name of the blob to download.")),
				mcp.WithString("filePath", mcp.Required(), mcp.Description("The local file path to save the blob to.")),
			)

			deleteBlobTool := mcp.NewTool(
				"delete-blob",
				mcp.WithDescription("Deletes a blob from a container in a storage account."),
				mcp.WithString("storageAccountName", mcp.Required(), mcp.Description("The name of the storage account.")),
				mcp.WithString("containerName", mcp.Required(), mcp.Description("The name of the container.")),
				mcp.WithString("blobName", mcp.Required(), mcp.Description("The name of the blob to delete.")),
			)

			// Storage Account Tools

			// List Storage Accounts
			s.AddTool(listAccountsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				cred, err := getCredential()
				if err != nil {
					return mcp.NewToolResultText("Failed to get Azure credential: " + err.Error()), nil
				}
				subId, err := getRequiredStringArg(request.GetArguments(), "subscriptionId")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				client, err := armstorage.NewAccountsClient(subId, cred, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create storage accounts client: " + err.Error()), nil
				}
				pager := client.NewListPager(nil)
				var accounts []*armstorage.Account
				for pager.More() {
					page, err := pager.NextPage(ctx)
					if err != nil {
						return mcp.NewToolResultText("Failed to list storage accounts: " + err.Error()), nil
					}
					accounts = append(accounts, page.Value...)
				}
				result, _ := json.MarshalIndent(accounts, "", "  ")
				return mcp.NewToolResultText(string(result)), nil
			})

			// Create Storage Account
			s.AddTool(createAccountTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				cred, err := getCredential()
				if err != nil {
					return mcp.NewToolResultText("Failed to get Azure credential: " + err.Error()), nil
				}
				name, err := getRequiredStringArg(request.GetArguments(), "storageAccountName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				group, err := getRequiredStringArg(request.GetArguments(), "resourceGroupName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				location, err := getRequiredStringArg(request.GetArguments(), "location")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				subId, err := getRequiredStringArg(request.GetArguments(), "subscriptionId")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				client, err := armstorage.NewAccountsClient(subId, cred, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create storage accounts client: " + err.Error()), nil
				}
				poller, err := client.BeginCreate(ctx, group, name, armstorage.AccountCreateParameters{
					Location: to.Ptr(location),
					Kind:     to.Ptr(armstorage.KindStorageV2),
					SKU:      &armstorage.SKU{Name: to.Ptr(armstorage.SKUNameStandardLRS)},
				}, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to start storage account creation: " + err.Error()), nil
				}
				resp, err := poller.PollUntilDone(ctx, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create storage account: " + err.Error()), nil
				}
				result, _ := json.MarshalIndent(resp.Account, "", "  ")
				return mcp.NewToolResultText(string(result)), nil
			})

			// Show Storage Account
			// Migrated from Azure CLI to Azure SDK for Go
			s.AddTool(showAccountTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				name, err := getRequiredStringArg(request.GetArguments(), "storageAccountName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				group, err := getRequiredStringArg(request.GetArguments(), "resourceGroupName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				subId, err := getRequiredStringArg(request.GetArguments(), "subscriptionId")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				cred, err := getCredential()
				if err != nil {
					return mcp.NewToolResultText("Failed to get Azure credential: " + err.Error()), nil
				}
				client, err := armstorage.NewAccountsClient(subId, cred, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create storage accounts client: " + err.Error()), nil
				}
				resp, err := client.GetProperties(ctx, group, name, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to get storage account details: " + err.Error()), nil
				}
				result, _ := json.MarshalIndent(resp.Account, "", "  ")
				return mcp.NewToolResultText(string(result)), nil
			})

			// List Containers (SDK)
			s.AddTool(listContainersTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				account, err := getRequiredStringArg(request.GetArguments(), "storageAccountName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				cred, err := getCredential()
				if err != nil {
					return mcp.NewToolResultText("Failed to get Azure credential: " + err.Error()), nil
				}
				serviceUrl := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
				client, err := azblob.NewClient(serviceUrl, cred, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create blob service client: " + err.Error()), nil
				}
				pager := client.NewListContainersPager(nil)
				var containers []interface{}
				for pager.More() {
					page, err := pager.NextPage(ctx)
					if err != nil {
						return mcp.NewToolResultText("Failed to list containers: " + err.Error()), nil
					}
					if page.ContainerItems != nil {
						for _, c := range page.ContainerItems {
							containers = append(containers, c)
						}
					} else {
						containers = append(containers, page)
					}
				}
				result, _ := json.MarshalIndent(containers, "", "  ")
				return mcp.NewToolResultText(string(result)), nil
			})

			// Create Container (SDK)
			s.AddTool(createContainerTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				account, err := getRequiredStringArg(request.GetArguments(), "storageAccountName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				container, err := getRequiredStringArg(request.GetArguments(), "containerName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				cred, err := getCredential()
				if err != nil {
					return mcp.NewToolResultText("Failed to get Azure credential: " + err.Error()), nil
				}
				serviceUrl := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
				client, err := azblob.NewClient(serviceUrl, cred, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create blob service client: " + err.Error()), nil
				}
				_, err = client.CreateContainer(ctx, container, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create container: " + err.Error()), nil
				}
				return mcp.NewToolResultText(fmt.Sprintf("Container '%s' created successfully.", container)), nil
			})

			// Delete Container (SDK)
			s.AddTool(deleteContainerTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				account, err := getRequiredStringArg(request.GetArguments(), "storageAccountName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				container, err := getRequiredStringArg(request.GetArguments(), "containerName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				cred, err := getCredential()
				if err != nil {
					return mcp.NewToolResultText("Failed to get Azure credential: " + err.Error()), nil
				}
				serviceUrl := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
				client, err := azblob.NewClient(serviceUrl, cred, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create blob service client: " + err.Error()), nil
				}
				_, err = client.DeleteContainer(ctx, container, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to delete container: " + err.Error()), nil
				}
				return mcp.NewToolResultText(fmt.Sprintf("Container '%s' deleted successfully.", container)), nil
			})

			// Delete Storage Account (SDK)
			// Migrated from Azure CLI to Azure SDK for Go
			s.AddTool(deleteAccountTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				account, err := getRequiredStringArg(request.GetArguments(), "storageAccountName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				group, err := getRequiredStringArg(request.GetArguments(), "resourceGroupName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				subId, err := getRequiredStringArg(request.GetArguments(), "subscriptionId")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				cred, err := getCredential()
				if err != nil {
					return mcp.NewToolResultText("Failed to get Azure credential: " + err.Error()), nil
				}
				client, err := armstorage.NewAccountsClient(subId, cred, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create storage accounts client: " + err.Error()), nil
				}
				_, err = client.Delete(ctx, group, account, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to start storage account deletion: " + err.Error()), nil
				}
				return mcp.NewToolResultText(fmt.Sprintf("Storage account '%s' deleted successfully from resource group '%s'.", account, group)), nil
			})

			// List Blobs (SDK)
			s.AddTool(listBlobsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				account, err := getRequiredStringArg(request.GetArguments(), "storageAccountName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				container, err := getRequiredStringArg(request.GetArguments(), "containerName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				cred, err := getCredential()
				if err != nil {
					return mcp.NewToolResultText("Failed to get Azure credential: " + err.Error()), nil
				}
				serviceUrl := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
				client, err := azblob.NewClient(serviceUrl, cred, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create blob service client: " + err.Error()), nil
				}
				pager := client.NewListBlobsFlatPager(container, nil)
				var blobs []interface{}
				for pager.More() {
					page, err := pager.NextPage(ctx)
					if err != nil {
						return mcp.NewToolResultText("Failed to list blobs: " + err.Error()), nil
					}
					if page.Segment != nil && page.Segment.BlobItems != nil {
						for _, b := range page.Segment.BlobItems {
							blobs = append(blobs, b)
						}
					} else {
						blobs = append(blobs, page)
					}
				}
				result, _ := json.MarshalIndent(blobs, "", "  ")
				return mcp.NewToolResultText(string(result)), nil
			})

			// Upload Blob (SDK)
			s.AddTool(uploadBlobTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				account, err := getRequiredStringArg(request.GetArguments(), "storageAccountName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				container, err := getRequiredStringArg(request.GetArguments(), "containerName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				filePath, err := getRequiredStringArg(request.GetArguments(), "filePath")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				blobName, err := getRequiredStringArg(request.GetArguments(), "blobName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				cred, err := getCredential()
				if err != nil {
					return mcp.NewToolResultText("Failed to get Azure credential: " + err.Error()), nil
				}
				serviceUrl := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
				client, err := azblob.NewClient(serviceUrl, cred, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create blob service client: " + err.Error()), nil
				}
				file, err := os.Open(filePath)
				if err != nil {
					return mcp.NewToolResultText("Failed to open file: " + err.Error()), nil
				}
				defer file.Close()
				_, err = client.UploadStream(ctx, container, blobName, file, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to upload blob: " + err.Error()), nil
				}
				return mcp.NewToolResultText(fmt.Sprintf("Blob '%s' uploaded successfully to container '%s'.", blobName, container)), nil
			})

			// Download Blob (SDK)
			s.AddTool(downloadBlobTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				account, err := getRequiredStringArg(request.GetArguments(), "storageAccountName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				container, err := getRequiredStringArg(request.GetArguments(), "containerName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				blobName, err := getRequiredStringArg(request.GetArguments(), "blobName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				filePath, err := getRequiredStringArg(request.GetArguments(), "filePath")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				cred, err := getCredential()
				if err != nil {
					return mcp.NewToolResultText("Failed to get Azure credential: " + err.Error()), nil
				}
				serviceUrl := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
				client, err := azblob.NewClient(serviceUrl, cred, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create blob service client: " + err.Error()), nil
				}
				resp, err := client.DownloadStream(ctx, container, blobName, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to download blob: " + err.Error()), nil
				}
				file, err := os.Create(filePath)
				if err != nil {
					return mcp.NewToolResultText("Failed to create file: " + err.Error()), nil
				}
				defer file.Close()
				_, err = io.Copy(file, resp.Body)
				if err != nil {
					return mcp.NewToolResultText("Failed to write blob to file: " + err.Error()), nil
				}
				return mcp.NewToolResultText(fmt.Sprintf("Blob '%s' downloaded successfully to '%s'.", blobName, filePath)), nil
			})

			// Delete Blob (SDK)
			s.AddTool(deleteBlobTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				account, err := getRequiredStringArg(request.GetArguments(), "storageAccountName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				container, err := getRequiredStringArg(request.GetArguments(), "containerName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				blobName, err := getRequiredStringArg(request.GetArguments(), "blobName")
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				cred, err := getCredential()
				if err != nil {
					return mcp.NewToolResultText("Failed to get Azure credential: " + err.Error()), nil
				}
				serviceUrl := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
				client, err := azblob.NewClient(serviceUrl, cred, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to create blob service client: " + err.Error()), nil
				}
				_, err = client.DeleteBlob(ctx, container, blobName, nil)
				if err != nil {
					return mcp.NewToolResultText("Failed to delete blob: " + err.Error()), nil
				}
				return mcp.NewToolResultText(fmt.Sprintf("Blob '%s' deleted successfully from container '%s'.", blobName, container)), nil
			})

			// Start the HTTP server on http://localhost:8081/mcp
			sseServer := server.NewStreamableHTTPServer(s,
				server.WithEndpointPath("/storage/mcp"),
			)
			if err := sseServer.Start("localhost:8081"); err != nil {
				return fmt.Errorf("Failed to start SSE server: %v", err)
			}

			return nil
		},
	}

	serverGroup.AddCommand(startCmd)

	return serverGroup
}

// Helper to get a DefaultAzureCredential as azcore.TokenCredential
func getCredential() (azcore.TokenCredential, error) {
	// For POC - this is just using DefaultAzureCredential since we can run this locally and still get a credential
	// In reality when running in Azure we would need to leverage proper OAuth flow
	// https://modelcontextprotocol.io/specification/2025-03-26/basic/authorization
	return azidentity.NewDefaultAzureCredential(nil)
}

// Helper to extract and validate a required string argument
func getRequiredStringArg(args map[string]interface{}, key string) (string, error) {
	val, ok := args[key]
	if !ok {
		return "", fmt.Errorf("Missing required argument: %s", key)
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("Invalid type for argument: %s, expected string", key)
	}
	if str == "" {
		return "", fmt.Errorf("Argument '%s' cannot be empty", key)
	}
	return str, nil
}
