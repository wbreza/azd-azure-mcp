package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

type mcpExtensionMetadata struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Namespace   string `json:"namespace"`
	Version     string `json:"version"`
	Installed   bool   `json:"installed"`
}

var toolClientCache = make(map[string]client.MCPClient)

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
				server.WithToolCapabilities(false),
				server.WithRecovery(),
				server.WithLogging(),
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

			childTools, extList, err := loadMcpChildToolsMetadata(ctx)
			if err != nil {
				return err
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
						ext := findExtensionByToolName(extList, toolName)
						if ext == nil {
							return mcp.NewToolResultText(fmt.Sprintf(`
								Tool %s not found
								Run again with the "learn" argument and empty "tool" to get a list of available tools and their parameters.
							`, toolName)), nil
						}

						toolClient, err := getOrStartMcpClient(ctx, *ext)
						if err != nil {
							return mcp.NewToolResultText(fmt.Sprintf("Failed to start tool client: %v", err)), nil
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

				ext := findExtensionByToolName(extList, toolName)
				if ext == nil {
					return mcp.NewToolResultText(fmt.Sprintf(`
						Tool %s not found
						Run again with the "learn" argument and empty "tool" to get a list of available tools and their parameters.
					`, toolName)), nil
				}

				toolClient, err := getOrStartMcpClient(ctx, *ext)
				if err != nil {
					return mcp.NewToolResultText(fmt.Sprintf("Failed to start tool client: %v", err)), nil
				}

				// Transform the incoming "parameters" argument into a properly formatted CallToolRequest for the child tool
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

			fmt.Println("Starting Azure MCP server...")

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

// Helper to get or start an MCP client for a tool, installing if needed
func getOrStartMcpClient(ctx context.Context, ext mcpExtensionMetadata) (client.MCPClient, error) {
	if cached, ok := toolClientCache[ext.ID]; ok {
		return cached, nil
	}

	if !ext.Installed {
		// Install the extension if not installed
		installCmd := exec.Command("azd", "ext", "install", ext.ID)
		installOut, err := installCmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to install extension %s: %w\n%s", ext.ID, err, string(installOut))
		}
	}

	// Use the namespace to build the server start command
	nsParts := strings.Split(ext.Namespace, ".")
	if len(nsParts) < 2 {
		return nil, fmt.Errorf("invalid namespace for extension: %s", ext.Namespace)
	}
	args := append([]string{}, nsParts...)
	args = append(args, "server", "start")

	mcpClient, err := client.NewStdioMCPClient("azd", nil, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to start MCP client for %s: %w", ext.ID, err)
	}

	if _, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP client for %s: %w", ext.ID, err)
	}

	toolClientCache[ext.ID] = mcpClient
	return mcpClient, nil
}

// Helper to find extension metadata by tool name
func findExtensionByToolName(extList []mcpExtensionMetadata, toolName string) *mcpExtensionMetadata {
	for i := range extList {
		if extList[i].ID == "mcp."+toolName {
			return &extList[i]
		}
	}
	return nil
}

func loadMcpChildToolsMetadata(ctx context.Context) ([]mcp.Tool, []mcpExtensionMetadata, error) {
	extCmd := exec.Command("azd", "ext", "list", "--tags", "azure,mcp", "--output", "json")
	extOut, err := extCmd.CombinedOutput()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get extension metadata: %w\n%s", err, string(extOut))
	}

	var extList []mcpExtensionMetadata
	if err := json.Unmarshal(extOut, &extList); err != nil {
		return nil, nil, fmt.Errorf("failed to parse extension metadata: %w", err)
	}

	childTools := []mcp.Tool{}
	for _, ext := range extList {
		if ext.ID == "mcp.azure" {
			continue // skip self
		}
		toolName := ext.ID[len("mcp."):]
		tool := mcp.NewTool(toolName,
			mcp.WithDescription(ext.Description),
		)
		childTools = append(childTools, tool)
	}

	return childTools, extList, nil
}
