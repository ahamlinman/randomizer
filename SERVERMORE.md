# High-Level Notes on Configuring `randomizer-server`

`randomizer-server` is an HTTP server providing the Slack slash command API for
the randomizer. This section provides general pointers on setting it up.

This guide **doesn't** cover:

- Proper server configuration and service management, whether directly on a
  server (systemd, Docker, Podman, etc.) or through an orchestrator
  (Kubernetes, HashiCorp Nomad, Docker Swarm, etc.).
- Setting up a reverse proxy to provide SSL termination for the randomizer API
  (as `randomizer-server` doesn't serve TLS out of the box).
- The specifics of configuring cloud credentials and access policies, should
  you wish to use a cloud provider for group storage or secrets.

If you're uncomfortable with these topics or the rest of this guide, and you
aren't interested in learning them, the AWS Lambda setup might be preferable
(see `SERVERLESS.md`).

In addition to the environment variables below, check `randomizer-server -help`
for CLI flags that you may wish to set, like the bind address for the server
(defaults to ":7636").

## Slack Token

Regardless of the group storage backend, you'll need to configure one of the
following environment variables to set the Slack slash command verification
token (the newer signing secret configuration isn't supported):

- `SLACK_TOKEN`: Set to the value of the token itself.
- `SLACK_TOKEN_SSM_PATH`: The path to an AWS SSM Parameter Store parameter
  containing the value of the verification token. This requires appropriate AWS
  configuration in the environment. You can also set `SLACK_TOKEN_SSM_TTL` to a
  Go duration to control how long the SSM lookup remains cached (default 2m).

## Storage Backends

### bbolt

[bbolt][bbolt] is the local key-value database engine behind systems like etcd,
Consul, InfluxDB 2.x, and more.

The bbolt backend's only prerequisite is persistent disk storage. The database
file will be locked by a single server process at a time, so note that this
backend won't support things like high availability or zero-downtime
deployment.

The bbolt backend is active by default if no other backends have environment
variables set. You can set `DB_PATH` to the desired location of the database on
disk, otherwise it defaults to "randomizer.db" in the current directory.

[bbolt]: https://go.etcd.io/bbolt

### DynamoDB

DynamoDB is Amazon's fully managed NoSQL key-value store.

The DynamoDB backend requires a pre-existing table with the randomizer schema.
The `randomizer-dbtools dynamodb create` command in this repo can help you set
this up. You can also reference `GroupsTable` in `CloudFormation.yaml`.

To activate the DynamoDB backend, set `DYNAMODB_TABLE` to the name of the
table. You may also need to configure [environment variables for the AWS
SDK][AWS vars]. (Note that other environment variables associated with DynamoDB
in the code are unstable, and are subject to removal or behavior changes.)

[AWS vars]: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html

### Google Cloud Firestore

Firestore is Google's fully managed document database. This mode is especially
useful to run `randomizer-server` on [Cloud Run][Cloud Run], with a level of
operational ease comparable to the AWS Lambda solution (though the randomizer
doesn't include infrastructure-as-code for it out of the box).

The Firestore backend requires a pre-existing database in a Google Cloud
project (only Native mode been tested, but Datastore mode may work too).
Note that the randomizer expects to use the full database, and that only the
"(default)" database in each Google Cloud project is eligible for the Firestore
free tier.

To activate the Firestore backend, set both of the following variables:

- `FIRESTORE_PROJECT_ID`: The ID (not name) of your Google Cloud project.
- `FIRESTORE_DATABASE_ID`: The ID of the database to use (for example,
  "(default)").

You may also need to configure [Application Default Credentials][ADC] for the
Google Cloud SDK.

[Cloud Run]: https://cloud.google.com/run
[ADC]: https://cloud.google.com/docs/authentication/application-default-credentials
