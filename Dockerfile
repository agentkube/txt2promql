# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /text2promql ./cmd/api/main.go

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /text2promql .
COPY configs/default.yaml ./configs/
EXPOSE 8080
CMD ["./text2promql"]
