package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// AzdToolMetadata implements ToolMetadata for azd extensions.
type AzdToolMetadata struct {
	extension mcpExtensionMetadata
}

// mcpExtensionMetadata holds azd extension metadata fields.
type mcpExtensionMetadata struct {
	ID            string   `json:"id"`
	Description   string   `json:"description"`
	Namespace     string   `json:"namespace"`
	Version       string   `json:"version"`
	LatestVersion string   `json:"latestVersion"`
	Installed     bool     `json:"installed"`
	Tags          []string `json:"tags"`
}

func (a *AzdToolMetadata) Tool() mcp.Tool {
	name := strings.TrimPrefix(a.extension.ID, "mcp.")
	return mcp.NewTool(name, mcp.WithDescription(a.extension.Description))
}

func (a *AzdToolMetadata) CreateClient(ctx context.Context) (*client.Client, error) {
	if a.extension.Installed {
		if a.extension.LatestVersion != a.extension.Version {
			currentVer, currentVerErr := semver.NewVersion(a.extension.Version)
			latestVer, latestVerErr := semver.NewVersion(a.extension.LatestVersion)
			if currentVerErr == nil && latestVerErr == nil && latestVer.GreaterThan(currentVer) {
				upgradeCmd := exec.Command("azd", "ext", "upgrade", a.extension.ID)
				upgradeOut, err := upgradeCmd.CombinedOutput()
				if err != nil {
					return nil, fmt.Errorf("failed to upgrade extension %s: %w\n%s", a.extension.ID, err, string(upgradeOut))
				}
			}
		}
	} else {
		installCmd := exec.Command("azd", "ext", "install", a.extension.ID)
		installOut, err := installCmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to install extension %s: %w\n%s", a.extension.ID, err, string(installOut))
		}
	}

	nsParts := strings.Split(a.extension.Namespace, ".")
	if len(nsParts) < 2 {
		return nil, fmt.Errorf("invalid namespace for extension: %s", a.extension.Namespace)
	}
	args := append([]string{}, nsParts...)
	args = append(args, "server", "start")

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "mcp.azure",
		Version: "1.0.0",
	}

	mcpClient, err := client.NewStdioMCPClient("azd", nil, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to start MCP client for %s: %w", a.extension.ID, err)
	}
	if _, err := mcpClient.Initialize(ctx, initRequest); err != nil {
		return nil, fmt.Errorf("failed to initialize Stdio MCP client for %s: %w", a.extension.ID, err)
	}
	return mcpClient, nil
}

// Loads azd extension tools as ToolMetadata.
func LoadAzdToolMetadata(ctx context.Context) ([]ToolMetadata, error) {
	extCmd := exec.Command("azd", "ext", "list", "--tags", "azure,mcp", "--output", "json")
	extOut, err := extCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get extension metadata: %w\n%s", err, string(extOut))
	}
	var extList []mcpExtensionMetadata
	if err := json.Unmarshal(extOut, &extList); err != nil {
		return nil, fmt.Errorf("failed to parse extension metadata: %w", err)
	}

	var result []ToolMetadata
	for _, ext := range extList {
		if ext.ID == "mcp.azure" {
			continue // skip self
		}
		result = append(result, &AzdToolMetadata{extension: ext})
	}

	return result, nil
}
