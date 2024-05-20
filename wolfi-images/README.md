## Wolfi base images for Sourcegraph containers

Rather than building our containers on top of an upstream image like `alpine:latest`, at Sourcegraph we build our own containers entirely from scratch using Bazel and [apko](https://github.com/chainguard-dev/apko/tree/main).

This directory contains the configuration for each of our **base images**. Base images contain all the dependencies that the various components of Sourcegraph require in order to run, such as packages, users, groups, directores, and environment variables. For example, the [gitserver](./gitserver.yaml) configuration file ensures that Git is installed.

To create the final images that are shipped and deployed, we take the **base image** and use Bazel to build and add our own binaries on top.

The structure of this directory is:

- `<image>.yaml` - [apko](https://github.com/chainguard-dev/apko/tree/main) configuration that declares the set of packages, users & groups, directories, and envars for each base image
- `<image>.lock.json` - a lockfile which contains precise versions and hashes of packages, used by Bazel for reproducible builds. Generated from `<image>.yaml` using `sg wolfi lock`.

## Getting started

See the [Add and Update Wolfi Base Images](https://docs-legacy.sourcegraph.com/dev/how-to/wolfi/add_update_images) docs for guides to add new images and updating existing images. For more background, see the [Wolfi](https://docs-legacy.sourcegraph.com/dev/background-information/wolfi#wolfi) docs.

### Quickstart

- `sg wolfi lock gitserver` - update the `.lock.json` for gitserver with the latest set of package versions
- `sg wolfi image gitserver` - build the gitserver **base image**

## High-level Architecture

       file
      ┌──────────┐
      │          │
      │          │
      │   YAML   ├────────┐
      │          │        │          sg wolfi image <image>
      │          │        │                                            bazel target
      └─────┬────┘        │                   OR                      ┌─────────────────────┐
            │             │                                           │                     │
            │             │     bazel build //<image>/:base_image     │                     │
       sg wolfi lock      ├──────────────────────────────────────────►│     :base_image     │
       (manual step)      │                                           │                     │
            │             │                                           │                     │
       file │             │                                           └──────────┬──────────┘
      ┌─────▼─────┐       │                                                      │
      │           │       │                                                      │
      │           │       │                                                      │
      │ Lockfile  ├───────┘                                                      │
      │           │            ┌─────────────────────────────────────────────────┘
      │           │            │
      └───────────┘            │
                               │
                               │      bazel rule
                               │     ┌──────────────────────────────────────┐
                               │     │                                      │
                               │     │  oci_image(                          │
                               │     │                                      │
                               │     │    name = "image"                    │
                               │     │                                      │
                               └─────┼──► base = ":base_image"              │
        Bazel-genenarated            │                                      │
        binaries and      ───────────┼──► tars = ":tar_sourcegraph_binary"  │
        other resources              │                                      │
                                     │    [...]                             │
                                     │                                      │
                                     │  )                                   │
                                     │                                      │
                                     └──────────────────────────────────────┘
