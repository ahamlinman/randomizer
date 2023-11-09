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

## Deployment Options

This repo provides two guides on deploying the randomizer API for use with
Slack:

- `SERVERLESS.md` is a fairly detailed walkthrough for deployment on [AWS
  Lambda][lambda], Amazon's managed function as a service platform.
- `SERVERMORE.md` is a fairly high-level guide for setting up the
  `randomizer-server` HTTP server, that assumes more background knowledge
  and/or willingness to dive into details of both standard server management
  and the randomizer implementation.

[lambda]: https://aws.amazon.com/lambda/
