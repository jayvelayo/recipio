BIN_PATH := bin

.PHONY: all backend frontend test run deploy deploy-down clean

all: frontend backend
	@echo ""
	@echo "Build artifacts are in $(BIN_PATH)/"
	@echo "Run with: make run"

dev:
	@trap 'kill 0' EXIT; \
	PORT=4003 go run backend/cmd/recipio-server/main.go & \
	npm run --prefix frontend dev

frontend:
	@echo "Building frontend..."
	npm run --prefix frontend build
	mkdir -p $(BIN_PATH)
	cp -r frontend/dist $(BIN_PATH)/

backend: frontend
	@echo "Building backend..."
	mkdir -p $(BIN_PATH)
	go build -C backend/cmd/recipio-server/ -o $(CURDIR)/$(BIN_PATH)/recipio-server

test:
	@echo "Running backend tests..."
	cd backend && go test ./...
	@echo "Running frontend tests..."
	cd frontend && npm test

run:
	@echo "Starting server at http://localhost:4002"
	./$(BIN_PATH)/recipio-server

deploy:
	docker compose -f deploy/docker-compose.yml --env-file .env up -d --build

deploy-down:
	docker compose -f deploy/docker-compose.yml --env-file .env down

clean:
	rm -rf $(BIN_PATH)
	rm -rf frontend/dist
