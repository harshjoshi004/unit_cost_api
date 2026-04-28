# Build stage
FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o app ./cmd

# Runtime
FROM debian:bookworm-slim

WORKDIR /app
COPY --from=builder /app/app .

EXPOSE 7000

CMD ["./app"]