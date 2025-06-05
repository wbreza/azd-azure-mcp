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
				"Azure Cosmos Accounts",
				"1.0.0",
				server.WithToolCapabilities(true),
				server.WithRecovery(),
				server.WithLogging(),
				server.WithPromptCapabilities(false),
				server.WithResourceCapabilities(false, false),
				server.WithInstructions("Supports tools for interacting with Azure accounts, subscriptions and locations."),
			)

			// Register Cosmos DB tools using the mcp.NewTool and s.AddTool pattern, matching mcp.resource
			// Each tool is defined with mcp.NewTool (name, description, parameters) and registered with a handler that validates arguments and calls the az CLI

			// Cosmos DB: Create Account
			createCosmosAccountTool := mcp.NewTool(
				"create-cosmosdb-account",
				mcp.WithDescription("Create a new Azure Cosmos DB account."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
			)
			s.AddTool(createCosmosAccountTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				cmd := exec.Command("az", "cosmosdb", "create", "--name", name, "--resource-group", rg)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: List Accounts
			listCosmosAccountsTool := mcp.NewTool(
				"list-cosmosdb-accounts",
				mcp.WithDescription("List Cosmos DB accounts in a resource group."),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
			)
			s.AddTool(listCosmosAccountsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				cmd := exec.Command("az", "cosmosdb", "list", "--resource-group", rg)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: Show Account
			showCosmosAccountTool := mcp.NewTool(
				"show-cosmosdb-account",
				mcp.WithDescription("Show details of a Cosmos DB account."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
			)
			s.AddTool(showCosmosAccountTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				cmd := exec.Command("az", "cosmosdb", "show", "--name", name, "--resource-group", rg)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: Update Account
			updateCosmosAccountTool := mcp.NewTool(
				"update-cosmosdb-account",
				mcp.WithDescription("Update a Cosmos DB account."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("set", mcp.Description("Properties to set (key=value)")),
			)
			s.AddTool(updateCosmosAccountTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				args := []string{"cosmosdb", "update", "--name", name, "--resource-group", rg}
				if set != "" {
					args = append(args, "--set", set)
				}
				cmd := exec.Command("az", args...)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: Delete Account
			deleteCosmosAccountTool := mcp.NewTool(
				"delete-cosmosdb-account",
				mcp.WithDescription("Delete a Cosmos DB account."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
			)
			s.AddTool(deleteCosmosAccountTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				cmd := exec.Command("az", "cosmosdb", "delete", "--name", name, "--resource-group", rg)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: List SQL Databases
			listSqlDatabasesTool := mcp.NewTool(
				"list-cosmosdb-sql-databases",
				mcp.WithDescription("List SQL databases in a Cosmos DB account."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
			)
			s.AddTool(listSqlDatabasesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				cmd := exec.Command("az", "cosmosdb", "sql", "database", "list", "--account-name", name, "--resource-group", rg)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: List SQL Containers
			listSqlContainersTool := mcp.NewTool(
				"list-cosmosdb-sql-containers",
				mcp.WithDescription("List containers in a SQL database."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("databaseName", mcp.Required(), mcp.Description("SQL database name")),
			)
			s.AddTool(listSqlContainersTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				dbArg, ok := request.GetArguments()["databaseName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: databaseName"), nil
				}
				db, ok := dbArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: databaseName, expected string"), nil
				}
				cmd := exec.Command("az", "cosmosdb", "sql", "container", "list", "--account-name", name, "--resource-group", rg, "--database-name", db)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: List MongoDB Databases
			listMongoDatabasesTool := mcp.NewTool(
				"list-cosmosdb-mongodb-databases",
				mcp.WithDescription("List MongoDB databases in a Cosmos DB account."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
			)
			s.AddTool(listMongoDatabasesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				cmd := exec.Command("az", "cosmosdb", "mongodb", "database", "list", "--account-name", name, "--resource-group", rg)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: List MongoDB Collections
			listMongoCollectionsTool := mcp.NewTool(
				"list-cosmosdb-mongodb-collections",
				mcp.WithDescription("List MongoDB collections in a database."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("databaseName", mcp.Required(), mcp.Description("MongoDB database name")),
			)
			s.AddTool(listMongoCollectionsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				dbArg, ok := request.GetArguments()["databaseName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: databaseName"), nil
				}
				db, ok := dbArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: databaseName, expected string"), nil
				}
				cmd := exec.Command("az", "cosmosdb", "mongodb", "collection", "list", "--account-name", name, "--resource-group", rg, "--database-name", db)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: Create SQL Database
			createSqlDatabaseTool := mcp.NewTool(
				"create-cosmosdb-sql-database",
				mcp.WithDescription("Create a new SQL database in a Cosmos DB account."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("databaseName", mcp.Required(), mcp.Description("SQL database name")),
			)
			s.AddTool(createSqlDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				dbArg, ok := request.GetArguments()["databaseName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: databaseName"), nil
				}
				db, ok := dbArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: databaseName, expected string"), nil
				}
				cmd := exec.Command("az", "cosmosdb", "sql", "database", "create", "--account-name", name, "--resource-group", rg, "--name", db)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: Create MongoDB Database
			createMongoDatabaseTool := mcp.NewTool(
				"create-cosmosdb-mongodb-database",
				mcp.WithDescription("Create a new MongoDB database in a Cosmos DB account."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("databaseName", mcp.Required(), mcp.Description("MongoDB database name")),
			)
			s.AddTool(createMongoDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				dbArg, ok := request.GetArguments()["databaseName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: databaseName"), nil
				}
				db, ok := dbArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: databaseName, expected string"), nil
				}
				cmd := exec.Command("az", "cosmosdb", "mongodb", "database", "create", "--account-name", name, "--resource-group", rg, "--name", db)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: Create SQL Container
			createSqlContainerTool := mcp.NewTool(
				"create-cosmosdb-sql-container",
				mcp.WithDescription("Create a new container in a Cosmos DB SQL database."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("databaseName", mcp.Required(), mcp.Description("SQL database name")),
				mcp.WithString("containerName", mcp.Required(), mcp.Description("Container name")),
				mcp.WithString("partitionKeyPath", mcp.Required(), mcp.Description("Partition key path (e.g. /myPartitionKey)")),
			)
			s.AddTool(createSqlContainerTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				dbArg, ok := request.GetArguments()["databaseName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: databaseName"), nil
				}
				db, ok := dbArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: databaseName, expected string"), nil
				}
				containerArg, ok := request.GetArguments()["containerName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: containerName"), nil
				}
				container, ok := containerArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: containerName, expected string"), nil
				}
				pkArg, ok := request.GetArguments()["partitionKeyPath"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: partitionKeyPath"), nil
				}
				pk, ok := pkArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: partitionKeyPath, expected string"), nil
				}
				cmd := exec.Command("az", "cosmosdb", "sql", "container", "create", "--account-name", name, "--resource-group", rg, "--database-name", db, "--name", container, "--partition-key-path", pk)
				return runAzCommandWithResult(cmd), nil
			})

			// Cosmos DB: Create MongoDB Collection
			createMongoCollectionTool := mcp.NewTool(
				"create-cosmosdb-mongodb-collection",
				mcp.WithDescription("Create a new collection in a Cosmos DB MongoDB database."),
				mcp.WithString("accountName", mcp.Required(), mcp.Description("Cosmos DB account name")),
				mcp.WithString("resourceGroup", mcp.Required(), mcp.Description("Resource group name")),
				mcp.WithString("databaseName", mcp.Required(), mcp.Description("MongoDB database name")),
				mcp.WithString("collectionName", mcp.Required(), mcp.Description("Collection name")),
				mcp.WithString("shard", mcp.Required(), mcp.Description("Shard (partition key path, e.g. /myPartitionKey)")),
			)
			s.AddTool(createMongoCollectionTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				accountArg, ok := request.GetArguments()["accountName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: accountName"), nil
				}
				account, ok := accountArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: accountName, expected string"), nil
				}
				rgArg, ok := request.GetArguments()["resourceGroup"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: resourceGroup"), nil
				}
				rg, ok := rgArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: resourceGroup, expected string"), nil
				}
				dbArg, ok := request.GetArguments()["databaseName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: databaseName"), nil
				}
				db, ok := dbArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: databaseName, expected string"), nil
				}
				collArg, ok := request.GetArguments()["collectionName"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: collectionName"), nil
				}
				coll, ok := collArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: collectionName, expected string"), nil
				}
				shardArg, ok := request.GetArguments()["shard"]
				if !ok {
					return mcp.NewToolResultText("Missing required argument: shard"), nil
				}
				shard, ok := shardArg.(string)
				if !ok {
					return mcp.NewToolResultText("Invalid type for argument: shard, expected string"), nil
				}
				cmd := exec.Command("az", "cosmosdb", "mongodb", "collection", "create", "--account-name", account, "--resource-group", rg, "--database-name", db, "--name", coll, "--shard", shard)
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
