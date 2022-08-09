#!/bin/sh
set -ex

docker="${DOCKER:-docker}"
dynamodb_image="${DYNAMODB_IMAGE:-amazon/dynamodb-local:latest}"
volume_name="${VOLUME_NAME:-dynamodb-randomizer}"
container_name="${CONTAINER_NAME:-dynamodb-randomizer}"
host_port="${HOST_PORT:-8000}"

if ! "$docker" volume inspect "$volume_name" >/dev/null 2>&1; then
  "$docker" volume create "$volume_name"
  "$docker" run --rm -v "$volume_name":/var/lib/randomizer -u 0:0 --entrypoint "" "$dynamodb_image" \
    chown dynamodblocal:dynamodblocal /var/lib/randomizer
fi

"$docker" run \
  --rm -d --name "$container_name" \
  -p "$host_port":8000 \
  -v "$volume_name":/var/lib/randomizer \
  "$dynamodb_image" -jar DynamoDBLocal.jar -dbPath /var/lib/randomizer
