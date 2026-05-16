# Recipio

Personal project to learn different technology stack. Using a real-world problem, "How do I collect all recipes from different social medias into one."

Also see: [Relevant XKCD](https://xkcd.com/927/)

### Architecture

Recipio uses a single-server architecture where the Go binary acts as both the web server and API server.

```
Browser
   │
   ▼
Go Server (:4002)
   ├── /* → serves React SPA (static files from dist/)
   └── /recipes, /meal-plans, ... → API handlers
```

The frontend uses relative URLs for all API calls, so everything resolves to the same origin regardless of host or port. No CORS configuration needed in this setup.

In development, Vite runs on port 4002 and proxies API requests to Go on port 4003, mirroring the production topology from the browser's perspective.

**Future:** API servers will be separated into their own cluster behind a load balancer. The web server and API server will have distinct origins at that point, requiring CORS configuration. Only if I get the chance to do this :-)

### Development Setup

```bash
make dev
```

### Testing

```bash
# Backend — all tests
cd backend && go test ./...

# Backend — by layer
go test ./internal/sqlite_db/...   # DB layer
go test ./cmd/recipio-server/...   # Handler integration tests

# Frontend — all tests
cd frontend && npm test

# Frontend — specific file
cd frontend && npm test -- tests/Recipes.test.jsx
```

Frontend tests use **Vitest** + **@testing-library/react** and cover:
- `RecipeList` — loading/empty states, search filtering, delete confirmation
- `ViewRecipe` — rendering ingredients/instructions/tags, edit mode toggle, save/cancel
- `AddRecipeForm` — manual entry, AI parse preview, error states, form submission

API calls are mocked via `vi.mock` so tests run without a backend.

### Database Snapshots

```bash
# Save a timestamped copy of the database to snapshots/
# Linux: ~/.cache/recipio/recipes.db
# macOS: ~/Library/Caches/recipio/recipes.db
./tools/snapshot-db.sh
```

### Production Deployment
```bash
make all   # builds frontend + backend into bin/
make run   # starts the server at http://localhost:4002
```

See `deploy/DEPLOYMENT.md` for deployment options including Docker and cloud platforms.
