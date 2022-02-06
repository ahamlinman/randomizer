#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

# https://www.amazontrust.com/repository/
sources=(
  https://www.amazontrust.com/repository/AmazonRootCA{1..4}.pem
  https://www.amazontrust.com/repository/SFSRootCAG2.pem
)

wget -O cert.pem "${sources[@]}"
