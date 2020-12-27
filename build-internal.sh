#!/bin/sh

# This script is intended to run within a BuildKit-powered Dockerfile build, to
# produce container images for multiple architectures.

set -eux

export CGO_ENABLED=0

GOOS="$(echo "$TARGETPLATFORM" | cut -d/ -f1)"
GOARCH="$(echo "$TARGETPLATFORM" | cut -d/ -f2)"
export GOOS GOARCH

if [ "$GOARCH" = "arm" ]; then
  GOARM="$(echo "$TARGETPLATFORM" | cut -d/ -f3 | sed 's/^v//')"
  export GOARM
fi

exec go build -v \
  -mod=vendor \
  -trimpath -ldflags="-s -w" \
  ./cmd/randomizer-server
