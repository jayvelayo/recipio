# Recipio

Personal project to learn different technology stack. Using a real-world problem, "How do I collect all recipes from different social medias into one."

Also see: [Relevant XKCD](https://xkcd.com/927/)

## 🚀 Quick Start

### Development
```bash
# Install frontend dependencies
cd src/frontend && npm install

# Start backend (Terminal 1)
cd src/backend && go run ./cmd/recipio-server

# Start frontend dev server (Terminal 2)
cd src/frontend && npm run dev
```

### Production Deployment
```bash
# Build and run
./deploy/deploy.sh
cd deploy/bin && ./recipio-server
```

### Configuration

Both frontend and backend use a shared `.env` file for configuration:

```bash
# Backend CORS origins (comma-separated)
ALLOWED_ORIGINS=http://192.168.1.170:4002,https://yourdomain.com

# Frontend API base URL
VITE_API_BASE=http://192.168.1.170:4002
```

**The backend automatically loads the `.env` file** - no need to export variables or rebuild!

### CORS Configuration

The server supports configurable CORS origins via the `ALLOWED_ORIGINS` environment variable:

```bash
# In .env file
ALLOWED_ORIGINS=http://192.168.1.170:4002,https://yourdomain.com
```

**Localhost origins are always allowed** for development convenience.

### Frontend Configuration

The frontend API base URL can be configured via the `VITE_API_BASE` environment variable:

```bash
# In .env file
VITE_API_BASE=http://192.168.1.170:4002
```

If not set, defaults to `http://localhost:4002`.

See `deploy/DEPLOYMENT.md` for comprehensive deployment options including Docker, cloud platforms, and more.