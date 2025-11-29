#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source $SCRIPT_DIR/common/setup_env.sh

setup_global_variables

# Download URLs based on godot release pattern
VERSION=$(cat $SCRIPT_DIR/version)
URL_PREFIX="https://github.com/jiepengtan/godot/releases/download/spx${VERSION}/"

# Worker mode package names from godot CI
WORKER_TEMPLATE="web-worker.zip"  # Based on our CI artifact name

# Download paths
DST_DIR="$GOPATH/bin"
TEMPLATE_DIR=""

# Set up platform-specific template directory
if [[ "$(uname)" == "Linux" ]]; then
    TEMPLATE_DIR="$HOME/.local/share/godot/export_templates/$ENGINE_VERSION"
elif [[ "$(uname)" == "Darwin" ]]; then
    TEMPLATE_DIR="$HOME/Library/Application Support/Godot/export_templates/$ENGINE_VERSION"
elif [[ "$(uname -o 2>/dev/null)" == "Msys" ]] || [[ "$(uname -o 2>/dev/null)" == "Cygwin" ]]; then
    TEMPLATE_DIR="$APPDATA/Godot/export_templates/$ENGINE_VERSION"
else
    echo "Unsupported OS for template directory setup"
    exit 1
fi

echo "===> Setting up web worker templates..."
echo "Version: $VERSION"
echo "URL Prefix: $URL_PREFIX"
echo "Template Directory: $TEMPLATE_DIR"

# Create destination and template directories
mkdir -p "$DST_DIR"
mkdir -p "$TEMPLATE_DIR"

# Download worker template if not exists
WORKER_TEMPLATE_FILE="$DST_DIR/gdspx${VERSION}_webworker.zip"
if [ -f "$WORKER_TEMPLATE_FILE" ]; then
    echo "Web worker template already exists, skipping download"
else
    echo "Downloading web worker template..."
    echo "URL: ${URL_PREFIX}${WORKER_TEMPLATE}"
    if curl -L -o "$WORKER_TEMPLATE_FILE" "${URL_PREFIX}${WORKER_TEMPLATE}"; then
        echo "Download successful: $WORKER_TEMPLATE_FILE"
    else
        echo "Error: Failed to download web worker template"
        echo "Make sure the godot release contains: $WORKER_TEMPLATE"
        exit 1
    fi
fi

# Setup template directory structure with worker mode templates
echo "===> Setting up template directory structure..." "$WORKER_TEMPLATE_FILE" "$TEMPLATE_DIR"

cp -f "$WORKER_TEMPLATE_FILE" "$TEMPLATE_DIR/web_dlink_nothreads_debug.zip"
cp -f "$WORKER_TEMPLATE_FILE" "$TEMPLATE_DIR/web_dlink_nothreads_release.zip"
cp -f "$WORKER_TEMPLATE_FILE" "$TEMPLATE_DIR/web_nothreads_debug.zip"
cp -f "$WORKER_TEMPLATE_FILE" "$TEMPLATE_DIR/web_nothreads_release.zip"
cp -f "$WORKER_TEMPLATE_FILE" "$TEMPLATE_DIR/web_dlink_debug.zip"
cp -f "$WORKER_TEMPLATE_FILE" "$TEMPLATE_DIR/web_dlink_release.zip"
cp -f "$WORKER_TEMPLATE_FILE" "$TEMPLATE_DIR/web_debug.zip"
cp -f "$WORKER_TEMPLATE_FILE" "$TEMPLATE_DIR/web_release.zip"

echo "===> Web worker setup complete"
echo "  - Template downloaded: $WORKER_TEMPLATE_FILE"
echo "  - Templates installed to: $TEMPLATE_DIR"
echo "  - Available as: web_worker_debug.zip, web_worker_release.zip"