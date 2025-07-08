<p align="center">
  <a href="https://github.com/promobase/unified-ads-mcp">
     <picture>
      <source srcset="https://raw.githubusercontent.com/promobase/identity/refs/heads/main/assets/main_logo_dark.svg" media="(prefers-color-scheme: dark)">
      <source srcset="https://raw.githubusercontent.com/promobase/identity/refs/heads/main/assets/main_logo.svg" media="(prefers-color-scheme: light)">
      <img src="packages/web/src/assets/logo-ornate-light.svg" alt="unified-ads-mcp logo">
    </picture>
  </a>
</p>

<h1 align="center">Unified Ads MCP</h1>

<p align="center">
  <strong>AI-powered advertising management across platforms</strong>
</p>

<p align="center">
  MCP (Model Context Protocol) server that enables Claude and other AI assistants to manage advertising campaigns programmatically across Facebook, Google Ads, and TikTok.
</p>

<p align="center">
  <a href="https://github.com/promobase/unified-ads-mcp/releases/latest">
    <img alt="GitHub Release" src="https://img.shields.io/github/v/release/promobase/unified-ads-mcp?style=flat-square&color=blue" />
  </a>
  <a href="https://github.com/promobase/unified-ads-mcp/actions/workflows/publish.yml">
    <img alt="Build Status" src="https://img.shields.io/github/actions/workflow/status/promobase/unified-ads-mcp/publish.yml?style=flat-square&branch=main" />
  </a>
  <a href="https://github.com/promobase/unified-ads-mcp/blob/main/LICENSE">
    <img alt="License" src="https://img.shields.io/github/license/promobase/unified-ads-mcp?style=flat-square" />
  </a>
  <a href="https://github.com/promobase/unified-ads-mcp/commits/main">
    <img alt="Commits" src="https://img.shields.io/github/commit-activity/m/promobase/unified-ads-mcp?style=flat-square" />
  </a>
</p>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-usage">Usage</a> •
  <a href="#-development">Development</a> •
  <a href="#-roadmap">Roadmap</a>
</p>

---

## Features

<table>
<tr>
<td width="33%" valign="top">

### Facebook Ads
- **162+ Tools** for complete ad management
- Campaign creation & optimization
- Audience targeting & insights
- Budget management
- Performance analytics

</td>
<td width="33%" valign="top">

### Coming Soon
- **Google Ads** integration
- **TikTok Business API** support
- Unified dashboard
- Cross-platform campaigns
- Automated optimization

</td>
<td width="33%" valign="top">

### AI-Powered
- Natural language commands
- Smart recommendations
- Automated reporting
- Performance predictions
- Budget optimization

</td>
</tr>
</table>

## Quick Start

### One-Line Install

```bash
curl -sSL https://raw.githubusercontent.com/promobase/unified-ads-mcp/main/install.sh | bash
```

### Manual Installation

<details>
<summary>Download from releases</summary>

