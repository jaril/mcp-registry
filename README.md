## MCP Registry

This project is a learning-focused clone of the original Model Context Protocol (MCP) Registry. It aims to replicate the architecture and functionality of the original MCP registry to provide insights into its design, metadata management, configuration handling, and version tracking capabilities.

## Quickstart

### With Docker

```bash
# Build the Docker image
docker build -t registry .

# Run the registry and MongoDB with docker compose
docker compose up
```

### Locally

```bash
# Run MongoDB with Docker on port 27017
docker run -d --name mongodb -p 27017:27017 mongo:latest

# Run the registry
go run main.go
```

## API Endpoints

- [x] GET /v0/health
- [x] GET /v0/servers
- [x] GET /v0/servers/{id}
- [x] GET /v0/ping
- [x] POST /v0/publish

## Configuration

The service can be configured using environment variables:

| Variable                            | Description                     | Default                     |
| ----------------------------------- | ------------------------------- | --------------------------- |
| `MCP_REGISTRY_APP_VERSION`          | Application version             | `dev`                       |
| `MCP_REGISTRY_DATABASE_TYPE`        | Database type                   | `mongodb`                   |
| `MCP_REGISTRY_COLLECTION_NAME`      | MongoDB collection name         | `servers_v2`                |
| `MCP_REGISTRY_DATABASE_NAME`        | MongoDB database name           | `mcp-registry`              |
| `MCP_REGISTRY_DATABASE_URL`         | MongoDB connection string       | `mongodb://localhost:27017` |
| `MCP_REGISTRY_GITHUB_CLIENT_ID`     | GitHub App Client ID            |                             |
| `MCP_REGISTRY_GITHUB_CLIENT_SECRET` | GitHub App Client Secret        |                             |
| `MCP_REGISTRY_LOG_LEVEL`            | Log level                       | `info`                      |
| `MCP_REGISTRY_SEED_FILE_PATH`       | Path to import seed file        | `data/seed.json`            |
| `MCP_REGISTRY_SEED_IMPORT`          | Import `seed.json` on first run | `true`                      |
| `MCP_REGISTRY_SERVER_ADDRESS`       | Listen address for the server   | `:8080`                     |
