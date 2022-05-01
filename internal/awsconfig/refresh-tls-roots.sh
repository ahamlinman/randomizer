#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")" || exit 1

# https://www.amazontrust.com/repository/
sources=(
  https://www.amazontrust.com/repository/AmazonRootCA{1..4}.pem
  https://www.amazontrust.com/repository/SFSRootCAG2.pem
)

set -x
exec wget -O cert.pem "${sources[@]}"
