package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"mcp.azure/internal/metadata"
)

// Update toolClientCache to use map[string]*client.Client
var toolClientCache = make(map[string]*client.Client)

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
					This server/tool provides real-time, programmatic access to all Azure products, services, and resources,
					as well as all interactions with the Azure Developer CLI (azd).
					Use this tool for any Azure control plane or data plane operation, including resource management and automation.
					To discover available capabilities, call the tool with the "learn" parameter to get a list of top-level tools.
					To explore further, set "learn" and specify a tool name to retrieve supported commands and their parameters.
					Once discovered, tools will be added and a notification will be sent for the added tools.
					Newly discovered tools can be invoked immediately.
				`),
			)

			// Load all tool metadata at startup
			azdTools, err := metadata.LoadAzdToolMetadata(ctx)
			if err != nil {
				return err
			}

			externalTools, err := metadata.LoadExternalToolMetadata(ctx)
			if err != nil {
				return err
			}

			allTools := append(azdTools, externalTools...)

			// Build []mcp.Tool for server registration and a map for fast lookup
			var childTools []mcp.Tool
			toolMetadataMap := make(map[string]metadata.ToolMetadata)
			for _, t := range allTools {
				meta := t.Metadata()
				childTools = append(childTools, meta)
				toolMetadataMap[meta.Name] = t
			}

			azureTool := mcp.NewTool(
				"azure",
				mcp.WithDescription(`
					This tool provides discovery of additional tools that provide real-time, programmatic access to all Azure products, services, and resources,
					This includes Azure control plane or data plane operation, including resource management and automation.
					To discover available capabilities, call the tool to get a list of top-level categories.
					To explore further, specify a category name to discovery additional tools for a category.
					Once discovered, tools will be added and a notification will be sent for the added tools.
					Newly discovered tools can be invoked immediately.
				`),
				mcp.WithString("category",
					mcp.Description("The azure tool to use to execute the operation."),
				),
			)

			s.AddTool(azureTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				categoryName, hasToolName := request.GetArguments()["category"].(string)

				if hasToolName && categoryName != "" {
					tm, ok := toolMetadataMap[categoryName]
					if !ok {
						return mcp.NewToolResultText(fmt.Sprintf(`
							Category %s not found
							Run again with the empty category to get a list of top level categories
						`, categoryName)), nil
					}

					// Caching at caller side
					cacheKey := categoryName
					toolClient, ok := toolClientCache[cacheKey]
					if !ok {
						var err error
						toolClient, err = tm.CreateClient(ctx)
						if err != nil {
							return mcp.NewToolResultText(fmt.Sprintf("Failed to start tool client: %v", err)), nil
						}

						toolClientCache[cacheKey] = toolClient
					}

					childTools, err := toolClient.ListTools(ctx, mcp.ListToolsRequest{})
					if err != nil {
						return nil, fmt.Errorf("failed to get child tools: %w", err)
					}

					for _, childTool := range childTools.Tools {
						s.AddTool(childTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
							toolCallResult, err := toolClient.CallTool(ctx, request)
							if err != nil {
								return mcp.NewToolResultText(fmt.Sprintf(`
									There was an error finding or calling tool.
									Failed to call tool: %s, Error: %v

									Run "learn" tool to discovery a list of available tools for a category.
								`, categoryName, err)), nil
							}
							return toolCallResult, nil
						})
					}

					if err := s.SendNotificationToClient(ctx, mcp.MethodNotificationToolsListChanged, map[string]any{
						"action": "added",
						"tools":  childTools,
					}); err != nil {
						return nil, fmt.Errorf("failed to send tools list changed notification: %w", err)
					}

					toolsJson, err := json.MarshalIndent(childTools, "", "  ")
					if err != nil {
						return nil, fmt.Errorf("failed get get learn content: %w", err)
					}
					learnContent := fmt.Sprintf(`
						New tools have been discovered for the '%s' category. They have been updated in the tools list.

						Here are the available command and their parameters for '%s' tool.
						If you do not find a suitable tool, run again with the "learn" argument and empty "tool" to get a list of available tools and their parameters.

						%s
					`, categoryName, categoryName, string(toolsJson))

					return mcp.NewToolResultText(learnContent), nil
				}

				toolsJson, err := json.MarshalIndent(childTools, "", "  ")
				if err != nil {
					return nil, fmt.Errorf("failed get get learn content: %w", err)
				}

				return mcp.NewToolResultText(string(toolsJson)), nil
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
