#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")" || exit 1

# https://www.amazontrust.com/repository/
sources=(
  https://www.amazontrust.com/repository/AmazonRootCA{1..4}.cer
  https://www.amazontrust.com/repository/SFSRootCAG2.cer
)

set -x
exec wget -O amazon-trust.cer "${sources[@]}"
