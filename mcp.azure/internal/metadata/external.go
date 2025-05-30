package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"mcp.azure/resources"
)

// ExternalToolMetadata implements ToolMetadata for mcp.json tools.
type ExternalToolMetadata struct {
	tool mcpJsonTool
}

// mcpJsonTool holds external tool metadata fields.
type mcpJsonTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

type mcpJson struct {
	Servers []mcpJsonTool `json:"servers"`
}

func (j *ExternalToolMetadata) Tool() mcp.Tool {
	return mcp.NewTool(j.tool.Name, mcp.WithDescription(j.tool.Description))
}

func (j *ExternalToolMetadata) CreateClient(ctx context.Context) (*client.Client, error) {
	endpoint := j.tool.URL
	if endpoint == "" {
		return nil, fmt.Errorf("missing 'url' property for tool %s in mcp.json", j.tool.Name)
	}
	streamingClient, err := client.NewStreamableHttpClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to start streaming MCP client for %s: %w", j.tool.Name, err)
	}
	if err := streamingClient.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start streaming MCP client for %s: %w", j.tool.Name, err)
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "mcp.azure",
		Version: "1.0.0",
	}

	if _, err := streamingClient.Initialize(ctx, initRequest); err != nil {
		return nil, fmt.Errorf("failed to initialize streaming MCP client for %s: %w", j.tool.Name, err)
	}
	return streamingClient, nil
}

// Loads external (mcp.json) tools as ToolMetadata.
func LoadExternalToolMetadata(ctx context.Context) ([]ToolMetadata, error) {
	data := resources.McpJson
	if len(data) == 0 {
		return nil, nil // not an error if missing
	}
	var meta mcpJson
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse mcp.json: %w", err)
	}

	var result []ToolMetadata
	for _, jt := range meta.Servers {
		result = append(result, &ExternalToolMetadata{tool: jt})
	}
	return result, nil
}
