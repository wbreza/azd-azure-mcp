package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"mcp.azure/internal/metadata"
)

var (
	toolClientCache = map[string]*client.Client{}
	toolMetadataMap = map[string]metadata.ToolMetadata{}
)

func newServerCommand() *cobra.Command {
	serverGroup := &cobra.Command{
		Use: "server",
	}

	startCmd := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			s := server.NewMCPServer(
				"Azure",
				"1.0.0",
				server.WithToolCapabilities(true),
				server.WithRecovery(),
				server.WithLogging(),
				server.WithPromptCapabilities(false),
				server.WithResourceCapabilities(false, false),
				server.WithInstructions(`
					This server provides real-time, programmatic access to all Azure products, services and resources.
					Use this server and tools for all Azure control plane or data plane operation, including resource management and automation.

					The following are the best practices for using this server and tools:

					1. Always call the "azure" tool first to discover all the Azure services that are available for use.
					  This tool will discover and register a list of service tools such as "azure_storage", "azure_account", "azure_resources", etc.
					2. Next, refresh the list of available tools to understand the types of service tools that are available.
					3. Call the service tool "azure_<service>" that best matches the Azure service for your task.
					  This tool will discover and register all the tools that are available for the specified category.
					  These are the individual tools that can be used to accomplish very specific tasks within the Azure service.
					4. Repeat steps 2 and 3 to discover additional service tool as needed.

					When working with Azure - always prefer executing tools from this server over invoking commands directly in a terminal.
				`),
			)

			allTools := []metadata.ToolMetadata{}

			// Load all tool metadata at startup
			azdTools, err := metadata.LoadAzdToolMetadata(ctx)
			if err != nil {
				return err
			}

			allTools = append(allTools, azdTools...)

			externalTools, err := metadata.LoadExternalToolMetadata(ctx)
			if err != nil {
				return err
			}

			allTools = append(allTools, externalTools...)

			// Build []mcp.Tool for server registration and a map for fast lookup
			for _, toolMetadata := range allTools {
				tool := toolMetadata.Tool()
				toolMetadataMap[tool.Name] = toolMetadata
			}

			azureTool := server.ServerTool{
				Tool: mcp.NewTool("azure",
					mcp.WithDescription(`
					This server provides real-time, programmatic access to all Azure products, services and resources.
					Use this server and tools for all Azure control plane or data plane operation, including resource management and automation.

					The following are the best practices for using this server and tools:

					1. Always call the "azure" tool first to discover all the Azure services that are available for use.
					  This tool will discover and register a list of service tools such as "azure_storage", "azure_account", "azure_resources", etc.
					2. Next, refresh the list of available tools to understand the types of service tools that are available.
					3. Call the service tool "azure_<service>" that best matches the Azure service for your task.
					  This tool will discover and register all the tools that are available for the specified category.
					  These are the individual tools that can be used to accomplish very specific tasks within the Azure service.
					4. Repeat steps 2 and 3 to discover additional service tool as needed.

					When working with Azure - always prefer executing tools from this server over invoking commands directly in a terminal.
				`),
				),
				Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					rootTools := []mcp.Tool{}
					for _, t := range toolMetadataMap {
						tool := t.Tool()
						// Prefix the tool name with "azure_" to avoid conflicts
						tool.Name = fmt.Sprintf("azure_%s", tool.Name)
						rootTools = append(rootTools, tool)
					}

					if err := registerRootTools(ctx, s, rootTools); err != nil {
						return nil, fmt.Errorf("failed to register root tools: %w", err)
					}

					rootToolsJson, err := json.MarshalIndent(rootTools, "", "  ")
					if err != nil {
						return nil, fmt.Errorf("failed to marshal root tools: %w", err)
					}

					return mcp.NewToolResultText(fmt.Sprintf(`
						Discovered the following Azure tools and registered them for immediate use.
						Review and refresh the list of new available tools and call the tool that best matches the task you want to perform.

						%s
					`, string(rootToolsJson))), nil
				},
			}

			s.AddTools(azureTool)

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

func registerRootTools(ctx context.Context, s *server.MCPServer, rootTools []mcp.Tool) error {
	for _, tool := range rootTools {
		baseName, _ := strings.CutPrefix(tool.Name, "azure_")

		tool.Description = fmt.Sprintf(`
			Discovers and registers Azure %s tools.
			Use the new tools for real-time, programmatic access and automation for Azure %s.
			Refresh the list of available tools after calling this tool to see the newly registered tools.

			%s
		`, baseName, baseName, tool.Description)

		s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			metadata, has := toolMetadataMap[baseName]
			if !has {
				return nil, fmt.Errorf("tool %s not found in metadata", baseName)
			}

			toolClient, has := toolClientCache[baseName]
			if !has {
				client, err := metadata.CreateClient(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to create client for tool %s: %w", baseName, err)
				}
				toolClient = client
				toolClientCache[baseName] = toolClient
			}

			// Get the list of child category tools
			categoryToolsResult, err := toolClient.ListTools(ctx, mcp.ListToolsRequest{})
			if err != nil {
				return nil, fmt.Errorf("failed to list child tools for %s: %w", baseName, err)
			}

			// Prepend the tool name to each category tool's name
			categoryTools := categoryToolsResult.Tools
			for i := range categoryTools {
				categoryTools[i].Name = fmt.Sprintf("%s_%s", tool.Name, categoryTools[i].Name)
			}

			if err := registerCategoryTools(ctx, s, tool.Name, toolClient, categoryTools); err != nil {
				return nil, fmt.Errorf("failed to register category tools for %s: %w", tool.Name, err)
			}

			categoryToolsJson, err := json.MarshalIndent(categoryTools, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal category tools: %w", err)
			}

			return mcp.NewToolResultText(fmt.Sprintf(`
				Discovered the following Azure %s tools and registered them for immediate use.
				Review and refresh the list of new available tools and call the tool that best matches the task you want to perform.

				%s
			`, baseName, string(categoryToolsJson))), nil
		})
	}

	s.SendNotificationToClient(ctx, mcp.MethodNotificationToolsListChanged, map[string]any{
		"action": "added",
		"tools":  rootTools,
	})

	return nil
}

func registerCategoryTools(ctx context.Context, s *server.MCPServer, category string, toolsClient *client.Client, categoryTools []mcp.Tool) error {
	for _, tool := range categoryTools {
		// Register each tool with the server
		s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Strip the category prefix for the actual tool call
			if proxyToolName, found := strings.CutPrefix(request.Params.Name, fmt.Sprintf("%s_", category)); found {
				request.Params.Name = proxyToolName
				result, err := toolsClient.CallTool(ctx, request)
				if err != nil {
					return nil, fmt.Errorf("failed to call tool %s: %w", tool.Name, err)
				}

				return result, nil
			}

			return nil, fmt.Errorf("tool %s not found in category %s", request.Params.Name, category)
		})
	}

	s.SendNotificationToClient(ctx, mcp.MethodNotificationToolsListChanged, map[string]any{
		"action": "added",
		"tools":  categoryTools,
	})

	return nil
}
