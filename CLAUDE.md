# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Recipio is a full-stack web app for storing recipes, creating meal plans, and generating grocery lists. Frontend: React/Vite/Tailwind. Backend: Go with native `net/http` (no framework) and SQLite.

## Commands

### Backend (Go)
```bash
cd src/backend
go run cmd/recipio-server/main.go   # Dev server (port 4002)
go test ./...                        # All tests
go test ./internal/sqlite_db/...    # DB layer tests only
go test ./cmd/recipio-server/...    # Handler integration tests only
```

### Frontend (React/Vite)
```bash
cd src/frontend
npm install
npm run dev      # Dev server (http://localhost:5173)
npm run build
npm run test
npm run lint
```

### Both at once
```bash
cd src && ./run_server_client.sh
```

### Production build
```bash
./deploy/deploy.sh           # Builds frontend + backend into deploy/bin/
cd deploy/bin && ./recipio-server
```

## Environment Configuration

Frontend and backend share a single `.env` file at the repo root. Frontend accesses it via symlink:
```bash
./setup.sh   # or: cd src/frontend && ln -sf ../.env .env
```

Key variables:
- `ALLOWED_ORIGINS` — comma-separated CORS origins for backend (localhost always allowed)
- `VITE_API_BASE` — frontend API URL (defaults to `http://localhost:4002`)

## Architecture

### Backend (`src/backend/`)

All HTTP logic lives in `cmd/recipio-server/`:
- `routes.go` — endpoint registration
- `handlers_recipes.go`, `handlers_mealplans.go`, `handlers_grocery.go`, `handlers_parse.go` — feature handlers
- `middleware.go` — CORS and shared utilities
- `types.go` — request/response structs

Handlers are higher-order functions returning `http.Handler`. Business logic is kept behind the `RecipeDatabase` interface defined in `internal/recipes/recipes_iface.go`, with SQLite implementation in `internal/sqlite_db/sqlite_db.go` and a mock in `internal/recipes/recipes_mock.go` for tests. The DB schema lives as a Go template in `internal/sqlite_db/schema.tmpl`.

### Frontend (`src/frontend/src/`)

- `App.jsx` — top-level layout (sidebar, header, auth gate)
- `routes.jsx` — React Router v7 route definitions
- `apiConfig.js` — API base URL (reads `VITE_API_BASE`)
- `pages/` — feature pages: `recipes/`, `mealplan/`, `grocery/`, `common/`
- `pages/common/AuthContext.jsx` — auth state via React Context + localStorage
- API calls are co-located in `*_apis.jsx` files next to their pages and use TanStack React Query for caching/state

### UI Design

Tailwind CSS with indigo as the primary color and pink as the accent. `tailwind.config.js` defines the custom color tokens. Use Feather icons via `react-icons`.

## Key Conventions

- Keep backend interface-driven: new data operations go through `RecipeDatabase`; add to the interface, implement in SQLite, and update the mock.
- Write integration tests for new handlers using `httptest` and the mock DB.
- Do not add new files outside the established directory structure.
- Write for maintainability over cleverness; reduce code duplication.
