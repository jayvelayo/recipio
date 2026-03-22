#!/bin/bash

REPO_ROOT=$(git rev-parse --show-toplevel)
BIN_PATH="bin/"

# Recipio Deployment Script
set -e  # Exit on any error

echo "🚀 Building Recipio for deployment..."

# Create bin directory for build artifacts
mkdir -p ${BIN_PATH}

# Build frontend
echo "📦 Building frontend..."
npm run --prefix src/frontend build

# Copy frontend build to bin directory
echo "📋 Copying frontend assets..."
cp -r src/frontend/dist ${BIN_PATH}/

# Build backend with embedded static files
echo "🔨 Building backend..."
go build -C src/backend/cmd/recipio-server/ -o ${REPO_ROOT}/${BIN_PATH}/recipio-server

echo "✅ Build complete!"
echo ""
echo "Build artifacts are in ${BIN_PATH}/"
echo ""
echo "To run the application:"
echo "  cd ${BIN_PATH}"
echo "  ./recipio-server"
echo ""
echo "The server will be available at http://localhost:4002"
echo "Both the API and the SPA will be served from the same server."