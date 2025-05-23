// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package cmd

import (
	// added for MCP server functionality
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

func newServerCommand() *cobra.Command {
	serverGroup := &cobra.Command{
		Use: "server",
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Get the context of the AZD project & environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			mcpServer := server.NewMCPServer("azd", "0.0.1",
				server.WithInstructions(
					"Provides tools to dynamically run the AZD (Azure Developer CLI) commands. "+
						"If a tool accepts a 'cwd', send the current working directory as the 'cwd' argument.",
				),
				server.WithLogging(),
			)

			registerTools(mcpServer)

			log.Println("Starting MCP server...")

			if err := server.ServeStdio(mcpServer); err != nil {
				return err
			}

			return nil
		},
	}

	serverGroup.AddCommand(startCmd)
	return serverGroup
}

func registerTools(s *server.MCPServer) {
	initTool := mcp.NewTool("init",
		mcp.WithDescription("Initializes a new azd project"),
		mcp.WithString("subscription",
			mcp.Description("The Azure subscription ID to use for provisioning "+
				"and deployment. This needs to be in UUID format. Use 'list-subscriptions' to get the list of subscriptions.")),
		mcp.WithString("location",
			mcp.Description("The primary Azure location to use for the infrastructure. This needs to be a valid Azure location. Use 'list-locations' to get the list of locations.")),
		mcp.WithString("template", mcp.Description("The azd template or git repository to use")),
		mcp.WithString("cwd",
			mcp.Description("The azd project directory"),
			mcp.Required(),
			mcp.DefaultString("."),
		),
		mcp.WithString("environment", mcp.Description("The azd environment to use")),
	)

	provisionTool := mcp.NewTool("provision",
		mcp.WithDescription(
			"Provisions infrastructure the resources for the azd project. "+
				"If the environment does not contain a location and subscription, set those first.",
		),
		mcp.WithBoolean("preview",
			mcp.Description("When preview is enabled azd will show the changes that will be made to the infrastructure but not actually apply them."),
			mcp.DefaultBool(false),
		),
		mcp.WithBoolean("skipState",
			mcp.Description("When skipState is enabled azd will not compare the current state of the infrastructure before provisioning."),
			mcp.DefaultBool(false),
		),
		mcp.WithString("cwd",
			mcp.Description("The azd project directory"),
			mcp.Required(),
			mcp.DefaultString("."),
		),
		mcp.WithString("environment", mcp.Description("The azd environment to use")),
	)

	envListTool := mcp.NewTool("list-environments",
		mcp.WithDescription("Lists the azd environments available for the current azd project."),
		mcp.WithString("cwd",
			mcp.Description("The azd project directory"),
			mcp.Required(),
			mcp.DefaultString("."),
		),
		mcp.WithString("environment", mcp.Description("The azd environment to use")),
	)

	newEnvTool := mcp.NewTool("create-environment",
		mcp.WithDescription("Creates a new azd environment"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The name of the azd environment to create")),
		mcp.WithString("cwd",
			mcp.Description("The azd project directory"),
			mcp.Required(),
			mcp.DefaultString("."),
		),
	)

	envGetValuesTool := mcp.NewTool("get-environment-values",
		mcp.WithDescription("Gets all the values of the current azd environment"),
		mcp.WithString("cwd",
			mcp.Description("The azd project directory"),
			mcp.Required(),
			mcp.DefaultString("."),
		),
		mcp.WithString("environment", mcp.Description("The azd environment to use")),
	)

	envSetTool := mcp.NewTool("set-environment-value",
		mcp.WithDescription("Sets a key value pair for the current azd environment."),
		mcp.WithString("cwd",
			mcp.Description("The azd project directory"),
			mcp.Required(),
			mcp.DefaultString("."),
		),
		mcp.WithString("environment", mcp.Description("The azd environment to use")),
		mcp.WithString("value",
			mcp.Required(),
			mcp.Description("The value of the azd environment to set"),
		),
		mcp.WithString("key",
			mcp.Required(),
			mcp.Description("The key of the azd environment to set"),
		),
	)

	deployTool := mcp.NewTool(
		"deploy",
		mcp.WithDescription(
			"Deploys the azd project. If the project was not provisioned, provision will need to happen first.",
		),
		mcp.WithString("cwd",
			mcp.Description("The azd project directory"),
			mcp.Required(),
			mcp.DefaultString("."),
		),
		mcp.WithString("environment", mcp.Description("The azd environment to use")),
	)

	showTool := mcp.NewTool("show",
		mcp.WithDescription("Shows the azd project configuration"),
		mcp.WithString("cwd",
			mcp.Description("The azd project directory"),
			mcp.Required(),
			mcp.DefaultString("."),
		),
		mcp.WithString("environment", mcp.Description("The azd environment to use")),
	)

	configShowTool := mcp.NewTool("global-config",
		mcp.WithDescription("Shows the current azd global / user configuration"),
	)

	templateListTool := mcp.NewTool("template-list",
		mcp.WithDescription("Find and lists all the available azd templates from Awesome AZD gallery. "+
			"The template list includes tags about the programming language, frameworks, and azure resources that are used."),
	)

	authLoginTool := mcp.NewTool(
		"auth-login",
		mcp.WithDescription(
			"Logs the user into azure using the azd CLI. This will open a browser window to authenticate the user.",
		),
		mcp.WithString("tenantId", mcp.Description("The Azure tenant ID to use for authentication.")),
	)

	authCheckStatusTool := mcp.NewTool("auth-check-status",
		mcp.WithDescription("Checks the status of the azd authentication. This will return a success or failure message."),
	)

	pipelineConfigTool := mcp.NewTool("pipeline-config",
		mcp.WithDescription("Configures the deployment pipeline for the AZD project to "+
			"connect securely to Azure."),
		mcp.WithString("provider",
			mcp.Description("The pipeline provider to use (github for Github Actions and azdo for Azure Pipelines).")),
		mcp.WithString("applicationServiceManagementReference",
			mcp.Description("Service Management Reference. This value must be a UUID.")),
		mcp.WithString("authType",
			mcp.Description("The authentication type used between the pipeline provider and Azure "+
				"(valid values: federated, client-credentials).")),
		mcp.WithString("principalId",
			mcp.Description("The client id of the service principal to use.")),
		mcp.WithString("principalName",
			mcp.Description("The name of the service principal to use.")),
		mcp.WithString("principalRole",
			mcp.Description("The roles to assign to the service principal. Defaults to "+
				"Contributor and User Access Administrator.")),
		mcp.WithString("remoteName",
			mcp.Description("The name of the git remote to configure the pipeline.")),
		mcp.WithString("cwd",
			mcp.Description("The azd project directory"),
			mcp.Required(),
			mcp.DefaultString("."),
		),
		mcp.WithString("environment", mcp.Description("The azd environment to use")),
	)

	upTool := mcp.NewTool("up",
		mcp.WithDescription("Runs a workflow to package, provision and deploy your application in a single step"),
		mcp.WithString("cwd",
			mcp.Description("The azd project directory"),
			mcp.Required(),
			mcp.DefaultString("."),
		),
		mcp.WithString("environment", mcp.Description("The azd environment to use")),
	)

	aiBuilderTool := mcp.NewTool("ai-builder",
		mcp.WithDescription("Guides when they need helping add AI capabilities to their project or application. Only add payload when you have all the answers."),
		mcp.WithString("payload", mcp.Description("The JSON payload of the collected questions and answers.")),
	)

	s.AddTool(initTool, invokeInit)
	s.AddTool(showTool, invokeShow)
	s.AddTool(provisionTool, invokeProvision)
	s.AddTool(deployTool, invokeDeploy)
	s.AddTool(configShowTool, invokeGlobalConfig)
	s.AddTool(envListTool, invokeEnvList)
	s.AddTool(newEnvTool, invokeNewEnv)
	s.AddTool(envGetValuesTool, invokeGetEnvValues)
	s.AddTool(envSetTool, invokeSetEnvValue)
	s.AddTool(templateListTool, invokeTemplateList)
	s.AddTool(authLoginTool, invokeAuthLogin)
	s.AddTool(authCheckStatusTool, invokeAuthCheckStatus)
	s.AddTool(pipelineConfigTool, invokePipelineConfig)
	s.AddTool(upTool, invokeUp)
	s.AddTool(aiBuilderTool, invokeAiBuilder)
}

type location struct {
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
}

type question struct {
	Id        string   `json:"id"`
	Question  string   `json:"question"`
	Choices   []string `json:"choices"`
	Answer    string   `json:"answers"`
	Condition string   `json:"condition"`
}

func invokeAiBuilder(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result := &mcp.CallToolResult{
		Content: []mcp.Content{},
	}

	questions := []question{
		{
			Id:       "scenario",
			Question: "What type of AI scenario are you building?",
			Choices: []string{
				"Regenerative Augmented Retrieval (RAG)",
				"AI Agent",
				"Other",
			},
			Answer: "[pending]",
		},
		{
			Id:       "require-custom-data",
			Question: "Does your project require custom data?",
			Choices: []string{
				"Yes",
				"No",
			},
			Answer: "[pending]",
		},
		{
			Id:       "custom-data-types",
			Question: "What type of data does your application need?",
			Choices: []string{
				"Structured data",
				"Unstructured data",
				"Images",
				"Audio",
				"Video",
				"Other",
			},
			Condition: "Only ask this question when the user selects 'Yes' to the question 'require-custom-data'.",
			Answer:    "[pending]",
		},
		{
			Id:       "custom-data-locations",
			Question: "Where is the custom data located?",
			Choices: []string{
				"Azure Blob Storage",
				"Azure Database",
				"Local file system",
				"Other",
			},
			Condition: "Only ask this question when the user selects 'Yes' to the question 'require-custom-data'.",
			Answer:    "[pending]",
		},
	}

	payload, hasPayload := request.GetArguments()["payload"]
	if !hasPayload {
		questionBytes, err := json.Marshal(questions)
		if err != nil {
			return nil, err
		}

		result.Content = append(result.Content,
			mcp.NewTextContent("Ask the user about the following questions one at a time:"),
			mcp.NewTextContent(string(questionBytes)),
			mcp.NewTextContent("After you have collected all the answers, call the `ai-builder` tool again with the questions and answers payload in the same JSON structure."),
		)
	} else {
		jsonPayload := payload.(string)

		log.Printf("AI Builder payload: \n%s\n", jsonPayload)

		if strings.Contains(jsonPayload, "[pending]") {
			result.Content = append(result.Content,
				mcp.NewTextContent("Please ask the next pending question in the following JSON payload and then call the `ai-builder` tool again with the updated payload."),
				mcp.NewTextContent(jsonPayload),
			)
		} else {
			result.Content = append(result.Content,
				mcp.NewTextContent("Added AI capabilities to the project."),
				mcp.NewTextContent(jsonPayload),
				mcp.NewTextContent("Next set of steps would be to provision the project with the `provision` tool and then deploy the project with the `deploy` tool."),
			)
		}
	}

	return result, nil
}

type subscription struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func invokeAuthLogin(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"auth", "login"}
	tenantId, hasTenantId := request.GetArguments()["tenantId"]
	if hasTenantId {
		args = append(args, "--tenant-id", tenantId.(string))
	}

	return execAzdCommand(request, args)
}

func invokeAuthCheckStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"auth", "login", "--check-status"}
	return execAzdCommand(request, args)
}

func invokeTemplateList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"template", "list", "--output", "json"}

	return execAzdCommand(request, args)
}

func invokeGetEnvValues(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"env", "get-values"}
	return execAzdCommand(request, args)
}

func invokeSetEnvValue(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"env", "set"}

	key, hasKey := request.GetArguments()["key"]
	if hasKey {
		args = append(args, key.(string))
	}

	value, hasValue := request.GetArguments()["value"]
	if hasValue {
		args = append(args, value.(string))
	}

	return execAzdCommand(request, args)
}

func invokeInit(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"init"}

	location, hasLocation := request.GetArguments()["location"]
	if hasLocation {
		args = append(args, "--location", location.(string))
	}

	subscription, hasSubscription := request.GetArguments()["subscription"]
	if hasSubscription {
		args = append(args, "--subscription", subscription.(string))
	}

	template, hasTemplate := request.GetArguments()["template"]
	if hasTemplate {
		args = append(args, "--template", template.(string))
	}

	result, err := execAzdCommand(request, args)
	if err == nil {
		result.Content = append(
			result.Content,
			mcp.NewTextContent(
				"Next an azd environment will need to be created. Please prompt the user for an environment name and then call the `env-new` tool.",
			),
		)
	}

	return result, err
}