1. Download the latest binary from [releases](https://github.com/promobase/unified-ads-mcp/releases)
2. Extract and make executable:
   ```bash
   unzip unified-ads-mcp-*.zip
   chmod +x unified-ads-mcp
   sudo mv unified-ads-mcp /usr/local/bin/
   ```
</details>

<details>
<summary>Build from source</summary>

```bash
git clone https://github.com/promobase/unified-ads-mcp
cd unified-ads-mcp
make deps
make codegen
make build
```
</details>

## Configuration

### 1. Get Facebook Access Token

1. Visit [Facebook Graph API Explorer](https://developers.facebook.com/tools/explorer/)
2. Select your app or create a new one
3. Add these permissions:
   - `ads_management`
   - `ads_read`
   - `business_management`
   - `read_insights`
4. Generate and copy your access token

### 2. Configure Claude Desktop

The installer automatically configures Claude Desktop. To manually configure:

<details>
<summary>Configuration paths</summary>

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

</details>

Add to your config:

```json
{
  "mcpServers": {
    "unified-ads": {
      "command": "/path/to/unified-ads-mcp",
      "env": {
        "FACEBOOK_ACCESS_TOKEN": "your_token_here"
      }
    }
  }
}
```

### 3. Restart Claude Desktop

## Usage

Once configured, ask Claude to help with your advertising needs:

<table>
<tr>
<td width="50%">

**Account Management**
```
"Show me all my Facebook ad accounts"
"What's the spend limit on account 123?"
"List accounts with active campaigns"
```

</td>
<td width="50%">

**Campaign Operations**
```
"Create a campaign with $50 daily budget"
"Pause all campaigns in account X"
"Duplicate the top performing campaign"
```

</td>
</tr>
<tr>
<td width="50%">

**Performance Analysis**
```
"Show me this week's performance metrics"
"Which ads have the lowest CPA?"
"Compare last month vs this month"
```

</td>
<td width="50%">

**Optimization**
```
"Increase budget for high-performing ads"
"Find underperforming ad sets"
"Suggest audience improvements"
```

</td>
</tr>
</table>

## Development

### Prerequisites

- Go 1.24+
- Make
- Facebook Developer Account

### Project Structure

```
unified-ads-mcp/
├── cmd/server/          # Main server entry point
├── internal/
│   └── facebook/        # Facebook API implementation
│       ├── api_specs/   # API specifications
│       ├── codegen/     # Code generation tools
│       └── generated/   # Generated API tools
├── scripts/             # Build and release scripts
└── Makefile            # Build automation
```

### Building

```bash
# Install dependencies
make deps

# Generate API tools from specs
make codegen

# Build binary with version info
make build

# Run HTTP server for testing
make run
```



## Available Tools

<details>
<summary><strong>162 Facebook Business API Tools</strong></summary>

### Core Objects & Operations

| Object | Tools | Key Operations |
|--------|-------|----------------|
| **AdAccount** | 111 | Campaigns, budgets, insights, audiences |
| **Campaign** | 13 | Create, update, manage campaign settings |
| **AdSet** | 19 | Targeting, scheduling, optimization |
| **Ad** | 13 | Creative management, status control |
| **AdCreative** | 6 | Asset management, creative optimization |

### Example Tool Usage

```javascript
// List campaigns
AdAccount_GET_campaigns

// Create campaign
Campaign_POST_

// Get insights
AdSet_GET_insights

// Update budgets
AdAccount_POST_
```

</details>

## Roadmap

<table>
<tr>
<td width="33%">

### Completed
- [x] Facebook Business API
- [x] MCP protocol implementation
- [x] 162+ advertising tools
- [x] Auto-configuration
- [x] Cross-platform binaries

</td>
<td width="33%">

### In Progress
- [ ] Google Ads integration
- [ ] TikTok Business API
- [ ] Rate limiting
- [ ] Batch operations
- [ ] Enhanced error handling

</td>
<td width="33%">

### Planned
- [ ] Unified analytics dashboard
- [ ] Cross-platform campaigns
- [ ] AI-powered optimization
- [ ] Webhook support
- [ ] Multi-account management

</td>
</tr>
</table>

## Contributing

We welcome contributions! Please check our [Contributing Guidelines](CONTRIBUTING.md) before submitting PRs.

<details>
<summary>Development setup</summary>

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a pull request

</details>

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built on [MCP (Model Context Protocol)](https://github.com/anthropics/mcp)
- Powered by [mcp-go](https://github.com/mark3labs/mcp-go)
- Facebook Business SDK specifications

---

<p align="center">
  <strong>Built for the AI-powered advertising future</strong>
</p>

<p align="center">
  <a href="https://github.com/promobase/unified-ads-mcp/issues">Report Bug</a> •
  <a href="https://github.com/promobase/unified-ads-mcp/issues">Request Feature</a> •
  <a href="https://github.com/promobase/unified-ads-mcp/discussions">Discussions</a>
</p>