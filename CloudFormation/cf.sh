#!/usr/bin/env bash

set -euo pipefail

usage () {
  cat <<EOF
cf.sh - Deploy the randomizer to your AWS account using CloudFormation

The AWS CLI must be installed and configured to use this script. For details
about configuration, see the AWS Documentation:

https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html

$0 build
  (Re)build the container image that will be deployed to AWS Lambda.

$0 upload <ECR repository>
  Upload the built container image to the provided ECR repository with Skopeo
  using a unique tag.

$0 deploy <stack name> [args...]
  Deploy the latest version of the CloudFormation stack using the latest
  container image, then print the URL for the deployed API.

  Additional arguments are passed directly to "aws cloudformation deploy". When
  deploying the stack for the first time, pass "--parameter-overrides
  SlackToken=<token>" to set the token used to authenticate requests from Slack.

$0 build-deploy <ECR repository> <stack name> [args...]
  Build, upload, and deploy all in one step.

$0 help
  Print this message.
EOF
}

build () (
  os=linux
  arch=arm64

  set -x

  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -v \
    -ldflags='-s -w' \
    -o randomizer-lambda \
    ../cmd/randomizer-lambda

  go run zeroimage.go \
    -entrypoint randomizer-lambda \
    -os "$os" -arch "$arch" \
    -output randomizer-lambda.tar
)

upload () (
  if ! type skopeo &>/dev/null; then
    echo "must install skopeo to upload container images" 1>&2
    return 1
  fi
  if [ ! -s randomizer-lambda.tar ]; then
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
  if ! skopeo login --get-login "$registry" &>/dev/null; then
    aws ecr get-login-password | skopeo login --username AWS --password-stdin "$registry"
  fi

  skopeo copy oci-archive:randomizer-lambda.tar docker://"$image"
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
      --parameter-overrides ImageUri="$(cat latest-image.txt)" \
      "$@"
  )

  echo -e "\\nThe Slack webhook is available at the following URL:"
  aws cloudformation describe-stacks \
    --stack-name "$stack_name" \
    --output text \
    --query 'Stacks[0].Outputs[0].OutputValue'
)

build-deploy () {
  local ecr_repository="$1"
  local stack_name="$2"
  shift 2

  build
  upload "$ecr_repository"
  deploy "$stack_name" "$@"
}

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
  help)
    usage
    ;;
  *)
    usage
    exit 1
    ;;
esac
