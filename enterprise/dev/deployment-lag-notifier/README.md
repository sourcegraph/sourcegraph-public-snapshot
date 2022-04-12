# Deployment Lag Notifer
https://github.com/sourcegraph/sourcegraph/issues/32878

The code for this is largely inspired by [Deployment Notifier](../deployment-notifer/README.md).

## Usage
Run `make`. The message (as would be sent to slack) should be printed out to the console.

If you would like to actually post the message, either configure the `-slack-webhook-url` flag or set the `SLACK_WEBHOOK_URL` environment variable and then run `make prod`.

## Flags

* `-dry-run` Print to stdout instead of sending to Slack
* `-env` Environment to check against (default "cloud", options: "cloud", "k8s", "preprod")
* `-slack-webhook-url` (env var: `SLACK_WEBHOOK_URL`) Slack webhook URL to post to. To get a webhook URL, add a configuration [here](https://sourcegraph.slack.com/apps/A0F7XDUAZ-incoming-webhooks?tab=settings&next_id=0):

## How it works
This is inspired by `sg live cloud`. It is run on a [fixed schedule]() to check that code deployed to Cloud is recent. If it detects that a deployment version on Cloud differs dramatically from the tip of `sourcegraph/sourcegraph@main`, an alert will be sent to a Slack channel
