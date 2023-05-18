# Wolfi

Sourcegraph is in the process of migrating from Alpine-based Docker images to Wolfi-based images. This section covers the migration, and highlight changes to our images build process.

## What is Wolfi?

[Wolfi](https://github.com/wolfi-dev) is a stripped-down open Linux distro designed for cloud-native containers. It has several key features that make it a good fit for distributing Sourcegraph:

* A distroless build system that lets us build containers with only the precise dependencies we need, reducing attack surface area and the number of components that need to be kept patched.
* A package repository that receives fast security patches and splits larger dependencies into smaller packages, keeping our images minimal and secure.
* High quality build-time SBOM (software bill of materials) support, allowing us to be transparent about our image composition.

## Why have we adopted Wolfi?

Adopting Wolfi helps significantly reduce the number of unpatched vulnerabilities in our container images, and move faster to patch them when new issues our found.

For full details, see [RFC 768: Harden container images by switching from Alpine to Wolfi](https://docs.google.com/document/d/1yQsXU7ekqPGjdkKItXKxROVNcJAYiGg3ZA70zJcLzIQ/edit#).

## How is the build process different from Alpine to Wolfi?

In our Alpine images, most of the work is done in the Dockerfile: packages, fetched, third-party dependencies are built, users and directories are created, and binaries are copied.

When building Wolfi images, most of this work is done outside of the Dockerfile:

- All third-party dependencies are [packaged](packages.md) as APKs (Alpine Packages - Wolfi shares this package format).
- Dependencies, user accounts, directory structure, and permissions are combined into a pre-built [base image](images.md).
- The Dockerfile then copies over Sourcegraph Go binaries and sets the entrypoint.

## What are base images?

Rather than using a customised upstream image like our [alpine-3.14](https://github.com/sourcegraph/sourcegraph/blob/main/docker-images/alpine-3.14/Dockerfile) base image, Wolfi base images are built from scratch using a [configuration file](https://github.com/sourcegraph/sourcegraph/tree/main/wolfi-images). This allows the image to be fully customised - for instance, an image doesn't need to include a shell or apk-tools. 

## More Information

- [Wolfi packages](packages.md)
- [Wolfi base images](images.md)
