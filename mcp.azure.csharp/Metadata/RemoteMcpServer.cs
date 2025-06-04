using System.Reflection;
using System.Text.Json;
using ModelContextProtocol.Client;
using ModelContextProtocol.Protocol;

namespace Azure.Mcp.Metadata;

public class RemoteMcpServer : IMcpClientProvider
{
    public string Id { get; set; } = string.Empty;
    public string Name { get; set; } = string.Empty;
    public string Url { get; set; } = string.Empty;
    public string Description { get; set; } = string.Empty;

    public ClientMetadata CreateMetadata() => new ClientMetadata
    {
        Id = Id,
        Name = Name,
        Description = Description
    };

    public async Task<IMcpClient> CreateClientAsync()
    {
        var transportOptions = new SseClientTransportOptions
        {
            Name = Name,
            TransportMode = HttpTransportMode.StreamableHttp,
            Endpoint = new Uri(Url),
        };

        var clientTransport = new SseClientTransport(transportOptions);
        var clientOptions = new McpClientOptions
        {
            ClientInfo = new Implementation
            {
                Name = "Azure",
                Version = "1.0.0",
            },
            ProtocolVersion = "2025-03-26",
        };

        var client = await McpClientFactory.CreateAsync(clientTransport, clientOptions);
        await client.PingAsync();

        return client;
    }
}

public class RemoteMcpServerList
{
    public List<RemoteMcpServer> Servers { get; set; } = new();
}

public static class RemoteMcpServerLoader
{
    public static async Task<List<RemoteMcpServer>> LoadRemoteMcpServersAsync()
    {
        var assembly = Assembly.GetExecutingAssembly();
        var resourceName = assembly.GetManifestResourceNames()
            .FirstOrDefault(n => n.EndsWith("mcp.json", StringComparison.OrdinalIgnoreCase));
        if (resourceName == null) return new List<RemoteMcpServer>();
        using var stream = assembly.GetManifestResourceStream(resourceName);
        if (stream == null) return new List<RemoteMcpServer>();
        using var reader = new StreamReader(stream);
        var json = await reader.ReadToEndAsync();
        try
        {
            var options = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
            var servers = JsonSerializer.Deserialize<RemoteMcpServerList>(json, options);
            return servers?.Servers ?? new List<RemoteMcpServer>();
        }
        catch { return new List<RemoteMcpServer>(); }
    }
}
