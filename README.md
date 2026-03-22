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

See `deploy/DEPLOYMENT.md` for comprehensive deployment options including Docker, cloud platforms, and more.