# syntax = docker.io/docker/dockerfile:1.2

# This Dockerfile requires BuildKit. When using `docker build`, set
# DOCKER_BUILDKIT=1.

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.15-alpine3.12 AS golang
FROM golang AS builder

ENV CGO_ENABLED=0
ARG TARGETPLATFORM

RUN \
  --mount=type=bind,target=/mnt/randomizer \
  --mount=type=cache,id=randomizer.go-pkg,target=/go/pkg \
  --mount=type=cache,id=randomizer.go-build,target=/root/.cache/go-build,from=golang,source=/root/.cache/go-build \
  cd /mnt/randomizer && \
  source ./targetplatform-go-env.sh && \
  go build -v \
    -mod=vendor \
    -trimpath -ldflags='-s -w' \
    -o /randomizer-server \
    ./cmd/randomizer-server

# ---

FROM scratch AS server-binary
COPY --from=builder /randomizer-server /

# ---

FROM docker.io/library/alpine:3.12 AS server-image

RUN apk add --no-cache ca-certificates
COPY --from=builder /randomizer-server /usr/local/bin/randomizer-server

EXPOSE 7636
ENTRYPOINT ["/usr/local/bin/randomizer-server"]
