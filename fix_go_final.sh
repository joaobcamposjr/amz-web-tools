#!/bin/bash
set -euo pipefail

echo "🔧 FINAL Go dependency fixer..."

# Clean everything
echo "🧹 Cleaning everything..."
go clean -modcache
rm -f go.sum

# Create minimal go.mod
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
EOF

# Download and force go.sum creation
echo "📦 Downloading dependencies..."
cd backend
go get ./...
cd ..

# Verify go.sum was created
if [ -f go.sum ]; then
    echo "✅ go.sum created successfully"
else
    echo "❌ go.sum not created, trying alternative method..."
    cd backend
    go list -m all
    cd ..
fi

# Try to build
echo "🔨 Testing build..."
cd backend
if go build -o /tmp/test-build .; then
    echo "✅ Build successful!"
    rm /tmp/test-build
else
    echo "⚠️ Build failed - check errors above"
fi
cd ..

echo "✅ Go dependency fix process completed!"
