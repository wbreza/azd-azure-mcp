using System.CommandLine;
using Microsoft.Extensions.DependencyInjection;

class Program
{
    public static async Task<int> Main(string[] args)
    {
        var services = new ServiceCollection();

        // Register commands
        services.AddTransient<ServerCommand>();

        var serviceProvider = services.BuildServiceProvider();
        var rootCommand = new RootCommand("Azure MCP CLI Tool");

        // Add commands from DI
        rootCommand.AddCommand(serviceProvider.GetRequiredService<ServerCommand>().Build());

        // ✅ Parse & execute
        return await rootCommand.InvokeAsync(args);
    }
}
