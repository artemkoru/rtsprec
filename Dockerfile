# Backend image: builds ./cmd/rtsprec and runs it as a non-root user.
# Configuration is read only from environment variables (see docs/deployment.md).

FROM golang:1.26.0-bookworm AS build
WORKDIR /src

# go-sqlite3 needs cgo, so the build stage keeps gcc (bookworm includes it)
COPY go.mod go.sum ./
RUN go mod download

COPY . .
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
