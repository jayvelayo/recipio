# Recipio

Personal project to learn different technology stack. Using a real-world problem, "How do I collect all recipes from different social medias into one."

Also see: [Relevant XKCD](https://xkcd.com/927/)

### Development Setup

1. **Configure environment:**
   ```bash
   # Edit .env file with your settings
   ALLOWED_ORIGINS=http://192.168.1.170:4002
   VITE_API_BASE=http://192.168.1.170:4002
   ```

2. **Setup symlinks:**
   ```bash
   cd frontend && ln -sf ../.env .env
   ```

3. **Start development servers:**
   ```bash
   # Terminal 1: Backend
   cd backend && go run cmd/recipio-server/main.go

   # Terminal 2: Frontend
   cd frontend && npm run dev
   ```

### Testing

```bash
# Backend — all tests
cd backend && go test ./...

# Backend — by layer
go test ./internal/sqlite_db/...   # DB layer
go test ./cmd/recipio-server/...   # Handler integration tests

# Frontend
cd frontend && npm test
```

### Database Snapshots

```bash
# Save a timestamped copy of the database to snapshots/
# Linux: ~/.cache/recipio/recipes.db
# macOS: ~/Library/Caches/recipio/recipes.db
./tools/snapshot-db.sh
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

**Setup:** Create a symlink for the frontend to access the `.env` file:
```bash
cd frontend && ln -sf ../.env .env
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