#!/usr/bin/env bash

set -euo pipefail

usage () {
  cat <<EOF
cf.sh - Deploy the randomizer to your AWS account using CloudFormation

The AWS CLI must be installed and configured to use this script. For details
about configuration, see the AWS Documentation:

https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html

$0 build
  (Re)build the Go binary that will be deployed to AWS Lambda.

$0 deploy <stack name> <S3 bucket> [args...]
  Deploy the randomizer using CloudFormation. This will upload the built Go
  binary to the provided S3 bucket, run a deployment, and print the URL for the
  deployed API.

  Additional arguments are passed directly to "aws cloudformation deploy". In
  particular, when deploying the stack for the first time, use
  "--parameter-overrides SlackToken=<token>" to set the token used to
  authenticate requests from Slack.

$0 build-deploy <stack name> <S3 bucket> [args...]
  Just run "build" and "deploy" together.

$0 clean
  Clean up the Go binary and CloudFormation package template.

$0 clean-bucket <S3 bucket> [args...]
  Remove all but the most recent 3 files from the provided S3 bucket. This is
  useful for cleaning up old Lambda deployment packages created by the deploy
  command. Additional arguments are passed directly to each instance of "aws s3
  rm".

$0 help
  Print this message.
EOF
}

build () (
  set -x

  mkdir -p dist
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags='-s -w' \
    -o dist/slack-lambda-handler \
    ../cmd/slack-lambda-handler
)

deploy () (
  if [ ! -d dist ]; then
    echo "must build the handler binary before deployment" 1>&2
    return 1
  fi

  stack_name="$1"
  s3_bucket="$2"
  shift 2

  set -x

  aws cloudformation package \
    --template-file Template.yaml \
    --output-template-file Package.yaml \
    --s3-bucket "$s3_bucket"

  aws cloudformation deploy \
    --template-file Package.yaml \
    --capabilities CAPABILITY_IAM \
    --stack-name "$stack_name" \
    --no-fail-on-empty-changeset \
    "$@"

  set +x

  echo -e "\\nThe Slack webhook is available at the following URL:"
  aws cloudformation describe-stacks \
    --stack-name "$stack_name" \
    --output text \
    --query 'Stacks[0].Outputs[0].OutputValue'
)

build-deploy () {
  build
  deploy "$@"
}

clean () (
  set -x
  rm -rf ./dist ./Package.yaml
)

clean-bucket () (
  bucket="$1"
  shift

  old_files="$(aws s3 ls "s3://$bucket" \
    | sort | head -n-3 \
    | awk "{ print \"s3://$bucket/\" \$4 }")"

  set -x
  for f in $old_files; do
    aws s3 rm "$@" "$f"
  done
)

cmd="${1:-help}"
[ "$#" -gt 0 ] && shift

case "$cmd" in
  build)
    build
    ;;
  deploy)
    deploy "$@"
    ;;
  build-deploy)
    build-deploy "$@"
    ;;
  clean)
    clean
    ;;
  clean-bucket)
    clean-bucket "$@"
    ;;
  help)
    usage
    ;;
  *)
    usage
    exit 1
    ;;
esac
