# Randomizer: Serverless Edition

This directory provides an AWS SAM template and a basic script to help you
deploy the randomizer slash command as a Lambda function in your AWS account.

**WARNING:** This README is going to be *extremely* rough. If you're not
already somewhat familiar with both AWS and Slack, you'll probably be lost. (I
hope to improve this in the future.)

First, install and configure the AWS CLI.

Then, create an S3 bucket to hold your packaged Go binaries for use by AWS
Lambda: `aws s3 mb s3://<bucket name>`

Then, go into Slack and start creating a new configuration for a slash command.
Copy the token from the authentication screen.

Finally, run the `cf.sh` script in this directory, filling in the various
fields as necessary:

```
./cf.sh build-deploy <stack name> <bucket name> --parameter-overrides SlackToken=<token>
```

When finished, the deployment script will print out a URL. Copy and paste it
into the slash command configuration screen, then save the slash command.

To re-deploy updated versions of the app, simply run `cf.sh` again with the
stack name and S3 bucket name. CloudFormation remembers the value of the Slack
authentication token on subsequent deployments.
