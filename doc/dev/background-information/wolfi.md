# Wolfi

Sourcegraph is migrating from Alpine-based Docker images to Wolfi-based images. This page is written to answer questions about the migration and highlight changes to our images build process.

## Why do we need to leave Alpine?

## Why Wolfi?

## How does the build process change?

In our Alpine images, most of the work was done in the Dockerfile - it installed Alpine packages, fetched and built third-party dependencies, created users and directories, and copied over binaries.

When building Wolfi images, most of this work is done outside of the Dockerfile:

- All third-party dependencies are packaged as Alpine Packages (APK files).
- Alpine packages and packaged third-party dependencies, user accounts, the directory structure and permissions are combined into a pre-built base image.
- The Dockerfile then simply copies over our binaries and sets the entrypoint.

### Advantages of packaging third-party dependencies

By packaging third-party dependencies as APK files, any fetching and compilation work is done in advance. This reduces build times and avoids failures caused by download timeouts or URLs changing.

## Packages

We package third-party dependencies such as Comby and p4-fusion as APK packages using [Melange](https://github.com/chainguard-dev/melange). You write package configurations as declarative YAML files, which follow a simple build pipeline and run in a sandbox to ensure isolation.

These YAML files are stored in [sourcegraph.git/wolfi-packages](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/wolfi-packages).

### How do I update an existing packaged dependency?

Open its YAML file in [sourcegraph.git/wolfi-packages](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/wolfi-packages).

Most dependencies are packaged by following a pipeline which downloads a specific release, checks its SHA hash, extracts the archive, and moves the binary to the final directory path. See the [Comby package](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L27-37) for a typical example.

To update the dependency to a newer version, update the [`package.version`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L3) field to the latest version. This value is usually substituted into a URL within the pipeline's [`fetch` step](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L30) as `${{package.version}}`. You will also need to update the [`expected-sha256`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L31), which can be found by downloading the release and running `sha256sum <file_name>`.

Several of our dependencies don't publish straightforward binary releases, so the pipeline may check out a Git repository on a specific branch or similar.

Once you've updated the package's YAML file, you can build it locally by running  `wolfi-packages/local-build.sh <package_name>`.

TODO: How to test the package (extract it), how to publish it using CI on a branch + main

### How do I create a new package?

When creating a new package, the rough flow is:

- Supply metadata such as the package name, version, and license
- Determine how the dependency will be fetched and built
  - If a binary release is available this is often the simplest way
  - If not, checking out a Git repository and building is an option. Projects typically include a Makefile, or building instructions in the README
- Test by building the package locally and iterating
- Push to `main` to add it to the Sourcegraph package repository

The [Melange documentation](https://edu.chainguard.dev/open-source/melange/overview/) provides some information and a YAML reference. The [wolfi-packages/](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/tree/wolfi-packages) directory contains a number of examples, as does the [Wolfi OS](https://github.com/wolfi-dev/os) repository (these typically build from source, so are more complex).



Build tips:
- spin up a sourcegraph-dev wolfi image, and run the build steps manually in there
- The [Melange documentation](https://edu.chainguard.dev/open-source/melange/melange-pipelines/) contains a list of available pipeline steps, which are common pre-defined steps you might need to execute 
- The Melange repository contains the exact code these pipelines runs https://github.com/chainguard-dev/melange/tree/main/pkg/build/pipelines

Testing tips:
- test your package by extracting it with tar.gz
- test your package by installing it in a Wolfi image that has your local package directory mounted (need to figure this out)

### Why is there a Wolfi repository and a Sourcegraph repository?
