# syntax=docker/dockerfile:1
FROM golang:1.23.12-bookworm AS build
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
COPY . .
RUN --mount=type=cache,id=yafa-go-modules,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,id=yafa-go-build-1.23-${TARGETARCH},target=/root/.cache/go-build,sharing=locked \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/yafa-api ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/yafa-context-import ./cmd/context-import

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl openssh-client && rm -rf /var/lib/apt/lists/* && useradd --uid 10001 --create-home yafa
COPY --from=build /out/yafa-api /usr/local/bin/yafa-api
COPY --from=build /out/yafa-context-import /usr/local/bin/yafa-context-import
USER 10001:10001
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 CMD curl --fail --silent http://127.0.0.1:8080/health/live || exit 1
ENTRYPOINT ["yafa-api"]
