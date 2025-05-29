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
					To execute an action, set the "tool", "command", and convert the users intent into the "parameters" based on the discovered schema.
					Always use this tool for any Azure or "azd" related operation requiring up-to-date, dynamic, and interactive capabilities.
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
					This server/tool provides real-time, programmatic access to all Azure products, services, and resources,
					as well as all interactions with the Azure Developer CLI (azd).
					Use this tool for any Azure control plane or data plane operation, including resource management and automation.
					To discover available capabilities, call the tool with the "learn" parameter to get a list of top-level tools.
					To explore further, set "learn" and specify a tool name to retrieve supported commands and their parameters.
					To execute an action, set the "tool", "command", and convert the users intent into the "parameters" based on the discovered schema.
					Always use this tool for any Azure or "azd" related operation requiring up-to-date, dynamic, and interactive capabilities.
				`),
				mcp.WithString("intent",
					mcp.Required(),
					mcp.Description("The intent of the operation the user wants to perform against azure."),
				),
				mcp.WithString("tool",
					mcp.Description("The azure tool to use to execute the operation."),
				),
				mcp.WithString("command",
					mcp.Description("The command to execute against the specified tool."),
				),
				mcp.WithObject("parameters",
					mcp.Description("The parameters to pass to the tool"),
				),
				mcp.WithBoolean("learn",
					mcp.Description("To learn about the tool and its supported child tools and parameters."),
					mcp.DefaultBool(false),
				),
			)

			s.AddTool(azureTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				toolName, hasToolName := request.GetArguments()["tool"].(string)

				learn, ok := request.GetArguments()["learn"].(bool)
				if ok && learn {
					if hasToolName && toolName != "" {
						tm, ok := toolMetadataMap[toolName]
						if !ok {
							return mcp.NewToolResultText(fmt.Sprintf(`
								Tool %s not found
								Run again with the "learn" argument and empty "tool" to get a list of available tools and their parameters.
							`, toolName)), nil
						}

						// Caching at caller side
						cacheKey := toolName
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
						toolsJson, err := json.MarshalIndent(childTools, "", "  ")
						if err != nil {
							return nil, fmt.Errorf("failed get get learn content: %w", err)
						}
						learnContent := fmt.Sprintf(`
							Here are the available command and their parameters for '%s' tool.
							If you do not find a suitable tool, run again with the "learn" argument and empty "tool" to get a list of available tools and their parameters.

							%s
						`, toolName, string(toolsJson))

						return mcp.NewToolResultText(learnContent), nil
					}

					toolsJson, err := json.MarshalIndent(childTools, "", "  ")
					if err != nil {
						return nil, fmt.Errorf("failed get get learn content: %w", err)
					}

					return mcp.NewToolResultText(string(toolsJson)), nil
				}

				commandName, hasCommandName := request.GetArguments()["command"].(string)
				if !hasToolName || !hasCommandName {
					return mcp.NewToolResultText(`
						The "tool" and "command" parameters are required when not learning
						Run again with the "learn" argument to get a list of available tools and their parameters.
						To learn about a specific tool, use the "tool" argument with the name of the tool.
					`), nil
				}
				tm, ok := toolMetadataMap[toolName]
				if !ok {
					return mcp.NewToolResultText(fmt.Sprintf(`
						Tool %s not found
						Run again with the "learn" argument and empty "tool" to get a list of available tools and their parameters.
					`, toolName)), nil
				}

				cacheKey := toolName
				toolClient, ok := toolClientCache[cacheKey]
				if !ok {
					var err error
					toolClient, err = tm.CreateClient(ctx)
					if err != nil {
						return mcp.NewToolResultText(fmt.Sprintf("Failed to start tool client: %v", err)), nil
					}
					toolClientCache[cacheKey] = toolClient
				}

				params := request.GetArguments()["parameters"]
				childRequest := request
				childRequest.Params.Name = commandName
				childRequest.Params.Arguments = params

				toolCallResult, err := toolClient.CallTool(ctx, childRequest)
				if err != nil {
					return mcp.NewToolResultText(fmt.Sprintf(`
						There was an error finding or calling tool and command.
						Failed to call tool: %s, command: %s, Error: %v

						Run again with the "learn" argument and the "tool" name to get a list of available tools and their parameters.
					`, toolName, commandName, err)), nil
				}
				return toolCallResult, nil
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
