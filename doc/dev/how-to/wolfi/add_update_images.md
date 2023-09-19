# Add and Update Wolfi Base Images

When writing a Dockerfile, you typically base your image on an upstream release such as Alpine. Historically, we've used our [alpine-3.14](https://github.com/sourcegraph/sourcegraph/blob/main/docker-images/alpine-3.14/Dockerfile) base image for this purpose.

Wolfi base images are built _from scratch_ using [apko](https://github.com/chainguard-dev/apko/tree/main). This allows the image to be fully customised - for instance, an image doesn't need to include a shell or apk-tools.

## How base images are built

Base images are defined using an apko YAML configuration file, found under [wolfi-images](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/wolfi-images).

These configuration files can be processed with apko, which will generate a base image. You can build these locally using `sg wolfi image <image-name>`.

## How to...

### Update base images for a new release

Before each release, we should update the base images to ensure we include any updated packages and vulnerability fixes.

This is currently a two-step process, which will be further automated in the future:

- Run [`wolfi-images/rebuild-images.sh`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@588463afbb0904c125cdcf78c7b182f43328504e/-/blob/wolfi-images/rebuild-images.sh) script, commit the updated YAML files, and merge to main.
- Wait for the `main` branch's Buildkite run to complete.
  - Buildkite will rebuild the base images and publish them to Dockerhub.
- Run `sg wolfi update-hashes` locally to update the base image hashes in `dev/oci_deps.bzl`. Commit these changes and merge to `main`.
  - This fetches the updated base image hashes from the images that were pushed to Dockerhub in the previous step.
- Backport the PR that updated `dev/oci_deps.bzl` to the release branch.

### Modify an existing base image

To modify a base image to add packages, users, or directories:

- Update its apko YAML configuration file, which can be found under [`wolfi-images/`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/)
- Build the image
  - To build locally, use `sg wolfi image <image-name>`.
  - To build on CI, push your changes and Buildkite will build your image and push it to our `us.gcr.io/sourcegraph-dev/` dev repository. Instructions for pulling this image will be shown at the top of the Buildkite page.
- Test your changes by exec-ing into the image, or update `dev/oci_deps.bzl` to point at your dev base image and build the full image with Bazel.
- Once happy, merge your changes and wait for the `main` branch's Buildkite run to complete.
  - Buildkite will rebuild the base image and publish it to Dockerhub.
- Run `sg wolfi update-hashes <image-name>` to update the hashes for the changed image in `dev/oci_deps.bzl`. Commit and merge these changes.

### Create a new base image

If your new image does not have any dependencies, use the [`sourcegraph`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/sourcegraph.yaml) base image.

Otherwise, you can create a new base image configuration file:

- Duplicate [`sourcegraph.yaml`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/sourcegraph.yaml) as a starting point.
- Add any required packages, users, directory structure, or metadata.
  - See [apko file format](https://github.com/chainguard-dev/apko/blob/main/docs/apko_file.md) for a full list of supported configuration.
  - See the other images under [`wolfi-images/`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/) and [`chainguard-images/images`](https://github.com/chainguard-images/images/tree/main/images) for examples and best practices.
- Build the image:
  - To build locally, use `sg wolfi image <image-name>`.
  - To build on CI, push your changes and Buildkite will build your image and push it to our `us.gcr.io/sourcegraph-dev/` dev repository.
- Test your changes by exec-ing into the image, or update `dev/oci_deps.bzl` to point at your dev base image and build the full image with Bazel.
- Commit your updated YAML file and merge it to main. Buildkite will build and publish your new image to Dockerhub.

Once complete, treat the published image it as a standard Docker image, and add it to `dev/oci_deps.bzl`.
