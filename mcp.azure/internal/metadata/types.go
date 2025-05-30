package metadata

import (
	"context"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// ToolMetadata provides a unified interface for tool metadata and client creation.
type ToolMetadata interface {
	Tool() mcp.Tool
	CreateClient(ctx context.Context) (*client.Client, error)
}
