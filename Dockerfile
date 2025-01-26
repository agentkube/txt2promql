# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /txt2promql ./cmd/api/main.go

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /txt2promql .
COPY configs/config.yaml ./configs/
EXPOSE 8080
CMD ["./txt2promql"]
