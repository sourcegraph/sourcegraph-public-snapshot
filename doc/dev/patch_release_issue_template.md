<!--
This template is used for patch releases.
It is not used for our monthly major/minor releases of Sourcegraph.
-->

# MAJOR.MINOR.PATCH Release (YYYY-MM-DD)

- [ ] Merge your fix into master.

## Cut a new sourcegraph/server release
- [ ] Get permission from the release captain to `git cherry-pick` your commit(s) onto the current release branch.
    - [ ] Ensure that it does not depend on any commits that aren't already in the release branch. If it does, those commit(s) it depends on will also need to be cherry-picked.
    - [ ] Push up the release branch with your cherry-picked commit(s) and make sure CI passes.
- [ ] Create an annotated git tag using `git tag -a $VERSION -m $VERSION`, where `$VERSION` is the patch version number.
- [ ] `git push --tags`. This publishes the new tag, triggering the Docker image for the new version to start building.
- [ ] Wait for the final Docker images to be available at https://hub.docker.com/r/sourcegraph/server/tags.

## Cut a new Sourcegraph cluster deployment release
- [ ] [Update the image tags](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/README.dev.md#updating-docker-image-tags) in [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/README.dev.md#updating-docker-image-tags) (e.g. [d0bb80](https://github.com/sourcegraph/deploy-sourcegraph/commit/d0bb80f559e7e9ef3b1915ddba72f4beff32276c))
- [ ] Wait for Renovate to pin the hashes in deploy-sourcegraph (e.g. [#190](https://github.com/sourcegraph/deploy-sourcegraph/pull/190/files)), or pin them yourself.
- [ ] Tag the release. See the [dev README in `deploy-sourcegraph`](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/README.dev.md#tag-the-release) for more info.

## Update the docs
- [ ] Update the documented version of Sourcegraph in `master` (e.g. `find . -type f -name '*.md' -exec sed -i '' -E 's/sourcegraph\/server:[0-9\.]+/sourcegraph\/server:$NEWVERSION/g' {} +`) (e.g. https://github.com/sourcegraph/sourcegraph/pull/2210/files <!-- TODO(nick): example that doesn't include latestReleaseDockerServerImageBuild -->).
- [ ] Update versions in docs.sourcegraph.com header ([example](https://github.com/sourcegraph/docs.sourcegraph.com/commit/d445c460c2da54fba4f56f647d656ca3311decf5))
- [ ] Update [`latestReleaseKubernetesBuild`](https://sourcegraph.sgdev.org/github.com/sourcegraph/enterprise@093a16521df58e6c49cf70e6a4832137e740265a/-/blob/cmd/frontend/internal/app/pkg/updatecheck/handler.go#L31:2) and [`latestReleaseDockerServerImageBuild`](https://sourcegraph.sgdev.org/github.com/sourcegraph/enterprise@093a16521df58e6c49cf70e6a4832137e740265a/-/blob/cmd/frontend/internal/app/pkg/updatecheck/handler.go#L26:2) in `master`.
- [ ] Create a new section for the patch version in the changelog. Verify that all changes that have been cherry picked onto the release branch have been moved to this section of the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/master/CHANGELOG.md) on `master`.
