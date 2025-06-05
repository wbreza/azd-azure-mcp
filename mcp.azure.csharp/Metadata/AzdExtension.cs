using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Text.Json;
using System.Threading.Tasks;
using ModelContextProtocol.Client;
using Semver;

namespace Azure.Mcp.Metadata;

public class AzdExtension : IMcpClientProvider
{
    public string Id { get; set; } = string.Empty;
    public string Name { get; set; } = string.Empty;
    public string Description { get; set; } = string.Empty;
    public string Namespace { get; set; } = string.Empty;
    public string Version { get; set; } = string.Empty;
    public string LatestVersion { get; set; } = string.Empty;
    public bool Installed { get; set; }
    public string Source { get; set; } = string.Empty;
    public List<string> Tags { get; set; } = new();

    public ClientMetadata CreateMetadata() => new ClientMetadata
    {
        Id = Id,
        Name = Name,
        Description = Description
    };

    private async Task InstallExtensionAsync()
    {
        var psi = new ProcessStartInfo
        {
            FileName = "azd",
            ArgumentList = { "ext", "install", Id },
            UseShellExecute = false,
            CreateNoWindow = true
        };
        using var proc = Process.Start(psi);
        if (proc != null)
        {
            await proc.WaitForExitAsync();
            if (proc.ExitCode != 0)
            {
                throw new Exception($"Failed to install azd extension: {Id}");
            }
        }
    }

    private async Task UpgradeExtensionAsync()
    {
        var psi = new ProcessStartInfo
        {
            FileName = "azd",
            ArgumentList = { "ext", "upgrade", Id },
            UseShellExecute = false,
            CreateNoWindow = true
        };
        using var proc = Process.Start(psi);
        if (proc != null)
        {
            await proc.WaitForExitAsync();
            if (proc.ExitCode != 0)
            {
                throw new Exception($"Failed to upgrade azd extension: {Id}");
            }
        }
    }

    private async Task EnsureExtensionIsInstalledAndUpToDateAsync()
    {
        if (!Installed)
        {
            await InstallExtensionAsync();
            Installed = true;
            Version = LatestVersion;
            return;
        }
        if (!string.IsNullOrEmpty(Version) && !string.IsNullOrEmpty(LatestVersion))
        {
            var current = SemVersion.Parse(Version, SemVersionStyles.Any);
            var latest = SemVersion.Parse(LatestVersion, SemVersionStyles.Any);
            if (current.ComparePrecedenceTo(latest) < 0)
            {
                await UpgradeExtensionAsync();
                Version = LatestVersion;
            }
        }
    }

    public async Task<IMcpClient> CreateClientAsync()
    {
        await EnsureExtensionIsInstalledAndUpToDateAsync();
        var nsParts = (Namespace ?? string.Empty).Split('.', StringSplitOptions.RemoveEmptyEntries);
        var arguments = new List<string>(nsParts) { "server", "start" };
        var transportOptions = new StdioClientTransportOptions
        {
            Name = Name,
            Command = "azd",
            Arguments = arguments,
        };
        var clientTransport = new StdioClientTransport(transportOptions);

        var clientOptions = new McpClientOptions { };
        return await McpClientFactory.CreateAsync(clientTransport, clientOptions);}
}

public static class AzdExtensionLoader
{
    private static readonly HashSet<string> IgnoreToolIds = new(StringComparer.OrdinalIgnoreCase)
    {
        "mcp.azure",
        "mcp.azure.csharp",
        "mcp.azure.ts"
    };

    public static async Task<List<AzdExtension>> ListAzdExtensionsAsync()
    {
        var result = new List<AzdExtension>();
        var psi = new ProcessStartInfo
        {
            FileName = "azd",
            ArgumentList = { "ext", "list", "--output", "json", "--tags", "mcp,azure" },
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            UseShellExecute = false,
            CreateNoWindow = true
        };
        using var proc = Process.Start(psi);
        if (proc == null) return result;
        var output = await proc.StandardOutput.ReadToEndAsync();
        await proc.WaitForExitAsync();
        if (proc.ExitCode != 0) return result;
        try
        {
            var options = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
            var arr = JsonSerializer.Deserialize<List<AzdExtension>>(output, options);
            if (arr != null)
            {
                foreach (var ext in arr)
                {
                    if (!IgnoreToolIds.Contains(ext.Id))
                    {
                        result.Add(ext);
                    }
                }
            }
        }
        catch { /* ignore parse errors */ }
        return result;
    }
}
