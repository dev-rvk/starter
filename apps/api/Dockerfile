# ── Build stage ────────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 required for distroless (no libc).
# -ldflags="-s -w" strips debug info (~30 % smaller binary).
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" -o bin/api ./cmd/api

# ── Runtime stage ──────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-nonroot:nonroot

WORKDIR /app
COPY --from=builder /app/bin/api ./api

USER nonroot:nonroot

EXPOSE 3002
ENTRYPOINT ["./api"]
