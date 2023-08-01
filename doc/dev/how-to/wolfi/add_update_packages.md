# Add and Update Wolfi Packages

## What are packages, and why do we use them?

Linux packages are bundles of software and related files that are designed to be easily installed and managed.

As well as using common packages from the Wolfi repository, we package all our third-party dependencies. This makes it easier to add and modify dependencies, reduces build times, and increases security. This page focuses on how we package these third-party dependencies.

For full details on why we use packages, see [RFC 769: Package container build dependencies as Alpine packages](https://docs.google.com/document/d/1VFxBECDErU5bR5uPDsiYREC_qfHDEDyOSaDTaH83nZU/edit#).

## Finding and building packages

Sourcegraph's container images use Wolfi, and the [Wolfi package repository](https://github.com/wolfi-dev/os) contains many common packages. If you need to add a new dependency to an image, you can search this repository by using `apk search`.

If we require a less common dependency such as `ctags` or `p4-fusion`, we can also build our own packages. All third-party dependencies should be packaged, rather than fetching and building dependencies in Dockerfiles.

Dependencies are packaged using [Melange](https://github.com/chainguard-dev/melange), using a declarative YAML file. Melange follows a sequence of build instructions (known as "pipelines"), and runs in a sandbox to ensure isolation.

All Sourcegraph package configs are stored in [sourcegraph.git/wolfi-packages](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/wolfi-packages).

## How dependencies are packaged

Dependencies are typically packaged in one of two ways:

- Binary releases: download a precompiled binary of the dependency at a specific version, check its SHA checksum, and then move it to the final directory path. See the [p4cli package](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@760db946dd9c3b23af69f2036b7a8c11e38307b4/-/blob/wolfi-packages/p4cli.yaml?L20-29) for an example.
- Source releases: download the source code of the dependency at a specific version, check its SHA checksum, build the binary, then move it to the final directory. See the [syntect-server](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/syntect-server.yaml?L29-45) package for an example.

## How to...

### Update an existing packaged dependency

It's common to need to update a package to the most recent release in order to pull in new features or security patches.

1. Find the relevant package manifest YAML file in [sourcegraph.git/wolfi-packages](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/wolfi-packages).

2. Update the [`package.version`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L3) field to the latest version. This is usually substituted in a URL within the pipeline's [`fetch` step](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L30) as `${{package.version}}`. You will also need to update the [`expected-sha256`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L31), which can be found by downloading the release and running `sha256sum <file_name>`.

- Depending on the package, this step may download a binary or source code. Projects release code in different ways, so the pipeline may check out a Git repository on a specific branch or download a `.tar.gz` file containing source code.

3. Optionally, build the package locally with `sg wolfi package <package-name>`. This will create an `.apk` file under `wolfi-packages/local-repo/packages/x86_64` which you can [inspect](#testing-packages).

4. Push your branch and create a PR. Buildkite will build the new version of the package, which you can [test](#testing-packages). Once merged to `main`, it will be added to the [Sourcegraph package repository](#sourcegraph-package-repository).

### Create a new package

Creating a new package should be an infrequent activity. Search the Wolfi package repository first, and if you're looking to build a common package then consider asking Chainguard to add it to the Wolfi repository. Feel free to reach out to #ask-security for assistance.

When creating a new package, the rough workflow is:

- Determine how the dependency will be fetched and built.
  - If a binary release is available, this is often the simplest way.
  - If only source releases are available, you'll need to download the source of a versioned release and build it.
  - Projects typically include a Makefile, or building instructions in their README or INSTALL.
- Add metadata such as the package name, version, and license.
- Iterate by [building](#building-packages) the package locally using `sg wolfi package <package-name>`.
- [Test your new package](#testing-packages)
- Once confident the package works as expected, create a PR and merge to `main` to add it to the Sourcegraph package repository.

## Packaging Tips

### Building packages

- The [wolfi-packages/](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/tree/wolfi-packages) directory and [Wolfi OS](https://github.com/wolfi-dev/os) repository are full of examples to base your package on.
- Read through the [Melange documentation](https://edu.chainguard.dev/open-source/melange/overview/).
- The [Melange documentation](https://edu.chainguard.dev/open-source/melange/melange-pipelines/) contains a list of available pipeline steps, which are common building blocks for building packages.
  - It can also be useful to refer to the [code that these pipelines run](https://github.com/chainguard-dev/melange/tree/main/pkg/build/pipelines).
- Spin up a dev Wolfi image with and run the build steps manually in there. This is useful for debugging, or for speeding up iteration on slow-building dependencies.
  - `docker run -it --entrypoint /bin/sh  us.gcr.io/sourcegraph-dev/wolfi-sourcegraph-dev-base:latest`
- You can build packages locally using `sg wolfi package <package-name>`.

### Testing packages

- `.apk` files are just `.tar.gz` files, so can be extracted with `tar zxvf package.apk`. This is useful for checking their contents.
  - After building locally with `sg wolfi package <package-name>`, packages can be found under `wolfi-packages/local-repo/packages/x86_64/`.
- Always try installing the package in a container, as this ensures that all runtime dependencies can be satisfied.
- After building a package locally with `sg wolfi package`, you can test it out in a specific base image by modifying that image's manifest (under `wolfi-images/`) from `package@sourcegraph` to `package@local` and run `sg wolfi image <image-name>`. This will build the base image using the package from your local repository.

## Sourcegraph package repository

We maintain our own package repository for custom dependencies that aren't in the Wolfi repository.

This is implemented as a GCP bucket. Buildkite is used to build and upload packages, as well as to [build and sign](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/dev/ci/scripts/wolfi/build-repo-index.sh) the repository index.

Currently, all packages are uploaded directly to the main Sourcegraph package repo. To provide an isolated development environment, in the future we plan to separate our dev and production repos:

- The `main/` directory only contains packages built from the main branch.
- Each branch which updates package manifests will have its own `branch-name` repository.

### Local package repository

After building a package locally with `sg wolfi package <package-name>`, it is stored in a local package repository under `wolfi-packages/local-repo/packages/`. Thes packages can be used in locally-built base images by referencing them using `package@local` in the image manifest and running `sg wolfi image <image-name>`.
