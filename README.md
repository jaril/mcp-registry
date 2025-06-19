# MCP Registry

This project is a learning-focused clone of the original [Model Context Protocol (MCP) Registry](https://github.com/modelcontextprotocol/registry). It aims to replicate the architecture and functionality of the original MCP registry to provide insights into its design, metadata management, configuration handling, and version tracking capabilities.

## 🚀 Features

### Core Functionality

- **Server Registry**: Store and retrieve MCP server metadata
- **Search & Discovery**: Find servers by name with flexible search
- **RESTful API**: Clean HTTP API with JSON responses
- **Health Monitoring**: Built-in health checks and status endpoints

### Architecture & Configuration

- **Modular Design**: Clean separation between models, storage, and HTTP layers
- **Configurable Storage**: Support for in-memory and persistent storage backends
- **Environment-Based Config**: Flexible configuration via environment variables and CLI flags
- **Graceful Shutdown**: Proper signal handling for production deployments

### Production Features

- **Structured Logging**: Request logging with timing and error tracking
- **CORS Support**: Cross-origin requests for web browser integration
- **Middleware System**: Extensible HTTP middleware pipeline
- **Development Tools**: Debug endpoints and development-specific features

## 📖 API Documentation

### Endpoints

| Method | Endpoint                       | Description                   |
| ------ | ------------------------------ | ----------------------------- |
| `GET`  | `/health`                      | Service health check          |
| `GET`  | `/servers`                     | List all registered servers   |
| `POST` | `/servers`                     | Register a new server         |
| `GET`  | `/servers/{id}`                | Get specific server details   |
| `GET`  | `/servers/count`               | Get server statistics         |
| `GET`  | `/servers/search?name={query}` | Search servers by name        |
| `GET`  | `/debug/config`                | View configuration (dev only) |

### Example Server Object

```json
{
  "id": "1",
  "name": "filesystem-server",
  "description": "A server for accessing local filesystem",
  "version": "1.0.0",
  "repository": "https://github.com/example/filesystem-server",
  "author": "Jane Doe",
  "tags": ["filesystem", "local", "files"],
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z"
}
```

## 🛠️ Installation & Usage

### Prerequisites

- Go 1.23 or later
- SQLite3 (for database storage)

### Quick Start

```bash
# Clone the repository
git clone <your-repo-url>
cd my-mcp-registry

# Install dependencies
go mod download

# Run with default configuration (in-memory storage)
go run main.go

# Run with persistent storage
go run main.go -storage sqlite

# Run with custom configuration
MCP_PORT=9000 MCP_LOG_LEVEL=debug go run main.go
```

### Configuration Options

Configuration is loaded from environment variables, command-line flags, and defaults (in that order of precedence).

#### Server Configuration

- `MCP_PORT` / `-port`: Server port (default: 8080)
- `MCP_HOST` / `-host`: Server host (default: localhost)
- `MCP_ENVIRONMENT` / `-env`: Environment (dev/staging/production, default: dev)
- `MCP_LOG_LEVEL` / `-log-level`: Log level (debug/info/warn/error, default: info)

#### Storage Configuration

- `MCP_STORAGE_TYPE` / `-storage`: Storage type (memory/sqlite, default: memory)
- `MCP_DATA_PATH`: Data directory path (default: ./data)

#### Feature Flags

- `MCP_ENABLE_CORS`: Enable CORS headers (default: true)
- `MCP_ENABLE_METRICS`: Enable metrics collection (default: false)

### Usage Examples

```bash
# Basic health check
curl http://localhost:8080/health

# List all servers
curl http://localhost:8080/servers

# Search for servers
curl "http://localhost:8080/servers/search?name=filesystem"

# Register a new server
curl -X POST -H "Content-Type: application/json" \
  -d '{
    "id": "my-server",
    "name": "custom-server",
    "description": "My custom MCP server",
    "version": "1.0.0",
    "repository": "https://github.com/me/my-server",
    "author": "Your Name",
    "tags": ["custom", "example"],
    "is_active": true
  }' \
  http://localhost:8080/servers

# Get server statistics
curl http://localhost:8080/servers/count
```

## Testing

For detailed testing instructions, please refer to [TESTING.md](TESTING.md).

## 🏗️ Architecture

### Project Structure

```
my-mcp-registry/
├── main.go                    # Application entry point
├── internal/                  # Private application packages
│   ├── config/               # Configuration management
│   │   └── config.go
│   ├── server/               # HTTP server and middleware
│   │   └── server.go
│   ├── models/               # Data models and business logic
│   │   └── server.go
│   ├── storage/              # Storage implementations
│   │   └── memory.go
│   └── handlers/             # HTTP request handlers
│       └── handlers.go
└── data/                     # Database files (created at runtime)
```

### Design Principles

- **Interface-Driven Development**: Storage layer uses interfaces for flexibility and testability
- **Dependency Injection**: Components receive their dependencies rather than creating them
- **Configuration-Based Behavior**: Runtime behavior controlled through configuration
- **Clean Error Handling**: Explicit error handling with custom error types
- **Thread Safety**: Safe for concurrent access with proper synchronization

## 🔄 Development Status

### ✅ Completed Features

- [x] **HTTP API Foundation**: RESTful endpoints with JSON serialization
- [x] **Clean Architecture**: Modular package structure with clear separation of concerns
- [x] **Configuration System**: Environment variables, CLI flags, and validation
- [x] **Graceful Shutdown**: Signal handling for clean application termination
- [x] **Request Middleware**: Logging, CORS, and request/response handling
- [x] **In-Memory Storage**: Thread-safe storage with mutex synchronization
- [x] **Validation System**: Input validation with detailed error reporting
- [x] **Development Tools**: Debug endpoints and development-specific features
- [x] **Database Integration**: SQLite storage with schema migrations
- [x] **Connection Pooling**: Database connection management and optimization
- [x] **Data Persistence**: Durable storage with transaction support
- [x] **SQL Query Layer**: Prepared statements and query optimization
- [x] **Unit Testing**: Comprehensive test suite for all packages
- [x] **Integration Testing**: End-to-end API testing
- [x] **Test Utilities**: Mock implementations and test helpers
- [x] **Code Coverage**: Coverage reporting and quality metrics
- [x] **Container Support**: Docker images and container orchestration

### 📋 Roadmap

- [ ] **Authentication**: API key-based authentication
- [ ] **Rate Limiting**: Request throttling and abuse prevention
- [ ] **Metrics & Monitoring**: Prometheus metrics and health monitoring
- [ ] **API Documentation**: Auto-generated OpenAPI/Swagger documentation
- [ ] **Multiple Storage Backends**: PostgreSQL, MySQL support
- [ ] **Caching Layer**: Redis-based caching for improved performance
- [ ] **Backup & Recovery**: Database backup and restoration tools

## 📝 License

This project is released under the MIT License. See [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [Model Context Protocol](https://modelcontextprotocol.io) - The official MCP specification
- [MCP Servers](https://github.com/modelcontextprotocol/servers) - Official collection of MCP server implementations

---

**Status**: Active Development | **Go Version**: 1.23+ | **API Version**: v0.1.0
