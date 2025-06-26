FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY . .


RUN go build -o /build/registry ./main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/registry .
COPY --from=builder /app/data/seed_2025_05_16.json /app/data/seed_2025_05_16.json
EXPOSE 8080

ENTRYPOINT ["./registry"]