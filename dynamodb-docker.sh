#!/usr/bin/env bash
set -xeuo pipefail

DOCKER="${DOCKER:-docker}"
VOLUME_NAME="${VOLUME_NAME:-dynamodb-randomizer}"
CONTAINER_NAME="${CONTAINER_NAME:-dynamodb-randomizer}"
HOST_PORT="${HOST_PORT:-8000}"

if ! "$DOCKER" volume inspect "$VOLUME_NAME" >/dev/null 2>&1; then
  "$DOCKER" volume create "$VOLUME_NAME"
fi

"$DOCKER" run \
  --rm -d --name "$CONTAINER_NAME" \
  -p "$HOST_PORT":8000 \
  -v "$VOLUME_NAME":/var/lib/randomizer \
  amazon/dynamodb-local -jar DynamoDBLocal.jar -dbPath /var/lib/randomizer
