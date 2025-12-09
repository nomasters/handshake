# Build stage
FROM cgr.dev/chainguard/go:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY cmd/ cmd/

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /handshake ./cmd/handshake

# Runtime stage
FROM cgr.dev/chainguard/static:latest

COPY --from=builder /handshake /handshake

ENTRYPOINT ["/handshake"]
