# BUILD TRACKER

Build Tracker is a server that listens for build events from Buildkite, stores them in Redis and sends notifications about builds if they've failed.

The server currently listens for two events:

- `build.finished`
- `job.finished`

For each `job.finished` event that is received, the corresponding `build` is updated with the job that has finished. On receipt of a `build.finished` event, the server will determine if the build has failed by going through all the contained jobs of the build. If one or more jobs have indeed failed, a notification will be sent over slack. As well as this, the server will trigger a Buildkite job to process CI and Bazel data for the build for analytics purposes.

## Deployment infrastructure

Build Tracker is deployed in MSP. See the auto-generated [Notion doc](https://www.notion.so/sourcegraph/Build-Tracker-infrastructure-operations-bd66bf25d65d41b4875874a6f4d350cc#711a335bc7554738823293334221a18b) for details around accessing the environment and observability systems.

It is fine to wipe Redis if there are any issues stemming from data inconsistencies, redsync lock problems etc.

## Notification testing

To test the notifications that get sent over slack you can pass the flag `-RunSlackIntegrationTest` as part of your test invocation, with some required configuration:

```sh
export SLACK_TOKEN='my valid token'
export BUILDKITE_WEBHOOK_TOKEN='optional'
export GITHUB_TOKEN='optional'
go test . -RunSlackIntegrationTest
```

You can enable Slack client debugging by exporting the following environment variable `BUILD_TRACKER_SLACK_DEBUG=1`
