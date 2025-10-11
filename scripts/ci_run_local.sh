#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd $SCRIPT_DIR
cd ../

# Check if Docker is running
echo "📋 Checking Docker status..."
if ! docker ps &> /dev/null; then
    echo "❌ Docker is not running!"
    echo "Please start Docker Desktop:"
    echo "  open -a Docker"
    echo ""
    echo "After starting, please wait about 30 seconds, then rerun this script."
    exit 1
fi
echo "✅ Docker is running"
echo ""

# Check if act is installed
echo "📋 Checking act installation..."
if ! command -v act &> /dev/null; then
    echo "❌ act is not installed!"
    echo "Please run: brew install act"
    exit 1
fi
echo "✅ act is installed ($(act --version))"
echo ""

target_workflow=.github/workflows/local_build.yml

echo "📋 可用的 jobs:"
act -W $target_workflow -l
echo ""

# Run the action
act push -W $target_workflow -j static-checks