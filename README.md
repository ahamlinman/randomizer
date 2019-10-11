# Randomizer

The randomizer is a slash command for Slack that randomizes the order of items
in a list.

Not sure what to get for lunch?

> **/randomize** burgers salad teriyaki

Need a code review from a teammate?

> **/randomize /save** myteam Alice Bob Charlie Dave Eve
>
> **/randomize** myteam

Whenever you're unsure, let the universe decide!

## Try the Demo

You'll need the [Go][go] toolchain installed to try the demo
program.

1. Clone this repository: `go get go.alexhamlin.co/randomizer`. (Or, with Go
   1.11+, just `git clone` to any location.)
1. Build the demo: `go build ./cmd/randomizer-demo`
1. See what to do next: `./randomizer-demo help`

The demo outputs the same text that would be sent to Slack, using proper
[message formatting][format]. Groups are saved in a [bbolt][bbolt] database
file in the current directory. This gives a rough taste of how the command
works, and is helpful for testing.

[go]: https://golang.org/
[format]: https://api.slack.com/docs/message-formatting
[bbolt]: https://go.etcd.io/bbolt

## Deploy to Your Slack Workspace

To start, you'll need to generate a slash command configuration in Slack and
get a verification token. Search for "Slash Commands" in your Slack app
directory, and add a new configuration (this is easier than creating a new
"app"). Choose a suitable name like `/randomize`, and continue. The token will
be listed on the following page. Copy it down for later.

Next, you'll need to make the randomizer's HTTP API available to fill in the
"URL" field of the configuration.

The easiest option is to run the API as a serverless application with AWS
Lambda, using the helpful templates and scripts in this repository. To deploy
the randomizer this way, continue with the README in the `CloudFormation/`
directory.

Alternatively, if you're comfortable with the brief configuration notes below,
you can run the API on your own server as a single binary or Docker container.

## Notes on Configuring Your Own Server

To run the randomizer on your own server, you'll need to pick a storage backend
for groups: A local [bbolt][bbolt] database file, or Amazon DynamoDB. The bbolt
database requires no additional setup, but is locked by a single API server
process at a time (meaning that it will be impossible to do things like
zero-downtime deployment, should you desire).

To use DynamoDB, you'll need to create a table in your AWS account, along with
appropriate IAM credentials / policies. See the `GroupsTable` resource in
`CloudFormation/Template.yaml` for the required DynamoDB table schema.

The following environment variables are required:

* **Always:** `SLACK_TOKEN`: Set to the verification token obtained above
* **bbolt only:** `DB_PATH`: Set to the desired location of the database
  file on disk (defaults to "randomizer.db" in the current directory)
* **DynamoDB only:**
   - `DYNAMODB_TABLE`: Set to the name of the DynamoDB table to use in your AWS
     account
   - AWS environment variables (for your IAM credentials) as described
     [here][AWS vars]

To get the API server binary, run `go build ./cmd/randomizer-server`. You can
also run the pre-built Docker image: `ahamlinman/randomizer`. By default, the
server listens on port 7636; this can be changed with a command line flag. Run
the server with `-help` for more details.

Topics not covered by these brief notes include:

* Proper service management, whether through Docker, a container orchestrator
  (Docker Swarm, Kubernetes, etc.), or a more traditional service manager
  (systemd, etc.)
* Setting up a reverse proxy to provide SSL termination for the randomizer API,
  should you desire this (hint: you very much should)
* The specifics of DynamoDB table creation / IAM configuration, should you wish
  to use DynamoDB

If you're uncomfortable with these topics (and aren't interested in learning
them), the AWS Lambda deployment option might be preferable.

[AWS vars]: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
