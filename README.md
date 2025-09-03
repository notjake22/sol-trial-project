# Solana API

A REST API service for Solana blockchain operations with caching, rate limiting, and authentication.

## Features

- ✅ Solana wallet balance checking
- ✅ Redis caching for performance
- ✅ MongoDB for persistent data
- ✅ API key authentication
- ✅ IP-based rate limiting (10 requests/IP)
- ✅ Concurrent request handling
- ✅ Queue-based processing
- ✅ Docker containerization
- ✅ GitHub Actions CI/CD
- ✅ Health check endpoint

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.23.1+ (for local development)

### Using Docker (Recommended)

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd sol-trial-project
   ```

2. **Start all services:**
   ```bash
   docker-compose up -d
   ```

3. **Check service status:**
   ```bash
   docker-compose ps
   ```

4. **Test the health endpoint:**
   ```bash
   curl http://localhost:8080/health
   ```

5. **Test the API (with authentication):**
   ```bash
   curl -X POST http://localhost:8080/api/get-balance \
     -H "Content-Type: application/json" \
     -H "x-api-key: test-api-key-12345" \
     -d '{
       "wallets": ["11111111111111111111111111111111"]
     }'
   ```

### Local Development

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Set environment variables:**
   ```bash
   export MONGO_URI="mongodb://localhost:27017"
   export MONGO_DB_NAME="Solana"
   export REDIS_URI="redis://localhost:6379"
   export RPC_URI="https://api.mainnet-beta.solana.com"
   export PORT="8080"
   ```

3. **Run the application:**
   ```bash
   go run cmd/main.go
   ```

## API Endpoints

### Health Check
- **GET** `/health` - Service health status (no auth required)

### Solana Operations
- **POST** `/api/get-balance` - Get wallet balance(s)
  - Headers: `x-api-key: <your-api-key>`
  - Body: `{"wallets": ["wallet1", "wallet2", ...]}`

## Deployment

### GitHub Actions Workflow

The project includes automated CI/CD pipeline:

1. **On Push/PR:** Runs tests and linting
2. **On Push to Main:** Builds and pushes Docker image
3. **On Push to Main:** Deploys to Ubuntu server

### Required GitHub Secrets

Set these in your GitHub repository settings:

- `HOST` - Ubuntu server IP address
- `USERNAME` - SSH username for the server
- `SSH_PRIVATE_KEY` - SSH private key for authentication
- `RPC_URI` - Your Solana RPC endpoint

### Manual Server Setup

1. **Install Docker on Ubuntu:**
   ```bash
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   sudo usermod -aG docker $USER
   ```

2. **Install Docker Compose:**
   ```bash
   sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   sudo chmod +x /usr/local/bin/docker-compose
   ```

3. **Pull and run:**
   ```bash
   mkdir ~/solana-api && cd ~/solana-api
   # Copy docker-compose.yml from the repository
   docker-compose up -d
   ```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `MONGO_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `MONGO_DB_NAME` | MongoDB database name | `Solana` |
| `REDIS_URI` | Redis connection string | `redis://localhost:6379` |
| `RPC_URI` | Solana RPC endpoint | Required |

### Rate Limiting

- **IP-based:** 10 requests per IP
- **Authenticated:** All API endpoints require valid API key
- **License tracking:** Usage is tracked per API key

## Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v ./internal/server/rest/handlers/
```

## Architecture

- **Handler Layer:** HTTP request handling (`internal/server/rest/handlers/`)
- **Service Layer:** Business logic (`internal/server/service/`)
- **Repository Layer:** Data access (`internal/server/repo/`)
- **Queue System:** Background processing (`pkg/queue/`)
- **Models:** Data structures (`pkg/models/`)

## License

This project is licensed under the MIT License.