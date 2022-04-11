# Deployment Lag Notifer
https://github.com/sourcegraph/sourcegraph/issues/32878

The code for this is largely inspired by [Deployment Notifier](../deployment-notifer/README.md).

## Usage

## Flags

## How it works
This is inspired by `sg live cloud`. It is run on a [fixed schedule]() to check that code deployed to Cloud is recent. If it detects that a deployment version on Cloud differs dramatically from the tip of `sourcegraph/sourcegraph@main`, an alert will be sent to a Slack channel
