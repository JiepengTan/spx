#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source $SCRIPT_DIR/common/setup_env.sh

setup_global_variables

# Get mode parameter (default: worker)
MODE="${1:-worker}"

# Validate mode
if [[ "$MODE" != "worker" && "$MODE" != "minigame" && "$MODE" != "miniprogram" ]]; then
    echo "Error: Invalid mode '$MODE'. Supported modes: worker, minigame, miniprogram"
    echo "Usage: $0 [mode]"
    exit 1
fi

# Download URLs based on godot release pattern
VERSION=$(cat $SCRIPT_DIR/version)
URL_PREFIX="https://github.com/jiepengtan/godot/releases/download/spx${VERSION}/"

# Template package names from godot CI based on mode
case "$MODE" in
    worker)
        TEMPLATE_NAME="web-worker.zip"
        ;;
    minigame)
        TEMPLATE_NAME="web-minigame.zip"
        ;;
    miniprogram)
        TEMPLATE_NAME="web-miniprogram.zip"
        ;;
esac

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

echo "===> Setting up web $MODE templates..."
echo "Mode: $MODE"
echo "Version: $VERSION"
echo "URL Prefix: $URL_PREFIX"
echo "Template Name: $TEMPLATE_NAME"
echo "Template Directory: $TEMPLATE_DIR"

# Create destination and template directories
mkdir -p "$DST_DIR"
mkdir -p "$TEMPLATE_DIR"

# Download template if not exists
TEMPLATE_FILE="$DST_DIR/gdspx${VERSION}_web${MODE}.zip"
if [ -f "$TEMPLATE_FILE" ]; then
    echo "Web $MODE template already exists, skipping download"
else
    echo "Downloading web $MODE template..."
    echo "URL: ${URL_PREFIX}${TEMPLATE_NAME}"
    if curl -L -o "$TEMPLATE_FILE" "${URL_PREFIX}${TEMPLATE_NAME}"; then
        echo "Download successful: $TEMPLATE_FILE"
    else
        echo "Error: Failed to download web $MODE template"
        echo "Make sure the godot release contains: $TEMPLATE_NAME"
        exit 1
    fi
fi

# Setup template directory structure with mode-specific templates
echo "===> Setting up template directory structure..." "$TEMPLATE_FILE" "$TEMPLATE_DIR"

cp -f "$TEMPLATE_FILE" "$TEMPLATE_DIR/web_dlink_nothreads_debug.zip"
cp -f "$TEMPLATE_FILE" "$TEMPLATE_DIR/web_dlink_nothreads_release.zip"
cp -f "$TEMPLATE_FILE" "$TEMPLATE_DIR/web_nothreads_debug.zip"
cp -f "$TEMPLATE_FILE" "$TEMPLATE_DIR/web_nothreads_release.zip"
cp -f "$TEMPLATE_FILE" "$TEMPLATE_DIR/web_dlink_debug.zip"
cp -f "$TEMPLATE_FILE" "$TEMPLATE_DIR/web_dlink_release.zip"
cp -f "$TEMPLATE_FILE" "$TEMPLATE_DIR/web_debug.zip"
cp -f "$TEMPLATE_FILE" "$TEMPLATE_DIR/web_release.zip"

echo "===> Web $MODE setup complete"
echo "  - Template downloaded: $TEMPLATE_FILE"
echo "  - Templates installed to: $TEMPLATE_DIR"
echo "  - Mode: $MODE"