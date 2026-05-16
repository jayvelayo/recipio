# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Recipio is a full-stack web app for storing recipes, creating meal plans, and generating grocery lists. Frontend: React/Vite/Tailwind. Backend: Go with native `net/http` (no framework) and SQLite.

## Commands

### Development
```bash
make dev     # Starts backend (port 4003) + frontend (port 4002) together
```

### Backend (Go)
```bash
cd backend
go test ./...                        # All tests
go test ./internal/sqlite_db/...    # DB layer tests only
go test ./cmd/recipio-server/...    # Handler integration tests only
```

### Frontend (React/Vite)
```bash
cd frontend
npm install
npm run build
npm run test
npm run lint
```

### Production build
```bash
make all   # Builds frontend + backend into deploy/bin/
make run   # Starts the server at http://localhost:4002
```

## Environment Configuration

No `.env` required for local development. Copy `.env.example` to `.env` to override defaults (e.g. `PORT`).

## Architecture

### Backend (`backend/`)

All HTTP logic lives in `cmd/recipio-server/`:
- `routes.go` — endpoint registration
- `handlers_recipes.go`, `handlers_mealplans.go`, `handlers_grocery.go`, `handlers_parse.go` — feature handlers
- `middleware.go` — CORS and shared utilities
- `types.go` — request/response structs

Handlers are higher-order functions returning `http.Handler`. Business logic is kept behind the `RecipeDatabase` interface defined in `internal/recipes/recipes_iface.go`, with SQLite implementation in `internal/sqlite_db/sqlite_db.go` and a mock in `internal/recipes/recipes_mock.go` for tests. The DB schema lives as a Go template in `internal/sqlite_db/schema.tmpl`.

### Frontend (`frontend/src/`)

- `App.jsx` — top-level layout (sidebar, header, auth gate)
- `routes.jsx` — React Router v7 route definitions
- `apiConfig.js` — deleted; API calls use relative URLs directly
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
