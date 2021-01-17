#!/bin/sh

GOOS="$(echo "$TARGETPLATFORM" | cut -d/ -f1)"
echo "GOOS=$GOOS"
export GOOS

GOARCH="$(echo "$TARGETPLATFORM" | cut -d/ -f2)"
echo "GOARCH=$GOARCH"
export GOARCH

if [ "$GOARCH" = "arm" ]; then
  GOARM="$(echo "$TARGETPLATFORM" | cut -d/ -f3 | sed 's/^v//')"
  echo "GOARM=$GOARM"
  export GOARM
fi
