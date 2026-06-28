# ============================================================================
# BUILDER STAGE
# ============================================================================
FROM golang:1.24-alpine AS builder

ARG SERVICE=gateway

WORKDIR /build

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    if [ "$SERVICE" = "gateway" ]; then \
        CGO_ENABLED=0 GOOS=linux go build \
            -ldflags="-s -w" \
            -o /dist/app ./cmd/gateway; \
    elif [ "$SERVICE" = "user" ]; then \
        CGO_ENABLED=0 GOOS=linux go build \
            -ldflags="-s -w" \
            -o /dist/app ./cmd/user; \
    fi

# ============================================================================
# GATEWAY SERVICE
# ============================================================================
FROM alpine:3.19 AS gateway

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /dist/app /app/server

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/server"]

# ============================================================================
# USER SERVICE
# ============================================================================
FROM alpine:3.19 AS user

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /dist/app /app/server

EXPOSE 8081

ENTRYPOINT ["/app/server"]
