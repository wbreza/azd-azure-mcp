using ModelContextProtocol.Client;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Azure.Mcp.Metadata;

public interface IMcpClientProvider
{
    ClientMetadata CreateMetadata();
    Task<IMcpClient> CreateClientAsync();
}

public class ClientMetadata
{
    public string Id { get; set; } = string.Empty;
    public string Name { get; set; } = string.Empty;
    public string Description { get; set; } = string.Empty;
}

public class McpClientProviderLoader : IDisposable
{
    private readonly Dictionary<string, IMcpClientProvider> _providerMap = new(StringComparer.OrdinalIgnoreCase);
    private readonly Dictionary<string, IMcpClient> _clientCache = new(StringComparer.OrdinalIgnoreCase);
    private bool _initialized = false;
    private bool _disposed = false;

    private async Task EnsureInitializedAsync()
    {
        if (_initialized)
        {
            return;
        }
        var azd = await AzdExtensionLoader.ListAzdExtensionsAsync();
        var remote = await RemoteMcpServerLoader.LoadRemoteMcpServersAsync();
        foreach (var provider in azd)
        {
            var meta = provider.CreateMetadata();
            _providerMap[meta.Id] = provider;
        }
        foreach (var provider in remote)
        {
            var meta = provider.CreateMetadata();
            _providerMap[meta.Id] = provider;
        }
        _initialized = true;
    }

    public async Task<List<ClientMetadata>> ListProviderMetadataAsync()
    {
        await EnsureInitializedAsync();
        var result = new List<ClientMetadata>();
        foreach (var provider in _providerMap.Values)
        {
            result.Add(provider.CreateMetadata());
        }
        return result;
    }

    public async Task<IMcpClient?> GetProviderClientAsync(string name)
    {
        await EnsureInitializedAsync();
        if (_clientCache.TryGetValue(name, out var cached))
        {
            return cached;
        }
        if (_providerMap.TryGetValue(name, out var provider))
        {
            var client = await provider.CreateClientAsync();
            _clientCache[name] = client;
            return client;
        }
        return null;
    }

    public void Dispose()
    {
        if (_disposed) return;
        foreach (var client in _clientCache.Values)
        {
            if (client is IDisposable d)
            {
                d.Dispose();
            }
        }
        _disposed = true;
    }
}