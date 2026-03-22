#!/bin/bash

# Recipio Deployment Script
set -e  # Exit on any error

echo "🚀 Building Recipio for deployment..."

# Create bin directory for build artifacts
mkdir -p deploy/bin

# Build frontend
echo "📦 Building frontend..."
cd src/frontend
npm run build
cd ../..

# Copy frontend build to bin directory
echo "📋 Copying frontend assets..."
cp -r src/frontend/dist deploy/bin/

# Build backend with embedded static files
echo "🔨 Building backend..."
cd src/backend
go build -o ../deploy/bin/recipio-server ./cmd/recipio-server
cd ..

echo "✅ Build complete!"
echo ""
echo "Build artifacts are in deploy/bin/"
echo ""
echo "To run the application:"
echo "  cd deploy/bin"
echo "  ./recipio-server"
echo ""
echo "The server will be available at http://localhost:4002"
echo "Both the API and the SPA will be served from the same server."