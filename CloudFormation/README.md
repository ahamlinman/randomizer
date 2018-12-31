# Randomizer: AWS Lambda Deployment

The randomizer supports deployment to AWS Lambda, allowing you to set it up
without the need to manage servers or other infrastructure. This directory
includes tools and instructions that will help you perform the deployment.

If you don't already have an AWS account, sign up at https://aws.amazon.com/ to
get started.

## Set Up the AWS CLI

The deployment script in this directory uses the AWS CLI (the `aws` tool) to
perform all of its work.

To start, install the AWS CLI (the `aws` tool) onto your system if you don't
already have it. See [Installing the AWS Command Line Interface][install] in
the AWS CLI User Guide for a full walkthrough of how to install the tool with
`pip`. On some systems, there might be other (possibly easier) ways to install
the tool. For example, if you're using macOS and have [Homebrew][brew]
installed, you can simply run `brew install awscli`.

After installation, see [Configuring the AWS CLI][configure] in the AWS CLI
User Guide to set up access to your AWS account from the CLI. This requires a
set of credentials from AWS; the guide explains how to obtain these if you're
not already familiar with AWS [IAM][IAM].

[install]: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html
[brew]: https://brew.sh
[configure]: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html
[IAM]: https://aws.amazon.com/iam/

## Create an S3 Bucket

When AWS receives a request from Slack, it will need to retrieve the compiled
randomizer code from [S3][S3] in order to run it. To facilitate this, we'll
create a bucket to hold that compiled code.

First, pick a name for your bucket. The name must be unique across *all*
buckets in Amazon S3, so pick something fairly specific and note it down.

Then, use the AWS CLI (which we configured above) to create the bucket in your
AWS account: `aws s3 mb s3://<bucket name>`. Of course, you should replace the
`<bucket name>` placeholder with the actual name of your bucket (and you should
continue this pattern throughout the guide).

[S3]: https://aws.amazon.com/s3/

## Run the Initial Deployment

Now, we're ready to use AWS [CloudFormation][CloudFormation] to deploy the
randomizer into our account, with all necessary resources (e.g. the DynamoDB
table for storing groups) automatically created and configured.

Similar to how you picked a bucket name, you'll also need to pick a name for
your CloudFormation "stack." Unlike your bucket name, this only needs to be
unique for your specific account. If you only need to deploy one copy of the
randomizer, a simple name like "Randomizer" should be enough. Note this down
along with your bucket name.

With both of the above names and your verification token from Slack (see
README.md one level up) available, run the following command from this
directory:

```
./cf.sh build-deploy <bucket name> <stack name> --parameter-overrides SlackToken=<token>
```

This command will automatically compile the randomizer code for AWS Lambda,
upload it to your S3 bucket, and set it up for use. After some time, the script
will finish and print the webhook URL for Slack. Copy and paste this into the
"URL" field of the Slack slash command configuration, and save it.

At this point, you should be able to use the randomizer in your Slack
workspace. Go ahead and try it out!

[CloudFormation]: https://aws.amazon.com/cloudformation/

## Upgrades and Maintenance

To upgrade the randomizer deployment in your AWS account, simply run the above
command inside a newer version of the randomizer repository without the Slack
token parameter override. For example:

```
./cf.sh build-deploy <bucket name> <stack name>
```

Run `./cf.sh help` to learn more about additional commands that might be
useful.

## Notes

* The CloudFormation template (Template.yaml) uses AWS [SAM][SAM] to simplify
  the setup of the Lambda function.
* `cf.sh` "packages" Template.yaml into Package.yaml before deployment to
  CloudFormation. This step involves uploading the Lambda handler binary to S3
  and replacing the local reference with an S3 URI.
* The DynamoDB table in the template is provisioned in On-Demand capacity mode.
  Note that this mode is not eligible for the AWS Free Tier. See the
  documentation for [Read/Write Capacity Mode][capacity mode] for more details.
* Estimating costs on AWS is never easy. Anecdotally, my Slack team at work
  (over 1,000 people) makes a little over 200 requests to the randomizer per
  month. Between the low volume and my relatively low use of AWS in general
  (letting me take advantage of free tiers on Lambda and DynamoDB even as an
  existing user), the randomizer is effectively free for me to run.

[SAM]: https://github.com/awslabs/serverless-application-model
[capacity mode]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.ReadWriteCapacityMode.html
