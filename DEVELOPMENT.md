## Development

If you want to develop the CLI, clone the repository and build/install it with `go install`

```
cd src-cli/
go install ./cmd/src
```

## Releasing

1.  If this is a non-patch release, update the changelog. Add a new section `## $MAJOR.MINOR` to [`CHANGELOG.md`](https://github.com/sourcegraph/src-cli/blob/main/CHANGELOG.md#unreleased) immediately under `## Unreleased changes`. Add new empty `Added`, `Changed`, `Fixed`, and `Removed` sections under `## Unreleased changes`.
2.  Find the latest version (either via the releases tab on GitHub or via git tags) to determine which version you are releasing.
3.  `VERSION=9.9.9 ./release.sh` (replace `9.9.9` with the version you are releasing)
4.  GitHub will automatically perform the release via the [goreleaser action](https://github.com/sourcegraph/src-cli/actions?query=workflow%3AGoreleaser). Once it has finished, **you need to confirm**:
    1. The [curl commands in the README](README.markdown#installation) fetch the latest version above.
    2. The [releases section of the repo sidebar](https://github.com/sourcegraph/src-cli) shows the correct version.
5.  Make the necessary updates to the main Sourcegraph repo:
    1. Update the `MinimumVersion` constant in the [src-cli package](https://github.com/sourcegraph/sourcegraph/tree/main/internal/src-cli/consts.go).
    2. Update the reference documentation by running `go generate ./doc/cli/references`.
    3. Commit the changes, and open a PR.

### Patch releases

If a backwards-compatible change is made _after_ a backwards-incompatible one, the backwards-compatible one should be re-released to older instances that support it.

A Sourcegraph instance returns the highest patch version with the same major and minor version as `MinimumVersion` as defined in the instance. Patch versions are reserved solely for non-breaking changes and minor bug fixes. This allows us to dynamically release fixes for older versions of `src-cli` without having to update the instance.

To release a bug fix or a new feature that is backwards compatible with one of the previous two minor version of Sourcegraph, cherry-pick the changes into a patch branch and re-releases with a new patch version. 

For example, suppose we have the the recommended versions.

| Sourcegraph version | Recommended src-cli version |
| ------------------- | --------------------------- | 
| `3.100`             | `3.90.5`                    |
| `3.99`              | `3.85.7`                    |

If a new feature is added to a new `3.91.6` release of src-cli and this change requires only features available in Sourcegraph `3.99`, then this feature should also be present in a new `3.85.8` release of src-cli. Because a Sourcegraph instance will automatically select the highest patch version, all non-breaking changes should increment only the patch version. 

Note that if instead the recommended src-cli version for Sourcegraph `3.99` was `3.90.4` in the example above, there is no additional step required, and the new patch version of src-cli will be available to both Sourcegraph versions.

## `src-cli` Docker image

Each release of `src` results in a new tag of the [`sourcegraph/src-cli` Docker image](https://hub.docker.com/r/sourcegraph/src-cli) being pushed to Docker Hub. This is handled by [goreleaser's Docker support](https://goreleaser.com/customization/docker/).

The main gotcha here is that the way goreleaser builds the Docker image is fairly difficult to replicate from the desktop: it builds a `src` binary without any runtime libc dependencies that can be installed in a `scratch` image, but unless you work on Alpine, your desktop is _not_ configured to build Go binaries like that.

As a result, there are two Dockerfiles in this project. goreleaser uses `Dockerfile.release`, which is replicated in a multi-stage `Dockerfile` that builds `src` in a builder container to ensure it's built in a way that can be tested.

### Testing

If you need to test a change to the Dockerfile (for example, due to a Renovate PR bumping the base image), you should pull that change, then build a local image with something like:

```bash
docker build -t local-src-cli .
```

After which you should be able to run:

```bash
docker run --rm local-src-cli
```

and get the normal help output from `src`.

## Dependent Docker images

`src batch apply` and `src batch preview` use a Docker image published as `sourcegraph/src-batch-change-volume-workspace` for utility purposes when the volume workspace is selected, which is the default on macOS. This [Docker image](./docker/batch-change-volume-workspace/Dockerfile) is built by [a Python script](./docker/batch-change-volume-workspace/push.py) invoked by the GitHub Action workflow described in [`docker.yml`](.github/workflows/docker.yml).

To build and develop this locally, you can build and tag the image with:

```sh
docker build -t sourcegraph/src-batch-change-volume-workspace - < docker/batch-change-volume-workspace/Dockerfile
```

To remove it and then force the upstream image on Docker Hub to be used again:

```sh
docker rmi sourcegraph/src-batch-change-volume-workspace
```
