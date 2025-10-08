#!/bin/bash

echo "🔧 Fixing frontend environment variables..."

cd /d02/projects/amz-web-tools

# Stop frontend
echo "🛑 Stopping frontend..."
pkill -f "node.*standalone.*server.js" 2>/dev/null || true
fuser -k 3000/tcp 2>/dev/null || true
sleep 3

# Set environment variable for build
export NEXT_PUBLIC_API_URL="http://52.206.225.24:8080/api/v1"

# Update all env files
echo "⚙️ Updating environment files..."
echo "NEXT_PUBLIC_API_URL=http://52.206.225.24:8080/api/v1" > .env.local
echo "NEXT_PUBLIC_API_URL=http://52.206.225.24:8080/api/v1" >> .env
sed -i 's/^NEXT_PUBLIC_API_URL=.*/NEXT_PUBLIC_API_URL=http:\/\/52.206.225.24:8080\/api\/v1/' .env

# Clean everything
echo "🧹 Cleaning everything..."
rm -rf .next
rm -rf node_modules/.cache

# Build with explicit env var
echo "🔨 Building with explicit env var..."
NEXT_PUBLIC_API_URL="http://52.206.225.24:8080/api/v1" npm run build

# Copy static files
echo "📦 Copying static files..."
cp -r .next/static .next/standalone/.next/
cp -r public .next/standalone/

# Create a wrapper script that sets the env var
echo "📝 Creating wrapper script..."
cat > .next/standalone/start.sh << 'EOF'
#!/bin/bash
export NEXT_PUBLIC_API_URL="http://52.206.225.24:8080/api/v1"
cd /d02/projects/amz-web-tools/.next/standalone
HOSTNAME=0.0.0.0 PORT=3000 node server.js
EOF

chmod +x .next/standalone/start.sh

# Start frontend with wrapper
echo "🚀 Starting frontend with wrapper..."
cd .next/standalone
nohup ./start.sh > /d02/logs/frontend.log 2>&1 &
echo $! > /d02/pids/frontend.pid

sleep 5

# Check if it started
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend started successfully!"
    echo ""
    echo "🎯 Frontend should now use: http://52.206.225.24:8080/api/v1"
else
    echo "❌ Frontend failed to start. Check logs:"
    tail -10 /d02/logs/frontend.log
fi

