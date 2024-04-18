# Add and Update Wolfi Base Images

When writing a Dockerfile, you typically base your image on an upstream release such as Alpine. Historically, we've used our [alpine-3.14](https://github.com/sourcegraph/sourcegraph/blob/main/docker-images/alpine-3.14/Dockerfile) base image for this purpose.

Wolfi base images are built _from scratch_ using [apko](https://github.com/chainguard-dev/apko/tree/main). This allows the image to be fully customised - for instance, an image doesn't need to include a shell or apk-tools.

## How base images are built

Base images are defined using an apko YAML configuration file, found under [wolfi-images](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/wolfi-images).

To make this YAML file declarative for Bazel, we generate a lockfile from the YAML similar to Go's `go.mod` or NPM's `package-lock.json`. This lockfile captures the exact versions and hashes of packages that will be included in the base image.

These lockfiles files are processed by the rules_apko Bazel ruleset, which outputs a base image.

You can see how this works locally:

- Edit `wolfi-images/gitserver.yaml` to add a new package
- Run `sg wolfi lock gitserver`
  - Inspect the gitserver lockfile at `wolfi-images/gitserver.lock.json` to see that it now includes your new package, pinned at a specific version
  - The lockfile may also include updated versions for other packages
- Run `sg wolfi image gitserver`
  - This uses Bazel to build the gitserver base image using the exact set of package versions defined in `gitserver.lock.json`.
  - This wraps `bazel run //<image>:base_tarball`, but will automatically run `sg wolfi lock` whenever the YAML is changed.

## How to...

### Update base image dependencies

Periodically, we should update the base images to ensure we include any updated packages and vulnerability fixes. To update the images:

- Review the [auto-update pull requests](https://github.com/sourcegraph/sourcegraph/pulls?q=is:pr+head:wolfi-auto-update/+is:open+) opened by Buildkite, and merge them.

Note: there are separate pull requests for `main` and the current release branch to avoid merge conflicts.

#### Automation

This process is automated by Buildkite, which runs a daily scheduled build to:

- Run `sg wolfi lock` to update the base image lockfiles, pulling in the latest versions of each package.
- Open, or update the previously-opened [pull request](https://github.com/sourcegraph/sourcegraph/pulls?q=is:pr+head:wolfi-auto-update/+is:open+).

The automatic PR is updated daily with the expectation it will be merged before each release, rather than merged daily.

To rerun the automation manually (perhaps to pick up a just-released package version or a change you made), open [Buildkite for sourcegraph/sourcegraph](https://buildkite.com/sourcegraph/sourcegraph) and choose New Build > Options > set Environment Variables to `WOLFI_BASE_REBUILD=true` and Create Build.

#### Manual image updates

If the automation fails and a manual update is needed, follow these steps:

- Run `sg wolfi lock` to pull in the latest package versions in all images
- Stage and commit changes using `git add wolfi-images/*.lock.json && git commit -m "Update base image lockfiles"`
- Create a PR and merge to the target branch.

### Modify an existing base image

To modify a base image, for example to add packages, users, or directories:

- Update the image's apko YAML configuration file, which can be found under [`wolfi-images/`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/)
- To build the **base image** (all depencencies listed in the YAML file, *without* Sourcegraph binaries) run `sg wolfi image <image-name>`
- To build the **full image** (all dependencies *and* Sourcegraph binaries):
  - Run `sg wolfi lock <image-name>`
  - Use Bazel to build the full image e.g. `bazel run //cmd/gitserver:image_tarball`
- Once happy, commit any changes in `wolfi-images/` and push. Your changes changes to the image will be reflected in the Buildkite pipeline, and in the production images once merged to `main`.

### Create a new base image

If your new image does not have any dependencies, use the [`sourcegraph`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/sourcegraph.yaml) base image.

Otherwise, you can create a new base image configuration file:

- Duplicate [`sourcegraph.yaml`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/sourcegraph.yaml) as a starting point.
- Add any required packages, users, directory structure, or metadata.
  - See [apko file format](https://github.com/chainguard-dev/apko/blob/main/docs/apko_file.md) for a full list of supported configuration.
  - See the other images under [`wolfi-images/`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/wolfi-images/) and [`chainguard-images/images`](https://github.com/chainguard-images/images/tree/main/images) for examples and best practices.
- Build the base image locally using `sg wolfi image <image-name>`
  - Test your changes by exec-ing into the image
- Review existing `BUILD.bazel` files like `cmd/gitserver/BUILD.bazel` for examples of how to declare an `oci_image()` build target that uses your new base image.
  - The key parts are to ensure the directory matches the YAML filename, calling `wolfi_base()`, and referencing `:wolfi_base_image`.
  - Test the build using the relevant Bazel targets e.g. `bazel run //cmd/gitserver:image_tarball`

This [pull request](https://github.com/sourcegraph/sourcegraph/pull/61881) provides a worked example of the above.
