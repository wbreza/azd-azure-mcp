using System.Text.Json;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;
using Spectre.Console;
using Json.Schema;
using Json.More;
using ModelContextProtocol.Client;
using ModelContextProtocol;

namespace Azure.Mcp.Tools;

[McpServerToolType]
public class AzureTool : McpServerTool
{
    public override Tool ProtocolTool => new Tool
    {
        Name = "azure",
        Description = """
            This server/tool provides real-time, programmatic access to all Azure products, services, and resources,
            as well as all interactions with the Azure Developer CLI (azd).
            Use this tool for any Azure control plane or data plane operation, including resource management and automation.
            To discover available capabilities, call the tool with the "learn" parameter to get a list of top-level tools.
            To explore further, set "learn" and specify a tool name to retrieve supported commands and their parameters.
            To execute an action, set the "tool", "command", and convert the users intent into the "parameters" based on the discovered schema.
            Always use this tool for any Azure or "azd" related operation requiring up-to-date, dynamic, and interactive capabilities.
        """,
        Annotations = new ToolAnnotations(),
        InputSchema = new JsonSchemaBuilder()
            .Type(SchemaValueType.Object)
            .Properties(
                ("intent", new JsonSchemaBuilder()
                    .Type(SchemaValueType.String)
                    .Description("The intent of the operation the user wants to perform against azure.")),
                ("tool", new JsonSchemaBuilder()
                    .Type(SchemaValueType.String)
                    .Description("The azure tool to use to execute the operation.")),
                ("command", new JsonSchemaBuilder()
                    .Type(SchemaValueType.String)
                    .Description("The command to execute against the specified tool.")),
                ("parameters", new JsonSchemaBuilder()
                    .Type(SchemaValueType.Object)
                    .Description("The parameters to pass to the tool command.")),
                ("learn", new JsonSchemaBuilder()
                    .Type(SchemaValueType.Boolean)
                    .Description("To learn about the tool and its supported child tools and parameters.")
                    .Default(false))
            )
            .AdditionalProperties(false)
            .Build()
            .ToJsonDocument()
            .RootElement
    };

    private readonly Metadata.McpClientProviderLoader _providerLoader = new Metadata.McpClientProviderLoader();

    public override async ValueTask<CallToolResponse> InvokeAsync(RequestContext<CallToolRequestParams> request, CancellationToken cancellationToken = default)
    {
        var args = request.Params?.Arguments;
        bool learn = false;
        string? tool = null;
        string? command = null;

        if (args != null)
        {
            if (args.TryGetValue("learn", out var learnElem) && learnElem.ValueKind == JsonValueKind.True)
            {
                learn = true;
            }
            if (args.TryGetValue("tool", out var toolElem) && toolElem.ValueKind == JsonValueKind.String)
            {
                tool = toolElem.GetString();
            }
            if (args.TryGetValue("command", out var commandElem) && commandElem.ValueKind == JsonValueKind.String)
            {
                command = commandElem.GetString();
            }
        }

        if (learn && string.IsNullOrEmpty(tool) && string.IsNullOrEmpty(command))
        {
            return await RootLearnModeAsync(request, cancellationToken);
        }
        else if (learn && !string.IsNullOrEmpty(tool) && string.IsNullOrEmpty(command))
        {
            return await ToolLearnModeAsync(request, tool!, cancellationToken);
        }
        else if (!learn && !string.IsNullOrEmpty(tool) && !string.IsNullOrEmpty(command))
        {
            return await CommandModeAsync(request, tool!, command!, cancellationToken);
        }

        return new CallToolResponse
        {
            Content =
            [
                new Content {
                    Type = "text",
                    Text = """
                        The "tool" and "command" parameters are required when not learning
                        Run again with the "learn" argument to get a list of available tools and their parameters.
                        To learn about a specific tool, use the "tool" argument with the name of the tool.
                    """
                }
            ]
        };
    }

