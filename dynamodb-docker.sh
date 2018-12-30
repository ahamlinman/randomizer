#!/usr/bin/env bash

set -xeuo pipefail

volume_name="dynamodb-randomizer"

if ! docker volume inspect "$volume_name" >/dev/null 2>&1; then
  docker volume create "$volume_name"
fi

container_name="dynamodb-randomizer"
host_port=8000

docker run \
  --rm -d --name "$container_name" \
  -p "$host_port":8000 \
  -v "$volume_name":/var/lib/randomizer \
  amazon/dynamodb-local -jar DynamoDBLocal.jar -dbPath /var/lib/randomizer
