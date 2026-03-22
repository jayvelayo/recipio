# Build stage
FROM node:18-alpine AS frontend-build

WORKDIR /app
COPY src/frontend/package*.json ./
RUN npm ci

COPY src/frontend/ ./
RUN npm run build

# Go build stage
FROM golang:1.21-alpine AS backend-build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY src/backend/ ./src/backend/
COPY --from=frontend-build /app/dist ./src/backend/dist/

WORKDIR /app/src/backend
RUN go build -o recipio-server ./cmd/recipio-server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary and static files
COPY --from=backend-build /app/src/backend/recipio-server .
COPY --from=backend-build /app/src/backend/dist ./dist/

EXPOSE 4002

CMD ["./recipio-server"]