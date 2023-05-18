# Wolfi Packages

## What are packages?

Packages are pre-built dependencies that we use in our container images.

The [Wolfi package repository](https://github.com/wolfi-dev/os) contains most common packages. If you need to add a new dependency to an image, you can search this repository by using `apk search`.

If we require a less common dependency such as `ctags` or `p4-fusion`, we can also build our own packages. These are stored in the [Sourcegraph package repository](#sourcegraph-package-repository). All third-party dependencies should be packaged, rather than fetching and building dependencies in Dockerfiles. This reduces build times, helps protect against supply-chain attacks, and prevents build failures caused by download timeouts or URL changes.

## How we build packages

Dependencies are packaged using . Package manifests are written as declarative YAML files, .

Dependencies are packaged using [Melange](https://github.com/chainguard-dev/melange), using a declarative YAML file. Melange follows build pipelines in these files, and runs in a sandbox to ensure isolation.

These YAML files are stored in [sourcegraph.git/wolfi-packages](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/wolfi-packages).

## How dependencies are packaged

Dependencies are typically packaged in one of two ways:
* Binary releases: download a precompiled binary of the dependency at a specific version, check its SHA checksum, and then move it to the final directory. See the [Comby package](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L27-37) for an example.
* Source releases: download the source code of the dependency at a specific version, check its SHA checksum, build the binary, then move it to the final directory. See the [syntect-server](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/syntect-server.yaml?L31-45) package for an example.

## Updating an existing packaged dependency

It's common to need to update a package to the most recent release in order to pull in new features or security patches.

1. Find the relevant package manifest YAML file in [sourcegraph.git/wolfi-packages](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/wolfi-packages).

2. Update the [`package.version`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L3) field to the latest version. This is usually substituted in a URL within the pipeline's [`fetch` step](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L30) as `${{package.version}}`. You will also need to update the [`expected-sha256`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L31), which can be found by downloading the release and running `sha256sum <file_name>`. 

  *  Several of our dependencies don't publish straightforward binary releases, so the pipeline may check out a Git repository on a specific branch or similar.

3. [Optional] Try building the package locally with the `local-build.sh` script. 

4. Push your branch and create a PR. Buildkite will build the new version of the package, which you can [test](#testing-packages). Once merged to `main`, it will be added to the Sourcegraph package repository.

## Creating a new package

Creating a new package should be an infrequent activity. Search the Wolfi package repository first, and if you're looking to build a common package then consider asking Chainguard to add it to the official repository. Feel free to reach out to #ask-security for help.

When creating a new package, the rough workflow is:

- Determine how the dependency will be fetched and built
  - If a binary release is available, this is often the simplest way
  - If only source releases are available, you'll need to download the source of a versioned release and build it. Projects typically include a Makefile, or building instructions in the README
- Add metadata such as the package name, version, and license
- [Test](#testing-packages) by building the package locally, and iterating
- Create a PR and merge to `main` to add it to the Sourcegraph package repository

## Packaging Tips

### Building packages

- The [wolfi-packages/](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/tree/wolfi-packages) directory and [Wolfi OS](https://github.com/wolfi-dev/os) repository are full of examples to base your package on.
- Read through the [Melange documentation](https://edu.chainguard.dev/open-source/melange/overview/)
- The [Melange documentation](https://edu.chainguard.dev/open-source/melange/melange-pipelines/) contains a list of available pipeline steps, which are common building blocks for building packages.
  - It can also be useful to refer to the [code that these pipelines run](https://github.com/chainguard-dev/melange/tree/main/pkg/build/pipelines)
- Spin up a dev Wolfi image with and run the build steps manually in there. This can be useful for debugging, or for speeding up iteration on slow-building dependencies
  - `docker run -it --entrypoint /bin/sh  us.gcr.io/sourcegraph-dev/wolfi-sourcegraph-dev-base:latest`

### Testing packages

- `.apk` files are just `.tar.gz` files, so can be extracted with `tar zxvf package.apk`. This is useful for checking their contents.

## Sourcegraph package repository

We maintain our own package repository for custom dependencies that aren't in the Wolfi repository.

This is implemented as a GCP bucket, and built using Buildkite.
