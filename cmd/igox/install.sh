#!/bin/bash
# Read app name from appname.txt file
go mod tidy
GOOS=js GOARCH=wasm go build -tags canvas -ldflags -checklinkname=0  -o igdspx.wasm