# Recipio Deployment Guide

## Local / LAN Deployment

### Single Binary (Recommended)
```bash
make all   # builds frontend + backend into bin/
make run   # starts the server at http://localhost:4002
```

The server serves both the frontend SPA and API at `http://localhost:4002`. No configuration required.

### Development Mode
```bash
make dev
```

---

## Docker Deployment (Home Server with SSL)

Uses the existing Traefik reverse proxy on `traefik_proxy` network for SSL termination.
SSL certs are managed by Traefik via Let's Encrypt + Cloudflare DNS challenge.

### Prerequisites

1. **DNS** — add an A record for `sarap.recipes` pointing to your home IP in Cloudflare.

2. **Port forwarding** — forward port 443 on your router to the server.

3. **Traefik** — ensure `sarap.recipes` is listed in `~/docker/traefik/traefik.toml` under `[[acme.domains]]` and restart Traefik to pick it up.

4. **Google OAuth redirect URI** — add `https://sarap.recipes/auth/google/callback` as an authorized redirect URI in Google Cloud Console.

5. **Environment variables** — copy `.env.example` to `.env` in the repo root and fill in:
   ```
   GOOGLE_CLIENT_ID=
   GOOGLE_CLIENT_SECRET=
   GOOGLE_REDIRECT_URI=https://sarap.recipes/auth/google/callback
   APP_URL=https://sarap.recipes
   ```

### Run

```bash
make deploy        # build and start in background
make deploy-down   # stop
```

Logs: `docker compose -f deploy/docker-compose.yml --env-file .env logs -f`

---

## Cloud Deployment

### Single-server (Railway, Render, Fly.io)
1. Connect your GitHub repository
2. Set build command: `make all`
3. Set start command: `./bin/recipio-server`

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
