<!--
This template is used for patch releases.
It is not used for our monthly major/minor releases of Sourcegraph.
-->

# MAJOR.MINOR.PATCH Release

- [ ] Merge your fix into master.

## Cut a new sourcegraph/server release
- [ ] Get permission from the release captain to `git cherry-pick` your commit(s) onto the current release branch.
    - [ ] Ensure that it does not depend on any commits that aren't already in the release branch. If it does, those commit(s) it depends on will also need to be cherry-picked.
    - [ ] Push up the release branch with your cherry-picked commit(s) and make sure CI passes.
- [ ] Create an annotated git tag using `git tag -a $VERSION -m $VERSION`, where `$VERSION` is the patch version number.
- [ ] `git push --tags`. This publishes the new tag, triggering the Docker image for the new version to start building.
- [ ] Wait for the final Docker images to be available at https://hub.docker.com/r/sourcegraph/server/tags.

## Cut a new Sourcegraph cluster deployment release
- [ ] In [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph):
    - [ ] Wait for Renovate to open a PR to update the image tags and merge that PR ([example](https://github.com/sourcegraph/deploy-sourcegraph/pull/199)).
    - [ ] Tag the `vMAJOR.MINOR.PATCH` release at this commit.

## Update the docs
- [ ] Update the documented version of Sourcegraph ([example](https://github.com/sourcegraph/sourcegraph/pull/2370/commits/701780fefa5809abb16669c9fb29738ec3bb2039)).
  ```
  find . -type f -name '*.md' -exec sed -i '' -E 's/sourcegraph\/server:[0-9\.]+/sourcegraph\/server:$NEWVERSION/g' {} +
  ```
- [ ] Update versions in docs.sourcegraph.com header ([example](https://github.com/sourcegraph/docs.sourcegraph.com/commit/d445c460c2da54fba4f56f647d656ca3311decf5))
- [ ] Update `latestReleaseKubernetesBuild` and `latestReleaseDockerServerImageBuild` ([example](https://github.com/sourcegraph/sourcegraph/pull/2370/commits/15925f2769564225e37013acb52d9d0b30e1336c)).
- [ ] Create a new section for the patch version in the changelog. Verify that all changes that have been cherry picked onto the release branch have been moved to this section of the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/master/CHANGELOG.md) on `master`.
