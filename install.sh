#!/bin/bash
set -e

# Unified Ads MCP Server Installer
REPO="promobase/unified-ads-mcp"
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="unified-ads-mcp"

# Check for --snapshot flag
SNAPSHOT=false
if [[ "$*" == *"--snapshot"* ]]; then
    SNAPSHOT=true
fi

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    amd64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Determine release to download
if [ "$SNAPSHOT" = true ]; then
    echo "Installing development snapshot..."
    RELEASE_TAG="snapshot"
    RELEASE_URL="https://api.github.com/repos/$REPO/releases/tags/snapshot"
else
    echo "Fetching latest release..."
    RELEASE_URL="https://api.github.com/repos/$REPO/releases/latest"
fi

# Get release info
RELEASE_INFO=$(curl -s "$RELEASE_URL")
RELEASE_TAG=$(echo "$RELEASE_INFO" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$RELEASE_TAG" ]; then
    echo "Error: Could not fetch release information"
    exit 1
fi

# Download URL
FILENAME="${BINARY_NAME}-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    FILENAME="${FILENAME}.exe"
fi
FILENAME_ZIP="${FILENAME}.zip"

# Get download URL from release assets
DOWNLOAD_URL=$(echo "$RELEASE_INFO" | grep -o "\"browser_download_url\": \"[^\"]*${FILENAME_ZIP}\"" | sed 's/.*: "\(.*\)"/\1/')

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find download URL for $FILENAME_ZIP"
    echo "Available assets:"
    echo "$RELEASE_INFO" | grep -o '"name": "[^"]*"' | sed 's/"name": "//g' | sed 's/"//g'
    exit 1
fi

echo "Downloading $BINARY_NAME ($RELEASE_TAG) for $OS/$ARCH..."

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download and extract
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if ! curl -sL "$DOWNLOAD_URL" -o "$FILENAME_ZIP"; then
    echo "Error: Failed to download $DOWNLOAD_URL"
    exit 1
fi

unzip -q "$FILENAME_ZIP"
chmod +x "$FILENAME"
mv "$FILENAME" "$INSTALL_DIR/$BINARY_NAME"

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