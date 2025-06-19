FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install build dependencies for CGO (needed for SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY . .

# Enable CGO and build
ENV CGO_ENABLED=1
RUN go build -o /build/registry ./main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/registry .
COPY --from=builder /app/internal/storage/migrations.sql ./internal/storage/
EXPOSE 8080

ENTRYPOINT ["./registry"]