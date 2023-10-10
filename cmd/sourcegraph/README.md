# Cody App

Cody App is a single-binary distribution of Sourcegraph that runs on your local machine.

**Status:** alpha (only for internal use at Sourcegraph)

## Development

```shell
sg start app
```

(Or `sg start single-program` if you want all of Sourcegraph, not just Cody functionality.)

If your are running app on a fresh database instance you also have to perform the following steps:

- After opening the web app you will be directed to `/sign-in`, NOT the local repo setup step that is shown in production.
- Select to sign up with a new user account (the link following the log in options).
- Give this user account site admin privileges by running `psql -h ~/.sourcegraph-psql -U sourcegraph sourcegraph -c 'update users set site_admin=true'`

## Usage

Cody App is in alpha (only for internal use at Sourcegraph).

Check the **Cody App release** bot in [`#app`](https://app.slack.com/client/T02FSM7DL/C04F9E7GUDP) (in the Sourcegraph internal Slack) for the latest release information.

## Build and release

### Local builds (without releasing)

To build it locally for all platforms (without releasing, uploading, or publishing it anywhere), run:

```shell
VERSION=0.0.0+dev dev/app/release.sh --snapshot
```

The builds are written to the `dist` directory.

If you just need a local build for your current platform, run `sg start app` (as mentioned in the [Development](#development) section) and then grab the `.bin/sourcegraph` binary. This binary does not have the web bundle (JavaScript/CSS) embedded into it.

### CI builds (releasing)

Our CI pipeline runs the `../../dev/app/release.sh` script, which uses [goreleaser](https://goreleaser.com/) to build for many platforms, package, and publish to the `sourcegraph-app-releases` Google Cloud Storage bucket.

#### Insiders

Insiders builds may be created from non-`main` branches; they are much more experimental. To create one push your current branch to the special `app/insiders` branch:

```shell
git push -f origin HEAD:app/insiders
```

Check the build status in [Buildkite `app/insiders` branch builds](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=app%2Finsiders).

### Official releases

Official releases may only be created from `main` (or sometimes `main-app` during a code freeze); to create one push the latest remote branch to the special `app/release` branch:

```shell
git fetch
git push -f origin origin/main:app/release
```

Check the build status in [Buildkite `app/release` branch builds](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=app%2Frelease).
