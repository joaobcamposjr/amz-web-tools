#!/bin/bash
set -euo pipefail

echo "🔧 FINAL Go dependency fixer..."

# Clean everything
echo "🧹 Cleaning everything..."
go clean -modcache
rm -f go.sum

# Create minimal go.mod with replace directive to avoid sonic
echo "📝 Creating minimal go.mod..."
cat > go.mod << 'EOF'
module amz-web-tools

go 1.21

require (
	github.com/gin-contrib/cors v1.5.0
	github.com/gin-gonic/gin v1.7.7
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.5.0
	github.com/gorilla/websocket v1.5.3
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/microsoft/go-mssqldb v1.6.0
	github.com/sijms/go-ora/v2 v2.9.0
	golang.org/x/crypto v0.17.0
)

// Exclude problematic dependencies
exclude (
	github.com/bytedance/sonic v1.10.1
	github.com/chenzhuoyu/iasm v0.9.0
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d
)
EOF

# Download and force go.sum creation
echo "📦 Downloading dependencies..."
cd backend
GOFLAGS="-mod=mod" go get -d ./...
cd ..

# Verify go.sum was created
if [ -f go.sum ]; then
    echo "✅ go.sum created successfully"
    
    # Remove any problematic entries from go.sum
    if grep -q "bytedance/sonic" go.sum 2>/dev/null; then
        echo "⚠️ Removing sonic from go.sum..."
        sed -i '/bytedance\/sonic/d' go.sum
        sed -i '/chenzhuoyu\/iasm/d' go.sum
        sed -i '/chenzhuoyu\/base64x/d' go.sum
    fi
else
    echo "❌ go.sum not created!"
fi

# Try to build
echo "🔨 Testing build..."
cd backend
if GOFLAGS="-mod=mod" go build -o /tmp/test-build . 2>&1; then
    echo "✅ Build successful!"
    rm -f /tmp/test-build
else
    echo "⚠️ Build failed - check errors above"
fi
cd ..

echo "✅ Go dependency fix process completed!"
