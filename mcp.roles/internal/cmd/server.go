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
				"Azure Role Assignments",
				"1.0.0",
				server.WithToolCapabilities(false),
				server.WithRecovery(),
				server.WithLogging(),
				server.WithInstructions("Supports tool automation for Azure role assignments, including listing, creating, and deleting role assignments and definitions."),
			)

			// az role assignment list
			roleAssignmentListTool := mcp.NewTool(
				"role-assignment-list",
				mcp.WithDescription("List role assignments."),
			)
			s.AddTool(roleAssignmentListTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				azCmd := exec.Command("az", "role", "assignment", "list")
				return runAzCommandWithResult(azCmd), nil
			})

			// az role assignment create
			roleAssignmentCreateTool := mcp.NewTool(
				"role-assignment-create",
				mcp.WithDescription("Create a new role assignment."),
				mcp.WithString("assignee", mcp.Required(), mcp.Description("The assignee principal (user, group, or service principal).")),
				mcp.WithString("role", mcp.Required(), mcp.Description("The role name or id.")),
				mcp.WithString("scope", mcp.Description("The scope at which the role assignment applies.")),
			)
			s.AddTool(roleAssignmentCreateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				args := []string{"role", "assignment", "create"}
				if assignee, ok := request.GetArguments()["assignee"].(string); ok && assignee != "" {
					args = append(args, "--assignee", assignee)
				}
				if role, ok := request.GetArguments()["role"].(string); ok && role != "" {
					args = append(args, "--role", role)
				}
				if scope, ok := request.GetArguments()["scope"].(string); ok && scope != "" {
					args = append(args, "--scope", scope)
				}
				azCmd := exec.Command("az", args...)
				return runAzCommandWithResult(azCmd), nil
			})

			// az role assignment delete
			roleAssignmentDeleteTool := mcp.NewTool(
				"role-assignment-delete",
				mcp.WithDescription("Delete a role assignment."),
				mcp.WithString("assignee", mcp.Description("The assignee principal (user, group, or service principal).")),
				mcp.WithString("role", mcp.Description("The role name or id.")),
				mcp.WithString("scope", mcp.Description("The scope at which the role assignment applies.")),
			)
			s.AddTool(roleAssignmentDeleteTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				args := []string{"role", "assignment", "delete"}
				if assignee, ok := request.GetArguments()["assignee"].(string); ok && assignee != "" {
					args = append(args, "--assignee", assignee)
				}
				if role, ok := request.GetArguments()["role"].(string); ok && role != "" {
					args = append(args, "--role", role)
				}
				if scope, ok := request.GetArguments()["scope"].(string); ok && scope != "" {
					args = append(args, "--scope", scope)
				}
				azCmd := exec.Command("az", args...)
				return runAzCommandWithResult(azCmd), nil
			})

			// az role definition list
			roleDefinitionListTool := mcp.NewTool(
				"role-definition-list",
				mcp.WithDescription("List custom and built-in role definitions."),
			)
			s.AddTool(roleDefinitionListTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				azCmd := exec.Command("az", "role", "definition", "list")
				return runAzCommandWithResult(azCmd), nil
			})

			// az role definition create
			roleDefinitionCreateTool := mcp.NewTool(
				"role-definition-create",
				mcp.WithDescription("Create a custom role definition."),
				mcp.WithString("roleDefinition", mcp.Required(), mcp.Description("The role definition JSON or file path.")),
			)
			s.AddTool(roleDefinitionCreateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				roleDef, _ := request.GetArguments()["roleDefinition"].(string)
				azCmd := exec.Command("az", "role", "definition", "create", "--role-definition", roleDef)
				return runAzCommandWithResult(azCmd), nil
			})

			// az role definition update
			roleDefinitionUpdateTool := mcp.NewTool(
				"role-definition-update",
				mcp.WithDescription("Update a custom role definition."),
				mcp.WithString("roleDefinition", mcp.Required(), mcp.Description("The role definition JSON or file path.")),
			)
			s.AddTool(roleDefinitionUpdateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				roleDef, _ := request.GetArguments()["roleDefinition"].(string)
				azCmd := exec.Command("az", "role", "definition", "update", "--role-definition", roleDef)
				return runAzCommandWithResult(azCmd), nil
			})

			// az role definition delete
			roleDefinitionDeleteTool := mcp.NewTool(
				"role-definition-delete",
				mcp.WithDescription("Delete a custom role definition."),
				mcp.WithString("name", mcp.Required(), mcp.Description("The name or id of the role definition.")),
			)
			s.AddTool(roleDefinitionDeleteTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				name, _ := request.GetArguments()["name"].(string)
				azCmd := exec.Command("az", "role", "definition", "delete", "--name", name)
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

func runAzCommandWithResult(cmd *exec.Cmd) *mcp.CallToolResult {
	output, err := cmd.CombinedOutput()
	result := string(output)
	if err != nil {
		result = result + "\n[error] " + err.Error()
	}
	return mcp.NewToolResultText(result)
}
