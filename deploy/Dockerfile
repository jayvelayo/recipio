# Build stage
FROM node:18-alpine AS frontend-build

WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Go build stage
FROM golang:1.26-alpine AS backend-build

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
COPY --from=frontend-build /app/dist ./dist/

RUN go build -o recipio-server ./cmd/recipio-server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary and static files
COPY --from=backend-build /app/backend/recipio-server .
COPY --from=backend-build /app/backend/dist ./dist/

EXPOSE 4002

CMD ["./recipio-server"]
