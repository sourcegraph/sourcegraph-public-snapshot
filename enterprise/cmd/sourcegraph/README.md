# Sourcegraph App

Sourcegraph App is a single-binary distribution of Sourcegraph that runs on your local machine.

**Status:** alpha (only for internal use at Sourcegraph)

## Development

```shell
sg start app
```

## Usage

Sourcegraph App is in alpha (only for internal use at Sourcegraph).

Check the **Sourcegraph App release** bot in [`#app`](https://app.slack.com/client/T02FSM7DL/C04F9E7GUDP) (in the Sourcegraph internal Slack) for the latest release information.

## Build and release

### Snapshot releases

> Sourcegraph App is in internal alpha and only has snapshot releases. There are no versioned or tagged releases yet.

To build and release a snapshot for other people to use, push a commit to the special `app/release-snapshot` branch:

```shell
git push -f origin HEAD:app/release-snapshot
```

This runs the `../../dev/app/release.sh` script in CI, which uses [goreleaser](https://goreleaser.com/) to build for many platforms, package, and publish to the `sourcegraph-app-releases` Google Cloud Storage bucket.

Check the build status in [Buildkite `app/release-snapshot` branch builds](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=app%2Frelease-snapshot).

### Local builds (without releasing)

To build it locally for all platforms (without releasing, uploading, or publishing it anywhere), run:

```shell
VERSION=0.0.0+dev enterprise/dev/app/release.sh --snapshot
```

The builds are written to the `dist` directory.

If you just need a local build for your current platform, run `sg start app` (as mentioned in the [Development](#development) section) and then grab the `.bin/sourcegraph` binary. This binary does not have the web bundle (JavaScript/CSS) embedded into it.
