# Recipio Deployment Guide

Your React SPA is already properly configured and ready for deployment! Here are several deployment options:

## 🚀 Quick Local Deployment

### Option 1: Single Binary (Recommended)
```bash
# Run the deployment script
./deploy/deploy.sh

# Or manually:
cd src/frontend && npm run build
cp -r dist ../deploy/bin/
cd ../backend && go build -o ../deploy/bin/recipio-server ./cmd/recipio-server
cd ../deploy/bin && ./recipio-server
```

The server will serve both the SPA and API at `http://localhost:4002`

### Option 2: Development Mode
```bash
# Terminal 1: Start backend
cd src/backend && go run ./cmd/recipio-server

# Terminal 2: Start frontend dev server
cd src/frontend && npm run dev
```

## 🐳 Docker Deployment

### Build and Run Container
```bash
# Build the image (run from project root)
docker build -f deploy/Dockerfile -t recipio .

# Run the container
docker run -p 4002:4002 recipio
```

### Using Docker Compose (with persistent data)
```bash
# Create data directory
mkdir data

# Run with docker-compose
docker-compose -f deploy/docker-compose.yml up -d
```

## ☁️ Cloud Deployment Options

### Vercel (Frontend Only)
```bash
# Install Vercel CLI
npm i -g vercel

# Deploy frontend
cd src/frontend
vercel --prod
```

### Railway, Render, or Fly.io (Full Stack)
1. Connect your GitHub repository
2. Set build command: `./deploy/deploy.sh`
3. Set start command: `./deploy/bin/recipio-server`
4. The platform will automatically detect Go and build your app

### Heroku
```yaml
# Create heroku.yml in root
build:
  docker:
    web: deploy/Dockerfile
run:
  web: ./deploy/bin/recipio-server
```

### Netlify (Frontend) + Railway (Backend)
- Deploy frontend to Netlify
- Deploy backend to Railway
- Update API_BASE in frontend to point to Railway URL

## 🔧 Production Optimizations

### Environment Variables
Create a `.env` file for configuration:
```
API_BASE=https://your-api-domain.com
NODE_ENV=production
```

### CDN for Static Assets
Consider using a CDN like Cloudflare for faster global delivery.

### SSL/TLS
Use services like Let's Encrypt or Cloudflare for HTTPS.

## 📊 Monitoring & Analytics

Consider adding:
- Error tracking (Sentry)
- Analytics (Google Analytics, Plausible)
- Performance monitoring (Lighthouse, Web Vitals)

## 🎯 Deployment Checklist

- [ ] Test build locally: `npm run build`
- [ ] Test API endpoints
- [ ] Verify CORS settings
- [ ] Check responsive design
- [ ] Test in different browsers
- [ ] Set up domain and SSL
- [ ] Configure monitoring
- [ ] Set up backups for database

Your SPA is production-ready with code splitting, minification, and optimized assets!