# Unified Ads MCP Server

A Model Context Protocol (MCP) server implementation for managing advertisements across multiple platforms, starting with Facebook Business API.

## Overview

This project provides a unified interface for managing advertisements through MCP, enabling LLMs to perform CRUD operations on ads, campaigns, and ad management tasks. It uses code generation to transform API specifications into MCP tools.

## Features

- **Facebook Business API Integration**: Full support for Facebook Marketing API operations
- **Code Generation**: Automatic generation of MCP tools from API specifications
- **Type Safety**: Strongly typed Go implementations
- **MCP Compliance**: Built using the mcp-go library for standard MCP protocol support

## Architecture

```
unified-ads-mcp/
├── cmd/facebook-mcp/      # Main executable
├── internal/facebook/     # Facebook-specific implementation
│   ├── api_specs/        # Facebook API specifications (JSON)
│   ├── codegen/          # Code generation logic
│   └── generated/        # Generated code (do not edit)
│       ├── types/        # Generated type definitions
│       ├── client/       # Generated client methods
│       ├── tools/        # Generated MCP tools
│       └── enums/        # Generated enum definitions
└── go.mod
```

## Prerequisites

- Go 1.24 or later
- Facebook Access Token with appropriate permissions

## Installation

```bash
go get github.com/yourusername/unified-ads-mcp
```

## Usage

### Running the MCP Server

```bash
export FACEBOOK_ACCESS_TOKEN=your_access_token_here
go run cmd/facebook-mcp/main.go
```

### Using with Claude Desktop

Add the following to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "facebook-ads": {
      "command": "/path/to/unified-ads-mcp",
      "env": {
        "FACEBOOK_ACCESS_TOKEN": "your_access_token_here"
      }
    }
  }
}
```

### Code Generation

To regenerate the code from Facebook API specifications:

```bash
cd internal/facebook/codegen
go run main.go ../api_specs/specs
```

## Available Tools

The server generates over 1400 tools from Facebook's API specifications, including:

- **Ad Management**: Create, read, update, delete ads
- **Campaign Operations**: Manage campaigns and ad sets
- **Insights**: Retrieve performance metrics and analytics
- **Creative Management**: Handle ad creatives and assets
- **Audience Targeting**: Configure targeting parameters

Example tools:
- `facebook_ad_get_insights` - Get ad performance insights
- `facebook_campaign_post_` - Create a new campaign
- `facebook_adaccount_get_ads` - List ads in an account

## Development

### Adding New Platforms

To add support for new advertising platforms:

1. Create a new directory under `internal/` (e.g., `internal/google-ads/`)
2. Add API specifications
3. Implement a code generator following the Facebook example
4. Generate MCP tools and client code

### Testing

```bash
go test ./...
```

## Security

- Never commit access tokens or credentials
- Use environment variables for sensitive configuration
- Follow Facebook's API security best practices

## License

[Your License]

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.
