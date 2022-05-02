#!/bin/sh

docker="${DOCKER:-docker}"
volume_name="${VOLUME_NAME:-dynamodb-randomizer}"
container_name="${CONTAINER_NAME:-dynamodb-randomizer}"
host_port="${HOST_PORT:-8000}"

if ! "$docker" volume inspect "$volume_name" >/dev/null 2>&1; then
  (set -x; "$docker" volume create "$volume_name") || exit $?
fi

set -x
"$docker" run \
  --rm -d --name "$container_name" \
  -p "$host_port":8000 \
  -v "$volume_name":/var/lib/randomizer \
  amazon/dynamodb-local -jar DynamoDBLocal.jar -dbPath /var/lib/randomizer
