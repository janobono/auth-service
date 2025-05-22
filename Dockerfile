# ----------- Stage 1: Build ----------
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/app ./cmd/auth-service

# ----------- Stage 2: Final image ----------
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/app .

# Default command
ENTRYPOINT ["/app/app"]