func invokeEnvList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"env", "list"}
	return execAzdCommand(request, args)
}

func invokeNewEnv(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"env", "new"}

	name, hasName := request.GetArguments()["name"]
	if hasName {
		args = append(args, name.(string))
	}

	result, err := execAzdCommand(request, args)
	if err == nil {
		result.Content = append(result.Content,
			mcp.NewTextContent(
				"Next we need to ensure the Azure location and subscription have been set. "+
					"You can check azd environment values with the `env-get-values` tool. "+
					"It will use the default values from the azd global configuration. "+
					"If they aren't found, prompt the user and set them with the `env-set` tool.",
			),
		)
	}

	return result, err
}

func invokeShow(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"show"}
	return execAzdCommand(request, args)
}

func invokeGlobalConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"config", "show"}
	return execAzdCommand(request, args)
}

func invokeProvision(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"provision"}

	preview, hasPreview := request.GetArguments()["preview"]
	if hasPreview && preview.(bool) {
		args = append(args, "--preview")
	}

	skipState, hasSkipState := request.GetArguments()["skipState"]
	if hasSkipState && skipState.(bool) {
		args = append(args, "--no-state")
	}

	result, err := execAzdCommand(request, args)
	if err == nil {
		result.Content = append(
			result.Content,
			mcp.NewTextContent(
				"If the user also wants to deploy the app code for the project you can use the `deploy` tool.",
			),
		)
	}

	return result, err
}

