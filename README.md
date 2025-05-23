# Azure MCP via `azd` extensions

- `mcp.azure` - Root extension that host the Azure MCP server with dynamic discovery
- `mcp.azd` - Extension MCP server for working with `azd` CLI
- `mcp.resource` - Extension MCP server for working with generic Azure resources
- `mcp.keyvault` - Extension MCP server for working with Azure KeyVault resources
- `mcp.storage` - Extension MCP server for working with Azure Storage Accounts

## Highlights

- Single server registration in VS Code
- Single tool "azure" in the server
- Root tool discovers capabilities via `azd` extensions.
- Extensions are dynamically installed and invoked by the root server
- Each `azd` extension such as "storage", "keyvault" etc hosts its own MCP server
- Root server acts as an MCP server and client and dynamically dispatches calls to the child servers
- Azure teams can build, ship and iterate without any releases or changes required to the root extension
- Extensions can be built in any language
