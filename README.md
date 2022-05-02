# Randomizer

The randomizer is a slash command webhook service for Slack that randomizes the
order of items in a list.

Not sure what to get for lunch?

> **/randomize** salad sandwich ramen

Need a code review from a teammate?

> **/randomize /save** myteam Alice Bob Carol Dave Eve
>
> **/randomize** myteam

Whenever you're unsure, let the universe decide!

## Try the Demo

You'll need the [Go][go] toolchain installed to try the demo program.

1. Clone this repository and `cd` into it
1. Build the demo: `go build ./cmd/randomizer-demo`
1. See what to do next: `./randomizer-demo help`

The demo outputs the same text that would be sent to Slack, using proper
[message formatting][format]. Groups are saved in a [bbolt][bbolt] database
file in the current directory. This gives a rough taste of how the command
works, and is helpful for testing.

[go]: https://golang.org/
[format]: https://api.slack.com/docs/message-formatting
[bbolt]: https://go.etcd.io/bbolt

## Serverless Deployment with AWS Lambda

See `SERVERLESS.md`.

## Notes on Configuring Your Own Server

To run the randomizer on your own server, you'll need to pick a storage backend
for groups: A local [bbolt][bbolt] database file, or Amazon DynamoDB. The bbolt
database requires no additional setup, but is locked by a single API server
process at a time (meaning that it will be impossible to do things like
zero-downtime deployment, should you desire).

To use DynamoDB, you'll need to create a table in your AWS account, along with
appropriate IAM credentials and policies. See the `GroupsTable` resource in
`CloudFormation/Template.yaml` for the required DynamoDB table schema.

The following environment variables are required:

- **Always:** One of the following:
  - `SLACK_TOKEN`: Set to the value of the verification token obtained above.
  - `SLACK_TOKEN_SSM_PATH`: Set to the name of an AWS SSM Parameter Store
    parameter containing the Slack verification token. With this option, you
    can use `SLACK_TOKEN_SSM_TTL` set to a Go duration to control how long the
    SSM lookup is cached for (default 2m).
- **bbolt only:** `DB_PATH`: Set to the desired location of the database file
  on disk (defaults to "randomizer.db" in the current directory).
- **DynamoDB only:** `DYNAMODB_TABLE`: Set to the name of the DynamoDB table to
  use in your AWS account.
- **When using SSM or DynamoDB:** [AWS environment variables][AWS vars]
  referencing appropriately configured IAM user or role credentials.

To get the API server binary, run `go build ./cmd/randomizer-server`. You can
also build and run a Docker image using the provided `Dockerfile`. By default,
the server listens on port 7636; this can be changed with a command line flag.
Run the server with `-help` for more details.

Topics not covered by these brief notes include:

- Proper service management, whether through Docker, a container orchestrator
  (Docker Swarm, Kubernetes, etc.), or a more traditional service manager
  (systemd, etc.).
- Setting up a reverse proxy to provide SSL termination for the randomizer API
  (as randomizer-server does not serve TLS out of the box).
- The specifics of configuring AWS IAM credentials and policies, adding the
  verification token to SSM, or creating the DynamoDB table, should you wish to
  use any of those features.

If you're uncomfortable with these topics (and aren't interested in learning
them), the AWS Lambda deployment option might be preferable.

[AWS vars]: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
