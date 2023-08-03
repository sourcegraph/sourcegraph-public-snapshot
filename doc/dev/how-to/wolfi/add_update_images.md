# Add and Update Wolfi Base Images

When writing a Dockerfile, you typically base your image on an upstream release such as Alpine. Historically, we've used our [alpine-3.14](https://github.com/sourcegraph/sourcegraph/blob/main/docker-images/alpine-3.14/Dockerfile) base image for this purpose.

Wolfi base images are built _from scratch_ using [apko](https://github.com/chainguard-dev/apko/tree/main). This allows the image to be fully customised - for instance, an image doesn't need to include a shell or apk-tools.

## How base images are built

Base images are defined using an apko YAML configuration file, found under [wolfi-images](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/wolfi-images).

These configuration files can be processed with apko, which will generate a base image. You can build these locally using `sg wolfi image <image-name>`.

## How to...

### Update base image packages

In order to pull in updated packages with new features or fixed vulnerabilities, we need to periodically rebuild the base images.

This is currently a two-step process, but will be automated in the future:

- Run the [`wolfi-images/rebuild-images.sh`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@588463afbb0904c125cdcf78c7b182f43328504e/-/blob/wolfi-images/rebuild-images.sh) script (with an optional argument to just update one base image), commit the updated YAML files, and merge to main.
  - This will trigger Buildkite to rebuild the base images and publish them.
- Update the relevant Dockerfiles with the new base image's `sha256` hash, commit the change, and merge to main.
  - NOTE: Currently we use the `latest` label, but we will switch to using a `sha256` tag once deployed in production.

### Modify an existing base image

To modify a base image to add packages, users, or directories:

- Update its apko YAML configuration file, which can be found under [`wolfi-images/`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/)
- Build and testing it locally using `sg wolfi image <image-name>`.
  - You can use this local image in your Dockerfiles, or exec into it directly.
- Once happy with changes, create a PR and merge to main. Buildkite will detect the changes and rebuild the base image.
- Update the relevant Dockerfiles with the new base image's `sha256` hash, commit the change, and merge to main.
  - NOTE: Currently we use the `latest` label, but we will switch to using a `sha256` tag once deployed in production.

### Create a new base image

If your new image does not have any dependencies, use the [`sourcegraph`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/sourcegraph.yaml) base image.

Otherwise, you can create a new base image configuration file:

- Duplicate [`sourcegraph.yaml`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/sourcegraph.yaml) as a starting point.
- Add any required packages, users, directory structure, or metadata.
  - See [apko file format](https://github.com/chainguard-dev/apko/blob/main/docs/apko_file.md) for a full list of supported configuration.
  - See the other images under [`wolfi-images/`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/) and [`chainguard-images/images`](https://github.com/chainguard-images/images/tree/main/images) for examples and best practices.
- Build your image locally with `sg wolfi image <image-name>`.
- Commit your updated YAML file and merge it to main. Buildkite will build and publish your new image.

Once complete, treat the published image it as a standard base image, and use it in your Dockerfile.
