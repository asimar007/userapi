# ---- Build stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies first.
COPY go.mod ./
COPY go.su[m] ./
RUN go mod download

# Copy source and build a static binary.
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=builder /app/server /app/server

USER app
EXPOSE 8080

ENTRYPOINT ["/app/server"]
