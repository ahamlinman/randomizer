#!/usr/bin/env bash
set -euo pipefail

usage () {
  cat <<EOF
cf.sh - Deploy the randomizer to your AWS account using CloudFormation

You must install the AWS CLI and Skopeo to use this script. See README.md for
details.

$0 build
  (Re)build the container image that will be deployed to AWS Lambda.

$0 upload <ECR repository>
  Upload the built container image to the provided ECR repository with Skopeo
  using a unique tag.

$0 deploy <stack name> [overrides...]
  Deploy the latest version of the CloudFormation stack using the latest
  container image, then print the URL for the deployed API.

  Additional arguments are passed to the "--parameter-overrides" option of "aws
  cloudformation deploy". When deploying the stack for the first time, pass
  "SlackToken=<token>" to set the token used to authenticate requests from
  Slack.

$0 build-deploy <ECR repository> <stack name> [overrides...]
  Build, upload, and deploy all in one step.

$0 current-image <stack name>
  Display the container image that the provided CloudFormation stack is
  currently configured to run.

$0 clean-repository <ECR repository>
  Remove all tags from the ECR repository, except for the 2 newest uploads in
  the repository or the last known image uploaded by this script.

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
    -o randomizer-lambda \
    ../cmd/randomizer-lambda

  go run go.alexhamlin.co/zeroimage@main \
    -os $os -arch $arch \
    randomizer-lambda
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
  if ! skopeo list-tags docker://"$repository" &>/dev/null; then
    aws ecr get-login-password \
    | skopeo login --username AWS --password-stdin "$registry"
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
      --parameter-overrides \
          ImageUri="$(cat latest-image.txt)" \
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

current-image () (
  stack_name="$1"
  aws cloudformation describe-stacks \
    --stack-name "$stack_name" \
    --output text \
    --query "Stacks[0].Parameters[?ParameterKey=='ImageUri'].ParameterValue"
)

clean-repository () (
  repository_name="$1"
  all_tags_list="$(
    aws ecr describe-images --repository-name "$repository_name" \
      --output text --query 'imageDetails[].imageTags[] | map(&[@], @)' \
    | sort -n
  )"

  keep_tags_list="$(echo "$all_tags_list" | tail -n2)"
  if [ -f latest-image.txt ]; then
    current_image="$(cat latest-image.txt)"
    current_tag="${current_image##*:}"
    keep_tags_list="$(printf '%s\n' "$keep_tags_list" "$current_tag" | sort -nu)"
  fi

  n_drop_tags=0
  drop_tags_cmd=(aws ecr batch-delete-image \
    --repository-name "$repository_name" \
    --image-ids)
  while read -r tag; do
    ((n_drop_tags+=1))
    drop_tags_cmd+=("imageTag=$tag")
  done < <(comm -23 <(echo "$all_tags_list") <(echo "$keep_tags_list"))

  if [ "$n_drop_tags" -eq 0 ]; then
    echo "Repository is clean enough; no tags to drop."
    return
  fi

  echo "Will keep the following tags:"
  echo "$keep_tags_list"
  echo
  echo "Will run the following command to drop all other tags:"
  echo "${drop_tags_cmd[*]}"
  echo
  read -rp "Press Enter to run, or Ctrl-C to quit..."

  set -x
  "${drop_tags_cmd[@]}"
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
