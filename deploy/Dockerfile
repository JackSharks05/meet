# meet — Go API server (forked from Timeful, AGPL-3.0; see /NOTICE).
# Multi-stage build → small static image. Build context is the repo root.

# --- build stage ---
FROM golang:1.23-alpine AS build
WORKDIR /src

# Download deps first (cached layer)
COPY server/go.mod server/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Build (cache the module + build caches across rebuilds)
COPY server/ ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/meet-server .

# --- run stage ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/meet-server /app/meet-server
# The server writes a logs.log in its workdir; make it writable by the app user.
RUN chown -R app:app /app
USER app
EXPOSE 3002
ENTRYPOINT ["/app/meet-server"]
