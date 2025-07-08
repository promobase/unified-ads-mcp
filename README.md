# Unified Ads MCP Server

An MCP (Model Context Protocol) server that enables Claude and other AI assistants to manage Facebook Ads programmatically. This server provides 160+ tools for comprehensive ad management, with plans to expand to Google Ads and TikTok Business API.

## What is MCP?

The Model Context Protocol (MCP) allows AI assistants like Claude to interact with external services and tools. This server implements MCP to give Claude the ability to:
- List and manage Facebook ad accounts
- Create and optimize campaigns
- Analyze ad performance and insights
- Manage budgets and targeting
- Handle creative assets

## Installation

### Quick Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/promobase/unified-ads-mcp/main/install.sh | bash
```

This will:
- Download the latest binary for your platform
- Install it to `~/.local/bin/unified-ads-mcp`
- Automatically configure Claude Desktop (creates backup of existing config)
- Prompt you to add your Facebook access token

### Manual Download

Download from the [releases page](https://github.com/promobase/unified-ads-mcp/releases)

### Build from Source

```bash
git clone https://github.com/promobase/unified-ads-mcp
cd unified-ads-mcp
make deps
make codegen
make build
```

## Setup

### 1. Get a Facebook Access Token

1. Go to [Facebook Graph API Explorer](https://developers.facebook.com/tools/explorer/)
2. Select your app or create a new one
3. Add the following permissions:
   - `ads_management`
   - `ads_read`
   - `business_management`
   - `read_insights`
4. Generate and copy your access token

### 2. Configure Claude Desktop

Add the server to your Claude Desktop configuration:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
**Linux**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "unified-ads": {
      "command": "/absolute/path/to/unified-ads-mcp",
      "env": {
        "FACEBOOK_ACCESS_TOKEN": "your_access_token_here"
      }
    }
  }
}
```

### 3. Restart Claude Desktop

After updating the configuration, restart Claude Desktop to load the MCP server.

## Usage in Claude

Once configured, you can ask Claude to:

- **List ad accounts**: "Show me all my Facebook ad accounts"
- **Get campaign insights**: "Get performance data for my active campaigns"
- **Create campaigns**: "Create a new campaign with a $50 daily budget"
- **Manage ads**: "Pause all ads in campaign X"
- **Analyze performance**: "Which ad sets are performing best this week?"

## Running Standalone (for development)

### Stdio Mode (default for MCP)
```bash
export FACEBOOK_ACCESS_TOKEN="your_access_token_here"
./unified-ads-mcp
```

### HTTP Mode (for testing)
```bash
./unified-ads-mcp --transport http
```

The HTTP server will listen on `:8080/mcp`

## Available Tools

The server currently provides 162 tools for Facebook Business API core objects:

- **AdAccount** (111 tools) - Manage ad accounts, campaigns, budgets, insights
- **Campaign** (13 tools) - Create and manage campaigns
- **AdSet** (19 tools) - Configure ad sets, targeting, budgets
- **Ad** (13 tools) - Create and manage individual ads
- **AdCreative** (6 tools) - Manage creative assets

### Example Tool Names
- `AdAccount_GET_campaigns` - List campaigns in an ad account
- `Campaign_POST_` - Create or update a campaign
- `AdSet_GET_insights` - Get performance insights for an ad set
- `Ad_POST_adcreatives` - Attach creatives to an ad

### Built-in Tools
- `health_check` - Check if the server is running
- `check_access_token` - Verify Facebook access token is configured

## Using with Claude

1. Install the MCP CLI tools
2. Configure your Claude desktop app to use this server
3. The tools will be available in your Claude conversation

## Development

### Project Structure
```
unified-ads-mcp/
├── cmd/server/          # Main server entry point
├── internal/facebook/   # Facebook-specific implementation
│   ├── api_specs/      # Facebook API specifications
│   ├── codegen/        # Code generation tools
│   └── generated/      # Generated API tools
└── Makefile            # Build commands
```

### Regenerating Tools
```bash
make codegen
```

This will regenerate all tools from the Facebook API specifications.

## Troubleshooting

1. **No access token error**: Set the `FACEBOOK_ACCESS_TOKEN` environment variable
2. **API errors**: Check that your token has the necessary permissions
3. **Build errors**: Run `make deps` to ensure all dependencies are installed

## Development

### Creating Releases

Releases are created manually using the release script:

```bash
# Patch release (default) - v0.0.1 -> v0.0.2
./scripts/release.sh

# Minor release - v0.1.2 -> v0.2.0
./scripts/release.sh --minor

# Major release - v1.2.3 -> v2.0.0
./scripts/release.sh --major
```

This will:
1. Increment the version
2. Update the VERSION file
3. Create a git tag
4. Push the tag to trigger the release workflow

The GitHub Actions workflow will then automatically build binaries for all platforms and create a release.

## Future Plans

- Add Google Ads API support
- Add TikTok Business API support
- Implement authentication management
- Add rate limiting and retry logic
- Support for batch operations