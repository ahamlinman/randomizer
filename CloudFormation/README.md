# Randomizer: AWS Lambda Deployment

The randomizer supports deployment to AWS Lambda, allowing you to set it up
without the need to manage servers or other infrastructure. This directory
includes tools and instructions that will help you perform the deployment.

If you don't already have an AWS account, sign up at https://aws.amazon.com/ to
get started.

## Install and Configure Required Tools

In addition to a working Go installation, the deployment script requires the
[AWS CLI][install-aws-cli]. Versions 1 and 2 should both work. If you happen to
be using [Homebrew][brew], you can install the AWS CLI with a single command:

```sh
brew install awscli
```

After installing the AWS CLI, see [Configuring the AWS CLI][configure] to set up
access to your AWS account. This requires a set of credentials from AWS; the
guide explains how to obtain these if you're not already familiar with [AWS
IAM][iam].

(TODO: Discuss what IAM policies the CLI user needs to have.)

[install-aws-cli]: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html
[brew]: https://brew.sh
[configure]: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html
[iam]: https://aws.amazon.com/iam/

## Create an ECR Registry

When AWS Lambda starts up your function, it will download the compiled
randomizer code from [Amazon Elastic Container Registry][ecr]. ECR stores
versions of the randomizer code in a "repository" in your account.

You can create a new ECR repository using the AWS CLI:

```sh
aws ecr create-repository --repository-name randomizer --image-tag-mutability IMMUTABLE
```

You can choose whatever `--repository-name` you'd like, as long as it's unique
within your AWS account. The `--image-tag-mutability` option is optional, but
provides an additional safeguard to prevent accidentally overwriting your live
randomizer code outside of the deployment script.

[ecr]: https://aws.amazon.com/ecr/

## Run the Initial Deployment

Now, we're ready to use AWS [CloudFormation][CloudFormation] to deploy the
randomizer into our account, with all necessary resources (e.g. the DynamoDB
table for storing groups) automatically created and configured.

Similar to how you picked a ECR repository name, you'll also need to pick a name
for your CloudFormation "stack." Like your repository name, this needs to be
unique within your AWS account. If you only need to deploy one copy of the
randomizer, a simple name like "Randomizer" should be enough.

With both of the above names and your verification token from Slack available
(see README.md one level up), run the following command from this directory:

```
./cf.sh build-deploy <repository name> <stack name> SlackToken=<token>
```

This command will automatically compile the randomizer code for AWS Lambda,
upload it to your ECR repository, and set it up for use. After some time, the
script will finish and print the webhook URL for Slack. Copy and paste this into
the "URL" field of the Slack slash command configuration, and save it.

At this point, you should be able to use the randomizer in your Slack
workspace. Go ahead and try it out!

[CloudFormation]: https://aws.amazon.com/cloudformation/

## Upgrades and Maintenance

To upgrade the randomizer deployment in your AWS account, run the above command
inside a newer version of the randomizer repository without the Slack token
parameter override. For example:

```
./cf.sh build-deploy <repository name> <stack name>
```

Run `./cf.sh help` to learn more about additional commands that might be
useful.

## Notes

* The deployment script runs [zeroimage][zeroimage] with `go run` to upload the
  compiled randomizer binary as a container image to your ECR repository.
* The CloudFormation template (Template.yaml) uses the [AWS SAM][sam]
  transformation to simplify the setup of the Lambda function.
* The DynamoDB table in the template is provisioned in On-Demand capacity mode.
  Note that this mode is not eligible for the AWS Free Tier. See the
  documentation for [Read/Write Capacity Mode][capacity mode] for more details.
* The default configuration enables [AWS X-Ray][x-ray] tracing for the function
  and its requests to DynamoDB. X-Ray is free for up to 100,000 traces per month
  for every AWS account, and it's useful to see where each request is spending
  time. However, you can turn it off by passing `XRayTracingEnabled=false` to
  the deployment script.
* Estimating costs on AWS is never easy. Anecdotally, my Slack team at work
  (over 1,000 people) makes a little over 200 requests to the randomizer per
  month. Between the low volume and my relatively low use of AWS in general
  (letting me take advantage of free tiers on Lambda and DynamoDB even as an
  existing user), the randomizer is effectively free for me to run.

[zeroimage]: https://github.com/ahamlinman/zeroimage
[sam]: https://github.com/awslabs/serverless-application-model
[capacity mode]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.ReadWriteCapacityMode.html
[x-ray]: https://aws.amazon.com/xray/
