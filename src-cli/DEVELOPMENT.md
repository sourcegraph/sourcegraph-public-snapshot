This file covers things that are good to know if you're developing or maintaining `src`. It is likely incomplete, and contributions would be most welcome!

## Contents

- [Developing `src`](#development)
- [Testing `src`](#testing)
- [Releasing `src`](#releasing)
- [How we use Docker](#docker)

## Development

If you want to develop the CLI, clone the repository and build/install it with `go install`

```
cd src-cli/
go install ./cmd/src
```

You can also run it directly without installation:

```
go run ./cmd/src
```

### Use `debug` build tag to debug batch changes functionality

Since `src batch apply` and `src batch preview` start up a TUI that gets updated repeatedly it's nearly impossible to do printf-debugging by printing debug information - the TUI would hide those or overwrite them.

To help with that, you can compile your src binary (or run the tests) with the `debug` build flag:

```
go build -tags debug -o ~/src ./cmd/src
```

This will cause the `./internal/batches/debug.go` file to be included in the build. In that file the `log` default package logger is setup to log to `~/.sourcegraph/src-cli.debug.log`.

That allows you to `tail -f ~/.sourcegraph/src-cli.debug.log` and use `log.Println()` in your code to debug.

## Testing

`src` uses normal Go testing patterns, and doesn't require any other dependencies for its test suite to be run.

You can test the entire tool with:

```sh
go test ./...
```

Or just a single package with:

```sh
go test ./internal/batches/workspace
```

We adhere to the [general Sourcegraph principles for testing](https://docs.sourcegraph.com/dev/background-information/testing_principles), as well as [the Go specific directions](https://docs.sourcegraph.com/dev/background-information/languages/testing_go_code), at least to the extent they apply to a standalone tool like `src`.

## Releasing

1.  Find the latest version (either via the releases tab on GitHub or via git tags) to determine which version you are releasing.
2.  (optional) If this is a non-patch release, update the changelog. Add a new section `## $MAJOR.MINOR` to [`CHANGELOG.md`](https://github.com/sourcegraph/src-cli/blob/main/CHANGELOG.md#unreleased) immediately under `## Unreleased changes`. Add new empty `Added`, `Changed`, `Fixed`, and `Removed` sections under `## Unreleased changes`. Open a pull request with the new changelog. Get the pull request merged before completing the next step.
3.  `VERSION=9.9.9 ./release.sh` (replace `9.9.9` with the version you are releasing)
4.  GitHub will automatically perform the release via the [Build and Release action](https://github.com/sourcegraph/src-cli/actions?query=workflow%3ABuild+and+Release). Once it has finished, **you need to confirm**:
    1. The brew formula shows the correct version.
       ```shell
        brew info sourcegraph/src-cli/src-cli
       ```
    2. The npm library shows the correct version.
       ```shell
       npm show @sourcegraph/src version
       ```
    3. The [releases section of the repo sidebar](https://github.com/sourcegraph/src-cli) shows the correct version.
5.  Make the necessary updates to the main Sourcegraph repo:
    1. Update the `MinimumVersion` constant in the [src-cli package](https://github.com/sourcegraph/sourcegraph/tree/main/internal/src-cli/consts.go).
    2. Update the reference documentation by running `go generate ./doc/cli/references`.
    3. Commit the changes, and open a PR.
6.  Once the version bump PR is merged and the commit is live on dotcom, check that the [curl commands in the README](README.md#installation) also fetch the new latest version.

### Patch releases

If a backwards-compatible change, such as a bug fix, is made _after_ a backwards-incompatible one, the backwards-compatible one should be re-released to older instances that support it.

A Sourcegraph instance returns the highest patch version with the same major and minor version as `MinimumVersion` as defined in the instance. Patch versions are reserved solely for non-breaking changes and minor bug fixes. This allows us to dynamically release fixes for older versions of `src-cli` without having to update the instance.

To release a bug fix or a new feature that is backwards compatible with one of the previous two minor version of Sourcegraph, cherry-pick the changes into a patch branch, then follow the steps above to release the patch, specifying the appropriate patch `VERSION`. The Build and Release action will automatically create the appropriate versioned release, without overwriting the latest one.

## Docker

We use Docker in a couple of different ways:

- `src` is pushed to [`sourcegraph/src-cli`](https://hub.docker.com/r/sourcegraph/src-cli) each [release](#releasing). More information on this [can be found below](#src-cli-docker-image).
- `src` uses [`sourcegraph/src-batch-change-volume-workspace`](https://hub.docker.com/r/sourcegraph/src-batch-change-volume-workspace) when executing batch changes on the `volume` workspace, which is the default on macOS. This image is also updated on [release](#releasing). More information on this [can also be found below](#dependent-docker-images).

### `src-cli` Docker image

Each release of `src` results in a new tag of the [`sourcegraph/src-cli` Docker image](https://hub.docker.com/r/sourcegraph/src-cli) being pushed to Docker Hub. This is handled by [goreleaser's Docker support](https://goreleaser.com/customization/docker/).

The main gotcha here is that the way goreleaser builds the Docker image is fairly difficult to replicate from the desktop: it builds a `src` binary without any runtime libc dependencies that can be installed in a `scratch` image, but unless you work on Alpine, your desktop is _not_ configured to build Go binaries like that.

As a result, there are two Dockerfiles in this project. goreleaser uses `Dockerfile.release`, which is replicated in a multi-stage `Dockerfile` that builds `src` in a builder container to ensure it's built in a way that can be tested.

#### Testing the Docker image

If you need to test a change to the Dockerfile (for example, due to a Renovate PR bumping the base image), you should pull that change, then build a local image with something like:

```bash
docker build -t local-src-cli .
```

After which you should be able to run:

```bash
docker run --rm local-src-cli
```

and get the normal help output from `src`.

### Dependent Docker images

`src batch apply` and `src batch preview` use a Docker image published as `sourcegraph/src-batch-change-volume-workspace` for utility purposes when the volume workspace is selected, which is the default on macOS. This [Docker image](./docker/batch-change-volume-workspace/Dockerfile) is built by [a Python script](./docker/batch-change-volume-workspace/push.py) invoked by the GitHub Action workflow described in [`docker.yml`](.github/workflows/docker.yml).

To build and develop this locally, you can build and tag the image with:

```sh
docker build -t sourcegraph/src-batch-change-volume-workspace - < docker/batch-change-volume-workspace/Dockerfile
```

To remove it and then force the upstream image on Docker Hub to be used again:

```sh
docker rmi sourcegraph/src-batch-change-volume-workspace
```
