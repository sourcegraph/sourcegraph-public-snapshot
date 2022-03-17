# Deployment Notifier

Deployment Notifier is a tool that analyzes changes in order to guess which pull requests have been deployed in a given environment.
It is meant to be included in deploymement pipeline to automatically notify teammates that their changes have been deployed.

## Usage

In a deployment repository (such as `sourcegraph/deploy-sourcegraph-*`) you can run the following command to post GitHub comments and Slack notifications.

```sh
deployment-notifier -environment $MY_ENV -sourcegraph.commit $MY_COMMIT
```

### Flags 

- `-github.token` (defaults to `$GITHUB_TOKEN`) 
- `-sourcegraph.commit` (optional) the SHA1 of the commit being deployed, used to find the pull requests to mention.
- `-sourcegraph.guess-commit` (optional) infers the SHA1 of the commit being deployed from the changes, supersedes `-sourcegraph.commit`.
- `-environment` either `preprod` or `production`
- `-pretend` (optional) do not post on Slack or GitHub, just print out what would be posted.
- `-slack.token` Slack Token used to find the matching Slack handle for pull request authors.
- `-slack.webhook` Slack webhook URL to post the notifications on.

## How it works

Deployment notifier works as following:

1. Request the commit currently running on the target environment
2. Compute the deployed applications based on the changes included in the latest commit.
3. Find all PRs that happened in between the commit found in 1. and the commit being currently deployed (`-sourcegraph.commit`)
4. Post a comment in each of those PRs with the list of applications computed in 2.
5. Post a Slack message pinging all pull request authors.

### Building it for usage in the `deploy-sourcegraph-*` repositories

Assuming you have correct GCP credentials set:

```sh 
./release.sh
```

## Contributing

### Local usage

To avoid spamming comments, Deployment Notifier comes with a few flags to help:

- `-pretend` prints a summary of what would be posted on the standard output.
- `-mock.live-commit` skips the request to a live environment and uses the given commit instead as the version.
  - This is useful to narrow down the found PR to a small number by pretending the last deploy is only a few commits away.

### Testing

The tests uses recorded responses from GitHub, to update the cassettes, uses the `-update` flag when running `go test`. Make sure
that the `GITHUB_TOKEN` environment is defined when doing so.
