#!/bin/bash
set -e

# Unified Ads MCP Server Installer
REPO="promobase/unified-ads-mcp"
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="unified-ads-mcp"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="x86_64" ;;
    amd64) ARCH="x86_64" ;;
    aarch64) ARCH="arm64" ;;
    arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin) OS="macOS" ;;
    linux) OS="Linux" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest release
echo "Fetching latest release..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "Error: Could not fetch latest release"
    exit 1
fi

# Download URL
FILENAME="${BINARY_NAME}_${LATEST_RELEASE#v}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$FILENAME"

echo "Downloading $BINARY_NAME $LATEST_RELEASE for $OS $ARCH..."

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download and extract
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if ! curl -sL "$DOWNLOAD_URL" -o "$FILENAME"; then
    echo "Error: Failed to download $DOWNLOAD_URL"
    exit 1
fi

tar -xzf "$FILENAME"
chmod +x "$BINARY_NAME"
mv "$BINARY_NAME" "$INSTALL_DIR/"

cd - > /dev/null
rm -rf "$TMP_DIR"

echo "✓ Installed $BINARY_NAME to $INSTALL_DIR/$BINARY_NAME"

# Add to PATH if needed
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "Add $INSTALL_DIR to your PATH:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi

# Auto-configure Claude Desktop
echo ""
echo "Configuring Claude Desktop..."

# Determine config path
case "$OS" in
    macOS) CONFIG_PATH="$HOME/Library/Application Support/Claude/claude_desktop_config.json" ;;
    Linux) CONFIG_PATH="$HOME/.config/Claude/claude_desktop_config.json" ;;
    *) echo "Please configure Claude Desktop manually"; exit 0 ;;
esac

# Create config directory if it doesn't exist
mkdir -p "$(dirname "$CONFIG_PATH")"

# Check if config exists and has content
if [ -f "$CONFIG_PATH" ] && [ -s "$CONFIG_PATH" ]; then
    # Backup existing config
    cp "$CONFIG_PATH" "$CONFIG_PATH.backup"
    
    # Check if unified-ads is already configured
    if grep -q '"unified-ads"' "$CONFIG_PATH"; then
        echo "✓ unified-ads already configured in Claude Desktop"
    else
        # Add unified-ads to existing config
        # This uses jq if available, otherwise falls back to manual method
        if command -v jq &> /dev/null; then
            jq '.mcpServers["unified-ads"] = {
                "command": "'$INSTALL_DIR/$BINARY_NAME'",
                "env": {
                    "FACEBOOK_ACCESS_TOKEN": "your_token_here"
                }
            }' "$CONFIG_PATH" > "$CONFIG_PATH.tmp" && mv "$CONFIG_PATH.tmp" "$CONFIG_PATH"
            echo "✓ Added unified-ads to Claude Desktop config"
        else
            echo "⚠️  Please manually add to $CONFIG_PATH:"
            echo ""
            cat << EOF
"unified-ads": {
  "command": "$INSTALL_DIR/$BINARY_NAME",
  "env": {
    "FACEBOOK_ACCESS_TOKEN": "your_token_here"
  }
}
EOF
        fi
    fi
else
    # Create new config
    cat > "$CONFIG_PATH" << EOF
{
  "mcpServers": {
    "unified-ads": {
      "command": "$INSTALL_DIR/$BINARY_NAME",
      "env": {
        "FACEBOOK_ACCESS_TOKEN": "your_token_here"
      }
    }
  }
}
EOF
    echo "✓ Created Claude Desktop config at $CONFIG_PATH"
fi

echo ""
echo "⚠️  IMPORTANT: Set your Facebook access token!"
echo ""
echo "1. Get your token at: https://developers.facebook.com/tools/explorer/"
echo "2. Edit $CONFIG_PATH"
echo "3. Replace 'your_token_here' with your actual token"
echo "4. Restart Claude Desktop"