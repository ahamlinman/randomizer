# Randomizer

```
go get go.alexhamlin.co/randomizer
```

The randomizer, quite simply, chooses random options out of lists.

More specifically, it can:

* Randomize the order of items in a list (naturally)
* Save "groups" of items for later use, in a local database or Amazon DynamoDB
* Run as a Slack slash command (on your own server or on AWS Lambda), or as a
  CLI tool for local testing

## Trying it Out

From the root of this repository, simply run `go build ./cmd/randomizer-demo`
to try out a local copy of the randomizer in your terminal. Start by running
`./randomizer-demo help` to see what you can do.

This is a good way to get a rough taste of what the command looks like in a
Slack workspace, especially if you'd like to quickly test out changes.

## Deploying to Your Server

From the root of this repository, run `go build ./cmd/randomizer-server`. The
resulting binary starts up a web server that provides a Slack slash command
endpoint at its root path.

The web server requires the `SLACK_TOKEN` environment variable to be set to the
token obtained from the settings page for your slash command. Run with `-help`
to see available options for the server.

## Deploying to AWS Lambda

See the README in the `CloudFormation/` directory for more information.

## Upcoming

* Watch my co-workers use this thing, and see what I can do to make it great.
* Write better READMEs, both here and for CloudFormation deployment.
