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
				"Azure Service Bus Namespaces",
				"1.0.0",
				server.WithToolCapabilities(true),
				server.WithRecovery(),
				server.WithLogging(),
				server.WithPromptCapabilities(false),
				server.WithResourceCapabilities(false, false),
				server.WithInstructions("Supports tools for interacting with Azure Service Bus namespaces."),
			)

			// Service Bus: Create Namespace
			createNamespaceTool := mcp.NewTool(
				"create-servicebus-namespace",
				mcp.WithDescription("Create a new Azure Service Bus namespace."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("location", mcp.Required(), mcp.Description("Azure region")),
			)
			s.AddTool(createNamespaceTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["name"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: name"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: name, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				locArg, ok := request.GetArguments()["location"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: location"), nil
				}
				loc, ok := locArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: location, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "namespace", "create", "--name", name, "--resource-group", rg, "--location", loc)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: List Namespaces
			listNamespacesTool := mcp.NewTool(
				"list-servicebus-namespaces",
				mcp.WithDescription("List Azure Service Bus namespaces in a resource group."),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
			)
			s.AddTool(listNamespacesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "namespace", "list", "--resource-group", rg)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Show Namespace
			showNamespaceTool := mcp.NewTool(
				"show-servicebus-namespace",
				mcp.WithDescription("Show details of a Service Bus namespace."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
			)
			s.AddTool(showNamespaceTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["name"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: name"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: name, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "namespace", "show", "--name", name, "--resource-group", rg)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Update Namespace
			updateNamespaceTool := mcp.NewTool(
				"update-servicebus-namespace",
				mcp.WithDescription("Update a Service Bus namespace."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("set", mcp.Description("Properties to set (key=value)")),
			)
			s.AddTool(updateNamespaceTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["name"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: name"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: name, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				setArg, _ := request.GetArguments()["set"]
				set, _ := setArg.(string)
				args := []string{"servicebus", "namespace", "update", "--name", name, "--resource-group", rg}
				if set != "" {
					args = append(args, "--set", set)
				}
				cmd := exec.Command("az", args...)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Delete Namespace
			deleteNamespaceTool := mcp.NewTool(
				"delete-servicebus-namespace",
				mcp.WithDescription("Delete a Service Bus namespace."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
			)
			s.AddTool(deleteNamespaceTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nameArg, ok := request.GetArguments()["name"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: name"), nil
				}
				name, ok := nameArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: name, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "namespace", "delete", "--name", name, "--resource-group", rg)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Create Queue
			createQueueTool := mcp.NewTool(
				"create-servicebus-queue",
				mcp.WithDescription("Create a new queue in a Service Bus namespace."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("queueName", mcp.Required(), mcp.Description("Queue name")),
			)
			s.AddTool(createQueueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				queueArg, ok := request.GetArguments()["queueName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: queueName"), nil
				}
				queue, ok := queueArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: queueName, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "queue", "create", "--namespace-name", ns, "--resource-group", rg, "--name", queue)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: List Queues
			listQueuesTool := mcp.NewTool(
				"list-servicebus-queues",
				mcp.WithDescription("List queues in a Service Bus namespace."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
			)
			s.AddTool(listQueuesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "queue", "list", "--namespace-name", ns, "--resource-group", rg)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Show Queue
			showQueueTool := mcp.NewTool(
				"show-servicebus-queue",
				mcp.WithDescription("Show details of a Service Bus queue."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("queueName", mcp.Required(), mcp.Description("Queue name")),
			)
			s.AddTool(showQueueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				queueArg, ok := request.GetArguments()["queueName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: queueName"), nil
				}
				queue, ok := queueArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: queueName, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "queue", "show", "--namespace-name", ns, "--resource-group", rg, "--name", queue)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Update Queue
			updateQueueTool := mcp.NewTool(
				"update-servicebus-queue",
				mcp.WithDescription("Update a Service Bus queue."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("queueName", mcp.Required(), mcp.Description("Queue name")),
				mcp.WithString("set", mcp.Description("Properties to set (key=value)")),
			)
			s.AddTool(updateQueueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				queueArg, ok := request.GetArguments()["queueName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: queueName"), nil
				}
				queue, ok := queueArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: queueName, expected string"), nil
				}
				setArg, _ := request.GetArguments()["set"]
				set, _ := setArg.(string)
				args := []string{"servicebus", "queue", "update", "--namespace-name", ns, "--resource-group", rg, "--name", queue}
				if set != "" {
					args = append(args, "--set", set)
				}
				cmd := exec.Command("az", args...)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Delete Queue
			deleteQueueTool := mcp.NewTool(
				"delete-servicebus-queue",
				mcp.WithDescription("Delete a Service Bus queue."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("queueName", mcp.Required(), mcp.Description("Queue name")),
			)
			s.AddTool(deleteQueueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				queueArg, ok := request.GetArguments()["queueName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: queueName"), nil
				}
				queue, ok := queueArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: queueName, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "queue", "delete", "--namespace-name", ns, "--resource-group", rg, "--name", queue)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Create Topic
			createTopicTool := mcp.NewTool(
				"create-servicebus-topic",
				mcp.WithDescription("Create a new topic in a Service Bus namespace."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("topicName", mcp.Required(), mcp.Description("Topic name")),
			)
			s.AddTool(createTopicTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				topicArg, ok := request.GetArguments()["topicName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: topicName"), nil
				}
				topic, ok := topicArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: topicName, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "topic", "create", "--namespace-name", ns, "--resource-group", rg, "--name", topic)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: List Topics
			listTopicsTool := mcp.NewTool(
				"list-servicebus-topics",
				mcp.WithDescription("List topics in a Service Bus namespace."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
			)
			s.AddTool(listTopicsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "topic", "list", "--namespace-name", ns, "--resource-group", rg)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Show Topic
			showTopicTool := mcp.NewTool(
				"show-servicebus-topic",
				mcp.WithDescription("Show details of a Service Bus topic."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("topicName", mcp.Required(), mcp.Description("Topic name")),
			)
			s.AddTool(showTopicTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				topicArg, ok := request.GetArguments()["topicName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: topicName"), nil
				}
				topic, ok := topicArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: topicName, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "topic", "show", "--namespace-name", ns, "--resource-group", rg, "--name", topic)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Update Topic
			updateTopicTool := mcp.NewTool(
				"update-servicebus-topic",
				mcp.WithDescription("Update a Service Bus topic."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("topicName", mcp.Required(), mcp.Description("Topic name")),
				mcp.WithString("set", mcp.Description("Properties to set (key=value)")),
			)
			s.AddTool(updateTopicTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				topicArg, ok := request.GetArguments()["topicName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: topicName"), nil
				}
				topic, ok := topicArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: topicName, expected string"), nil
				}
				setArg, _ := request.GetArguments()["set"]
				set, _ := setArg.(string)
				args := []string{"servicebus", "topic", "update", "--namespace-name", ns, "--resource-group", rg, "--name", topic}
				if set != "" {
					args = append(args, "--set", set)
				}
				cmd := exec.Command("az", args...)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Delete Topic
			deleteTopicTool := mcp.NewTool(
				"delete-servicebus-topic",
				mcp.WithDescription("Delete a Service Bus topic."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("topicName", mcp.Required(), mcp.Description("Topic name")),
			)
			s.AddTool(deleteTopicTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				topicArg, ok := request.GetArguments()["topicName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: topicName"), nil
				}
				topic, ok := topicArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: topicName, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "topic", "delete", "--namespace-name", ns, "--resource-group", rg, "--name", topic)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Create Subscription
			createSubscriptionTool := mcp.NewTool(
				"create-servicebus-subscription",
				mcp.WithDescription("Create a new subscription in a Service Bus topic."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("topicName", mcp.Required(), mcp.Description("Topic name")),
				mcp.WithString("subscriptionName", mcp.Required(), mcp.Description("Subscription name")),
			)
			s.AddTool(createSubscriptionTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				topicArg, ok := request.GetArguments()["topicName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: topicName"), nil
				}
				topic, ok := topicArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: topicName, expected string"), nil
				}
				subArg, ok := request.GetArguments()["subscriptionName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: subscriptionName"), nil
				}
				sub, ok := subArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: subscriptionName, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "topic", "subscription", "create", "--namespace-name", ns, "--resource-group", rg, "--topic-name", topic, "--name", sub)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: List Subscriptions
			listSubscriptionsTool := mcp.NewTool(
				"list-servicebus-subscriptions",
				mcp.WithDescription("List subscriptions in a Service Bus topic."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("topicName", mcp.Required(), mcp.Description("Topic name")),
			)
			s.AddTool(listSubscriptionsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				topicArg, ok := request.GetArguments()["topicName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: topicName"), nil
				}
				topic, ok := topicArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: topicName, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "topic", "subscription", "list", "--namespace-name", ns, "--resource-group", rg, "--topic-name", topic)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Show Subscription
			showSubscriptionTool := mcp.NewTool(
				"show-servicebus-subscription",
				mcp.WithDescription("Show details of a Service Bus subscription."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("topicName", mcp.Required(), mcp.Description("Topic name")),
				mcp.WithString("subscriptionName", mcp.Required(), mcp.Description("Subscription name")),
			)
			s.AddTool(showSubscriptionTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				topicArg, ok := request.GetArguments()["topicName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: topicName"), nil
				}
				topic, ok := topicArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: topicName, expected string"), nil
				}
				subArg, ok := request.GetArguments()["subscriptionName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: subscriptionName"), nil
				}
				sub, ok := subArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: subscriptionName, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "topic", "subscription", "show", "--namespace-name", ns, "--resource-group", rg, "--topic-name", topic, "--name", sub)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Update Subscription
			updateSubscriptionTool := mcp.NewTool(
				"update-servicebus-subscription",
				mcp.WithDescription("Update a Service Bus subscription."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("topicName", mcp.Required(), mcp.Description("Topic name")),
				mcp.WithString("subscriptionName", mcp.Required(), mcp.Description("Subscription name")),
				mcp.WithString("set", mcp.Description("Properties to set (key=value)")),
			)
			s.AddTool(updateSubscriptionTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				topicArg, ok := request.GetArguments()["topicName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: topicName"), nil
				}
				topic, ok := topicArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: topicName, expected string"), nil
				}
				subArg, ok := request.GetArguments()["subscriptionName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: subscriptionName"), nil
				}
				sub, ok := subArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: subscriptionName, expected string"), nil
				}
				setArg, _ := request.GetArguments()["set"]
				set, _ := setArg.(string)
				args := []string{"servicebus", "topic", "subscription", "update", "--namespace-name", ns, "--resource-group", rg, "--topic-name", topic, "--name", sub}
				if set != "" {
					args = append(args, "--set", set)
				}
				cmd := exec.Command("az", args...)
				return runAzCommandWithResult(cmd), nil
			})

			// Service Bus: Delete Subscription
			deleteSubscriptionTool := mcp.NewTool(
				"delete-servicebus-subscription",
				mcp.WithDescription("Delete a Service Bus subscription."),
				mcp.WithString("namespaceName", mcp.Required(), mcp.Description("Service Bus namespace name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("topicName", mcp.Required(), mcp.Description("Topic name")),
				mcp.WithString("subscriptionName", mcp.Required(), mcp.Description("Subscription name")),
			)
			s.AddTool(deleteSubscriptionTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				nsArg, ok := request.GetArguments()["namespaceName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: namespaceName"), nil
				}
				ns, ok := nsArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: namespaceName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				topicArg, ok := request.GetArguments()["topicName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: topicName"), nil
				}
				topic, ok := topicArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: topicName, expected string"), nil
				}
				subArg, ok := request.GetArguments()["subscriptionName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: subscriptionName"), nil
				}
				sub, ok := subArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: subscriptionName, expected string"), nil
				}
				cmd := exec.Command("az", "servicebus", "topic", "subscription", "delete", "--namespace-name", ns, "--resource-group", rg, "--topic-name", topic, "--name", sub)
				return runAzCommandWithResult(cmd), nil
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
