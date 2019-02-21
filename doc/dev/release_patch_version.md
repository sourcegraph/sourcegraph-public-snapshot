# Release a patch version of Sourcegraph

To release a patch version of Sourcegraph, follow these instructions. Patch releases are released to fix bugs or security issues.

1. Merge your fix into master.

## Cut a new sourcegraph/server release
1. Get permission from the release captain to cherry pick your commit(s) onto the current release branch.
    - Ensure that it does not depend on any commits that aren't already in the release branch. If it does, those commit(s) it depends on will also need to be cherry-picked.
    - Push up the release branch with your cherry-picked commit(s) and make sure CI passes.
1. Create an annotated git tag using `git tag -a $VERSION -m $VERSION`, where `$VERSION` is the patch version number.
1. Then, `git push --tags`. This publishes the new tag, triggering the Docker image for the new version to start building. Once CI passes, the new image will be published to Docker Hub and will be available for use.

## Cut a new Sourcegraph cluster deployment release
1. Check out the `master` branch of the [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) repository.
1. Update the Docker image tags to the latest versions. See the [dev README in deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/README.dev.md#updating-docker-image-tags) for more info.
1. Tag the release. See the [dev README in `deploy-sourcegraph`](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/README.dev.md#tag-the-release) for more info.

## Update the docs
1. In sourcegraph/sourcegraph, update the docs to replace all instances of the old version number with the new version number. This includes updating [`latestReleaseDockerServerImageBuild`](https://sourcegraph.sgdev.org/github.com/sourcegraph/enterprise@093a16521df58e6c49cf70e6a4832137e740265a/-/blob/cmd/frontend/internal/app/pkg/updatecheck/handler.go#L26:2) and the [`latestReleaseKubernetesBuild`](https://sourcegraph.sgdev.org/github.com/sourcegraph/enterprise@093a16521df58e6c49cf70e6a4832137e740265a/-/blob/cmd/frontend/internal/app/pkg/updatecheck/handler.go#L31:2) variable. Updating these two variables require the new version tags for the Docker image and cluster deployment to be published.
1. Add a new Changelog entry stating what changed in your patch version.