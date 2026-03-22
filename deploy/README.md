# Recipio Deployment

This folder contains all deployment-related files and build artifacts for the Recipio application.

## Structure

- `deploy.sh` - Deployment script that builds both frontend and backend
- `bin/` - Build artifacts (created by deploy.sh)
  - `dist/` - Frontend build output
  - `recipio-server` - Backend Go binary
- `Dockerfile` - Multi-stage Docker build for containerized deployment
- `docker-compose.yml` - Docker Compose configuration with persistent storage
- `nginx.conf` - Nginx configuration for static file serving
- `DEPLOYMENT.md` - Comprehensive deployment guide
- `.dockerignore` - Docker ignore file

## Quick Start

```bash
# Build and run locally
./deploy.sh
cd bin && ./recipio-server

# Or use Docker
docker build -f deploy/Dockerfile -t recipio .
docker run -p 4002:4002 recipio
```

See `DEPLOYMENT.md` for detailed deployment options.