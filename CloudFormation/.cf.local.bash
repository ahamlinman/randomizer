# This file is meant to be sourced.
# shellcheck disable=SC2034

project=randomizer
go_src=../cmd/randomizer-lambda

params_usage="$(cat <<EOF
  SlackTokenSSMName=<name>  The name of the AWS SSM parameter that stores the
                            Slack verification token. No default.

  SlackTokenSSMTTL=2m       The time to cache a successful lookup of the token
                            from SSM, as a Go duration.

  XRayTracingEnabled=true   If "true", enable AWS X-Ray tracing for the
                            function and its requests to AWS services.
EOF
)"

print-stack-output () (
  stack_name="$1"
  echo "The Slack webhook is available at the following URL:"
  aws cloudformation describe-stacks \
    --stack-name "$stack_name" \
    --output text \
    --query 'Stacks[0].Outputs[0].OutputValue'
)
