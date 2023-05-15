# Wolfi

Sourcegraph is in the process of migrating from Alpine-based Docker images to Wolfi-based images. This section covers the migration, and highlight changes to our images build process.

## What is Wolfi?

[Wolfi](https://github.com/wolfi-dev) is a stripped-down open Linux distro designed for cloud-native containers. It has several key features that make it a good fit for distributing Sourcegraph:

* A distroless build system that lets us build containers with only the precise dependencies we need, reducing attack surface area and the number of components that need to be kept patched.
* A package repository that receives fast security patches and splits larger dependencies into smaller packages, keeping our images minimal and secure.
* High quality build-time SBOM (software bill of materials) support, allowing us to be transparent about our image composition.

## Why are we adopting Wolfi?

Adopting Wolfi will help significantly reduce the number of unpatched vulnerabilities in our container images, and move faster to patch them when new issues our found.

For full details, see [RFC 768: Harden container images by switching from Alpine to Wolfi](https://docs.google.com/document/d/1yQsXU7ekqPGjdkKItXKxROVNcJAYiGg3ZA70zJcLzIQ/edit#).

## How is the build process different from Alpine to Wolfi?

In our Alpine images, most of the work is done in the Dockerfile: packages, fetched, third-party dependencies are built, users and directories are created, and binaries are copied.

When building Wolfi images, most of this work is done outside of the Dockerfile:

- All third-party dependencies are packaged as APKs (Alpine Packages - Wolfi shares this package format).
- Dependencies, user accounts, directory structure, and permissions are combined into a pre-built **base image**.
- The Dockerfile then copies over Sourcegraph Go binaries and sets the entrypoint.

## Base Images

Rather than using a customised upstream image like our [alpine-3.14](https://github.com/sourcegraph/sourcegraph/blob/main/docker-images/alpine-3.14/Dockerfile) base image, Wolfi base images are built from scratch using a [configuration file](https://github.com/sourcegraph/sourcegraph/tree/main/wolfi-images). This allows the image to be fully customised - for instance, an image doesn't need to include a shell or apk-tools. 


# Packages

The [Wolfi package repository](https://github.com/wolfi-dev/os) contains most common packages. If you need to add a new dependency, you can check this repository or use `apk search`.

If we require a less common dependency such as ctags or p4-fusion, we can build our own packages. All third-party dependencies should be packaged, rather than fetching and building dependencies in Dockerfiles. This reduces build times, helps protect against supply-chain attacks, and prevents build failures caused by download timeouts or URL changes.

Dependencies are packaged using [Melange](https://github.com/chainguard-dev/melange). Package manifests are written as declarative YAML files, which follow a simple build pipeline and run in a sandbox to ensure isolation.

These YAML files are stored in [sourcegraph.git/wolfi-packages](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/wolfi-packages).

## How dependencies are packaged

Dependencies are typically packaged in one of two ways:
* Binary releases download a precompiled binary of the dependency at a specific version, check its SHA checksum, and then move it to the final directory. See the [Comby package](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L27-37) for an example.
* Source releases download the source code of the dependency at a specific version, check its SHA checksum, build the binary, then move it to the final directory. See the [syntect-server](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/syntect-server.yaml?L31-45) package for an example.

## Updating an existing packaged dependency

It's common to need to update a package to the most recent release in order to pull in new features or security patches.

1. Find the relevant package manifest YAML file in [sourcegraph.git/wolfi-packages](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/wolfi-packages).

2. Update the [`package.version`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L3) field to the latest version. This is usually substituted in a URL within the pipeline's [`fetch` step](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L30) as `${{package.version}}`. You will also need to update the [`expected-sha256`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@321e0e9d01fa23b83bef57c1e69076866094af20/-/blob/wolfi-packages/comby.yaml?L31), which can be found by downloading the release and running `sha256sum <file_name>`. 

  *  Several of our dependencies don't publish straightforward binary releases, so the pipeline may check out a Git repository on a specific branch or similar.

3. Push your branch and create a PR. Buildkite will build the new version of the package, which you can [test](#testing-packages) before merging to main.



## Creating a new package

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

## Testing packages

Testing tips:
- test your package by extracting it with tar.gz
- test your package by installing it in a Wolfi image that has your local package directory mounted (need to figure this out)

### Why is there a Wolfi repository and a Sourcegraph repository?

# Wolfi images
