#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

source .cf.local.bash

binary_name="${binary_name:-$(basename "$go_src")}"
tarball_name="${tarball_name:-$binary_name.tar}"

usage () {
  cat <<EOF
cf.sh - Manage $project deployments with AWS CloudFormation

You must install the AWS CLI and Skopeo to use this script.

$0 build
  (Re)build the container image that will be deployed to AWS Lambda.

$0 upload <ECR repository>
  Upload the built container image to the provided ECR repository with Skopeo
  using a unique tag.

$0 deploy <stack name> [overrides...]
  Deploy the latest version of the CloudFormation stack using the latest
  container image, then print information about the deployed stack. Additional
  arguments are optional "Key=Value" parameter overrides. This stack supports
  the following overrides:

$params_usage

$0 build-deploy <ECR repository> <stack name> [overrides...]
  Build, upload, and deploy all in one step.

$0 current-image <stack name>
  Display the container image that the provided CloudFormation stack is
  currently configured to run.

$0 clean-repository <ECR repository> <stack names...>
  Remove all tags from the ECR repository that are not used by one of the listed
  CloudFormation stacks.

$0 help
  Print this message.
EOF
}

build () (
  os=linux
  arch=arm64

  set -x

  CGO_ENABLED=0 GOOS=$os GOARCH=$arch \
    go build -v \
    -ldflags='-s -w' \
    -o "$binary_name" \
    "$go_src"

  go run go.alexhamlin.co/zeroimage@main build \
    --target-os "$os" \
    --target-arch "$arch" \
    --output "$tarball_name" \
    "$binary_name"
)

upload () (
  if ! type skopeo &>/dev/null; then
    echo "must install skopeo to upload container images" 1>&2
    return 1
  fi
  if [ ! -s "$tarball_name" ]; then
    echo "must build a container image before uploading" 1>&2
    return 1
  fi

  repository_name="$1"
  repository="$(aws ecr describe-repositories \
    --repository-names "$repository_name" \
    --query 'repositories[0].repositoryUri' \
    --output text)"
  registry="${repository%%/*}"
  tag="$(date +%s)"
  image="$repository:$tag"

  set -x
  if ! skopeo list-tags docker://"$repository" &>/dev/null; then
    aws ecr get-login-password \
    | skopeo login --username AWS --password-stdin "$registry"
  fi

  skopeo copy oci-archive:"$tarball_name" docker://"$image"
  echo "$image" > latest-image.txt
)

deploy () (
  if [ ! -f latest-image.txt ]; then
    echo "must upload a container image before deploying" 1>&2
    return 1
  fi

  stack_name="$1"
  shift

  (
    set -x
    aws cloudformation deploy \
      --template-file Template.yaml \
      --capabilities CAPABILITY_IAM \
      --stack-name "$stack_name" \
      --no-fail-on-empty-changeset \
      --parameter-overrides \
          ImageUri="$(cat latest-image.txt)" \
          "$@"
  )

  echo
  print-stack-output "$stack_name"
)

build-deploy () {
  local ecr_repository="$1"
  local stack_name="$2"
  shift 2

  build
  upload "$ecr_repository"
  deploy "$stack_name" "$@"
}

current-image () (
  stack_name="$1"
  aws cloudformation describe-stacks \
    --stack-name "$stack_name" \
    --output text \
    --query "Stacks[0].Parameters[?ParameterKey=='ImageUri'].ParameterValue"
)

clean-repository () (
  repository_name="$1"
  shift

  if [ "$#" -lt 1 ]; then
    echo "must specify at least one stack name" 1>&2
    return 1
  fi

  all_tags_list="$(
    aws ecr describe-images --repository-name "$repository_name" \
      --output text --query 'imageDetails[].imageTags[] | map(&[@], @)' \
    | sort -n
  )"

  declare -a keep_tags
  for stack in "$@"; do
    image="$(current-image "$stack")"
    keep_tags+=("${image##*:}")
  done
  keep_tags_list="$(printf '%s\n' "${keep_tags[@]}" | sort -nu)"

  n_deleted_tags=0
  delete_tags_cmd=(aws ecr batch-delete-image \
    --repository-name "$repository_name" \
    --image-ids)
  while read -r tag; do
    ((n_deleted_tags+=1))
    delete_tags_cmd+=("imageTag=$tag")
  done < <(comm -23 <(echo "$all_tags_list") <(echo "$keep_tags_list"))

  if [ "$n_deleted_tags" -eq 0 ]; then
    echo "Repository is clean enough; no tags to delete."
    return
  fi

  echo "Will keep the following tags used by the listed stacks:"
  echo "$keep_tags_list"
  echo
  echo "Will run the following command to delete all other tags:"
  echo "${delete_tags_cmd[*]}"
  echo
  read -rp "Press Enter to run, or Ctrl-C to quit..."

  set -x
  "${delete_tags_cmd[@]}"
)

cmd="${1:-help}"
[ "$#" -gt 0 ] && shift

case "$cmd" in
  build)
    build
    ;;
  upload)
    upload "$@"
    ;;
  deploy)
    deploy "$@"
    ;;
  build-deploy)
    build-deploy "$@"
    ;;
  current-image)
    current-image "$@"
    ;;
  clean-repository)
    clean-repository "$@"
    ;;
  help)
    usage
    ;;
  *)
    usage
    exit 1
    ;;
esac
