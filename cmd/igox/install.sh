#!/bin/bash
# Read app name from appname.txt file
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd $SCRIPT_DIR

go mod tidy
GOOS=js GOARCH=wasm go build -tags canvas -ldflags -checklinkname=0  -o gdspx.wasm