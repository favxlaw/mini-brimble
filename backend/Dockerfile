FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server .

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    git \
    curl \
    docker.io \
    && rm -rf /var/lib/apt/lists/*

RUN curl -sSL https://railpack.com/install.sh | sh

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]