    private async Task<CallToolResponse> RootLearnModeAsync(RequestContext<CallToolRequestParams> request, CancellationToken cancellationToken)
    {
        var providerMetadataList = await _providerLoader.ListProviderMetadataAsync();
        var tools = new List<Tool>();
        foreach (var meta in providerMetadataList)
        {
            tools.Add(new Tool
            {
                Name = meta.Id,
                Description = meta.Description,
            });
        }
        var toolsResult = new ListToolsResult
        {
            Tools = tools,
        };

        var toolsJson = JsonSerializer.Serialize(toolsResult, new JsonSerializerOptions { WriteIndented = true });

        return new CallToolResponse
        {
            Content =
            [
                new Content {
                    Type = "text",
                    Text = $"""
                        Here are the available list of tools.
                        Next, identify the tool you want to learn about and run again with the "learn" argument and the "tool" name to get a list of available commands and their parameters.

                        {toolsJson}
                    """
                }
            ]
        };
    }

    private async Task<CallToolResponse> ToolLearnModeAsync(RequestContext<CallToolRequestParams> request, string tool, CancellationToken cancellationToken)
    {
        var client = await _providerLoader.GetProviderClientAsync(tool);
        if (client == null)
        {
            return new CallToolResponse
            {
                Content =
                [
                    new Content {
                        Type = "text",
                        Text = $"""
                            Tool '{tool}' not found.
                            Run again with the "learn" argument and empty "tool" to get a list of available tools and their parameters.
                        """
                    }
                ]
            };
        }

        var listToolsResult = await client.ListToolsAsync();
        var toolsJson = JsonSerializer.Serialize(listToolsResult, new JsonSerializerOptions { WriteIndented = true });

        return new CallToolResponse
        {
            Content =
            [
                new Content {
                    Type = "text",
                    Text = $"""
                        Here are the available command and their parameters for '{tool}' tool.
                        If you do not find a suitable tool, run again with the "learn" argument and empty "tool" to get a list of available tools and their parameters.
                        Next, identify the command you want to execute and run again with the "tool", "command", and "parameters" arguments.

                        {toolsJson}
                    """
                }
            ]
        };
    }

    private async Task<CallToolResponse> CommandModeAsync(RequestContext<CallToolRequestParams> request, string tool, string command, CancellationToken cancellationToken)
    {
        var args = request.Params?.Arguments;
        JsonElement parameters = default;
        if (args != null)
        {
            if (args.TryGetValue("parameters", out var paramElem))
            {
                parameters = paramElem;
            }
        }

        IMcpClient? client;

        try
        {
            client = await _providerLoader.GetProviderClientAsync(tool);
            if (client == null)
            {
                return new CallToolResponse
                {
                    Content =
                    [
                        new Content {
                            Type = "text",
                            Text = $"""
                                Tool '{tool}' not found.
                                Run again with the "learn" argument and empty "tool" to get a list of available tools and their parameters.
                            """
                        }
                    ]
                };
            }
        }
        catch (Exception ex)
        {
            return new CallToolResponse
            {
                Content =
                [
                    new Content {
                        Type = "text",
                        Text = $"""
                            There was an error connecting to the tool.
                            Failed to get tool: {tool}
                            Error: {ex.Message}
                        """
                    }
                ]
            };
        }

        var toolArgs = new Dictionary<string, object?>();
        if (parameters.ValueKind == JsonValueKind.Object)
        {
            toolArgs = JsonSerializer.Deserialize<Dictionary<string, object?>>(parameters.GetRawText()) ?? new();
        }

        try
        {
            return await client.CallToolAsync(command, toolArgs, cancellationToken: cancellationToken);
        }
        catch (Exception ex)
        {
            return new CallToolResponse
            {
                Content =
                [
                    new Content {
                        Type = "text",
                        Text = $"""
                            There was an error finding or calling tool and command.
                            Failed to call tool: {tool}, command: {command}
                            Error: {ex.Message}

                            Run again with the "learn" argument and the "tool" name to get a list of available tools and their parameters.
                        """
                    }
                ]
            };
        }
    }
}