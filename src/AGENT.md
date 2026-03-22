# Description

This repo contains code for 'Recipio' a webapp designed to store/view recipes, create a meal plan, and generate grocery lists based on the meal plan. 

# Project Structure

The repo is divided into two stacks, the frontend and backend.

## Frontend 

This is the web UI written in React / Javascript using Vite. To start the web server, you can run

```
cd src/frontend
npm run dev
```

## Backend

The backend is written in Golang with a sqlite as its database. The goal is to be as modular as possible to easily scale the system. 

To build and run the backend, run:
```
cd src/backend
go run cmd/recipio-server/recipio_server.go
```

# Rules

1. Add as many unit tests, and integration tests as much as possible
2. Try to reduce code duplication as much as possible
3. Write for maintainability over cleverness