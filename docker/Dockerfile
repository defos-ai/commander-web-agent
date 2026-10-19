# syntax=docker/dockerfile:1
FROM golang:1.26.7-bookworm AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOWORK=off
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -trimpath -buildvcs=false -o /out/commander-web-agent ./cmd

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /state \
    && chown 1000:1000 /state
COPY --from=build /out/commander-web-agent /usr/local/bin/commander-web-agent
ENV BUS_URL=file:///state/bus STORAGE_URL=file:///state/storage
USER 1000:1000
WORKDIR /state
ENTRYPOINT ["/usr/local/bin/commander-web-agent"]
