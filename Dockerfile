# syntax = docker.io/docker/dockerfile:1.5

# NOTE: This Dockerfile requires BuildKit. When using `docker build`, set
# DOCKER_BUILDKIT=1.

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.20-alpine3.18 AS build
ENV CGO_ENABLED=0
ARG TARGETPLATFORM
RUN \
  --mount=type=bind,target=/mnt/randomizer \
  --mount=type=cache,id=randomizer.go-pkg,target=/go/pkg \
  --mount=type=cache,id=randomizer.go-build,target=/root/.cache/go-build \
  cd /mnt/randomizer && \
  source ./targetplatform-go-env.sh && \
  go build -v \
    -mod=vendor \
    -ldflags='-s -w' \
    -o /randomizer-server \
    ./cmd/randomizer-server


FROM scratch AS server-binary
COPY --link --from=build /randomizer-server /


FROM docker.io/library/alpine:3.18 AS server-image
COPY --link --from=build /randomizer-server /usr/local/bin/randomizer-server
EXPOSE 7636
ENTRYPOINT ["/usr/local/bin/randomizer-server"]
