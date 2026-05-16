# Recipio Deployment Guide

## Local / LAN Deployment

### Single Binary (Recommended)
```bash
make all   # builds frontend + backend into deploy/bin/
make run   # starts the server at http://localhost:4002
```

The server serves both the frontend SPA and API at `http://localhost:4002`. No configuration required.

### Development Mode
```bash
make dev
```

---

## Docker Deployment

```bash
# Build
docker build -f deploy/Dockerfile -t recipio .

# Run
docker run -p 4002:4002 recipio

# With persistent data
mkdir -p data
docker-compose -f deploy/docker-compose.yml up -d
```

---

## Cloud Deployment

### Single-server (Railway, Render, Fly.io)
1. Connect your GitHub repository
2. Set build command: `make all`
3. Set start command: `./deploy/bin/recipio-server`

No CORS configuration needed — Go serves both the frontend and API from the same origin.

### Separate frontend + backend clusters
If the frontend (e.g. Vercel, Netlify, CDN) and backend API are on different domains, CORS will need to be configured in `middleware.go` and the frontend API calls will need an explicit base URL instead of relative paths.

For HTTPS deployments, ensure CORS allowed origins use `https://`.

---

## Deployment Checklist

- [ ] Run `make all` and test the binary locally with `make run`
- [ ] Verify API endpoints respond correctly
- [ ] Check CORS if frontend and backend are on separate domains
- [ ] Set up SSL/TLS (Let's Encrypt, Cloudflare)
- [ ] Set up database backups (`~/.cache/recipio/recipes.db`)
