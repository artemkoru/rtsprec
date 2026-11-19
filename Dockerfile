# Backend image: builds the web UI and ./cmd/rtsprec, runs the binary as a non-root user.
# Configuration is read only from environment variables (see docs/deployment.md).

# Web UI: production build of web/ (embedded into the binary by web/embed.go)
FROM node:20-bookworm-slim AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26.0-bookworm AS build
WORKDIR /src

# go-sqlite3 needs cgo, so the build stage keeps gcc (bookworm includes it)
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Replace the committed placeholder with the built UI before compiling the embed
COPY --from=web /web/dist ./web/dist
RUN CGO_ENABLED=1 go build -o /out/rtsprec ./cmd/rtsprec

FROM debian:bookworm-slim AS runtime
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --uid 10001 --create-home rtsprec

WORKDIR /home/rtsprec
COPY --from=build /out/rtsprec /usr/local/bin/rtsprec

USER rtsprec
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/rtsprec"]
