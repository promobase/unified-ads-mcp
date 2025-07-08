# Unified Ads MCP Server

MCP (Model Context Protocol) server for Facebook Business API, with plans to expand to Google Ads and TikTok Business API.

## Prerequisites

1. Go 1.21 or higher
2. Facebook Access Token (get from [Facebook Graph API Explorer](https://developers.facebook.com/tools/explorer/))

## Setup

1. Install dependencies:
```bash
make deps
```

2. Generate the API tools:
```bash
make codegen
```

3. Set your Facebook access token:
```bash
export FACEBOOK_ACCESS_TOKEN="your_access_token_here"
```

## Running the Server

### Stdio Mode (default)
```bash
make build
./unified-ads-mcp
```

### HTTP Mode
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

## Future Plans

- Add Google Ads API support
- Add TikTok Business API support
- Implement authentication management
- Add rate limiting and retry logic
- Support for batch operations