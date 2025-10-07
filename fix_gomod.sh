#!/bin/bash
set -euo pipefail

echo "🔧 Fixing go.mod dependencies..."

cd /d02/projects/amz-web-tools

# Remove problematic dependency
sed -i '/github.com\/chenzhuoyu\/iasm/d' go.mod

# Clean go cache
go clean -modcache

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

echo "✅ go.mod fixed successfully!"
echo "🚀 Now run: ./init_server_fixed.sh"