func invokeDeploy(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"deploy"}

	serviceName, hasServiceName := request.GetArguments()["serviceName"]
	if hasServiceName {
		args = append(args, serviceName.(string))
	}

	result, err := execAzdCommand(request, args)
	if err == nil {
		result.Content = append(result.Content,
			mcp.NewTextContent(
				"The user might want to setup a CI/CD pipeline for the project. "+
					"If so, they can run the `pipeline-config` tool.",
			),
		)
	}

	return result, err
}

func invokePipelineConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"pipeline", "config"}
	if v, ok := request.GetArguments()["provider"]; ok {
		args = append(args, "--provider", v.(string))
	}
	if v, ok := request.GetArguments()["applicationServiceManagementReference"]; ok {
		args = append(args, "-m", v.(string))
	}
	if v, ok := request.GetArguments()["authType"]; ok {
		args = append(args, "--auth-type", v.(string))
	}
	if v, ok := request.GetArguments()["principalId"]; ok {
		args = append(args, "--principal-id", v.(string))
	}
	if v, ok := request.GetArguments()["principalName"]; ok {
		args = append(args, "--principal-name", v.(string))
	}
	if v, ok := request.GetArguments()["principalRole"]; ok {
		args = append(args, "--principal-role", v.(string))
	}
	if v, ok := request.GetArguments()["remoteName"]; ok {
		args = append(args, "--remote-name", v.(string))
	}

	return execAzdCommand(request, args)
}

func invokeUp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"up"}
	return execAzdCommand(request, args)
}

func appendGlobalFlags(args []string, request mcp.CallToolRequest) []string {
	cwd, hasCwd := request.GetArguments()["cwd"]
	if hasCwd {
		args = append(args, "--cwd", fmt.Sprintf("%s", cwd.(string)))
	}

	environment, hasEnvironment := request.GetArguments()["environment"]
	if hasEnvironment {
		args = append(args, "-e", fmt.Sprintf("%s", environment.(string)))
	}

	if request.GetArguments()["debug"] != nil {
		args = append(args, "--debug")
	}

	args = append(args, "--no-prompt")

	return args
}

func execAzdCommand(request mcp.CallToolRequest, args []string) (*mcp.CallToolResult, error) {
	result := &mcp.CallToolResult{
		Content: []mcp.Content{},
	}

	args = appendGlobalFlags(args, request)

	log.Printf("Running command: azd %s\n",
		strings.Join(args, " "))
	resultBytes, err := exec.Command("azd", args...).CombinedOutput()
	if err != nil {
		azdOutput := string(resultBytes)
		log.Printf("Error executing azd command: %s\n", azdOutput)

		if strings.Contains(azdOutput, "ERROR: no project exists") {
			result.Content = append(
				result.Content,
				mcp.NewTextContent(
					"An azd project has not been initialized yet.  Run the `init` tool create create a new azd project.",
				),
			)
		}

		if strings.Contains(azdOutput, "ERROR: infrastructure has not been provisioned.") {
			result.Content = append(
				result.Content,
				mcp.NewTextContent(
					"The azd project has not been provisioned yet.  Run the `provision` tool to provision the azd project.",
				),
			)
		}

		if strings.Contains(azdOutput, "Enter a new environment name") {
			result.Content = append(
				result.Content,
				mcp.NewTextContent(
					"An azd environment has not been created yet. Prompt the user for an environment name, then run the `env-new` tool create create a new azd environment.",
				),
			)
		}

		if strings.Contains(azdOutput, "fetching current principal") || strings.Contains(azdOutput, "not logged in") {
			result.Content = append(
				result.Content,
				mcp.NewTextContent(
					"The user is not logged in yet or the Azure auth token was not found or expired. Run the `auth-login` tool to authenticate a new session with Azure.",
				),
			)
		}

		if strings.Contains(azdOutput, "no default response for prompt") {
			result.Content = append(
				result.Content,
				mcp.NewTextContent(
					"The tool requires user input.  Please prompt the user for input defined in the error message and then call the the same tool again.",
				),
			)
		}

		// If we matched on any known scenarios, we can return early
		if len(result.Content) > 0 {
			return result, nil
		}
	}

	result.Content = append(result.Content, mcp.NewTextContent(string(resultBytes)))

	return result, err
}
