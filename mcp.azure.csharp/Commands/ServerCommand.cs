using System.CommandLine;
using System.ComponentModel;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

public class ServerCommand
{
    public Command Build()
    {
        var serverCommand = new Command("server", "Manage the MCP server operations.");
        var startCommand = new Command("start", "Start the MCP server.");
        startCommand.SetHandler(this.Run);

        serverCommand.AddCommand(startCommand);
        return serverCommand;
    }

    public async Task Run()
    {
        var builder = Host.CreateApplicationBuilder();
        builder.Logging.AddConsole(consoleLogOptions =>
        {
            // Configure all logs to go to stderr
            consoleLogOptions.LogToStandardErrorThreshold = LogLevel.Trace;
        });

        builder.Services
            .AddMcpServer((options) =>
            {
                options.ServerInfo = new Implementation
                {
                    Name = "Azure",
                    Version = "1.0.0"
                };
                options.ServerInstructions = """
                    This server/tool provides real-time, programmatic access to all Azure products, services, and resources,
                    as well as all interactions with the Azure Developer CLI (azd).
                    Use this tool for any Azure control plane or data plane operation, including resource management and automation.
                    To discover available capabilities, call the tool with the "learn" parameter to get a list of top-level tools.
                    To explore further, set "learn" and specify a tool name to retrieve supported commands and their parameters.
                    To execute an action, set the "tool", "command", and convert the users intent into the "parameters" based on the discovered schema.
                    Always use this tool for any Azure or "azd" related operation requiring up-to-date, dynamic, and interactive capabilities.
                """;
            })
            .WithStdioServerTransport()
            .WithTools([new Azure.Mcp.Tools.AzureTool()]);

        await builder.Build().RunAsync();
    }
}