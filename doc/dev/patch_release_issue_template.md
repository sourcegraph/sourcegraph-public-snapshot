<!--
This template is used for patch releases.
It is not used for our monthly major/minor releases of Sourcegraph.
See [release_issue_template.md](release_issue_template.md) for the monthly release checklist.
-->

# MAJOR.MINOR.PATCH Release

- [ ] Create a checklist of the changes that you want to release (i.e. open or merged PRs).
- [ ] Communicate your intentions by sending a message to #dev-announce that includes a link to this issue.
- [ ] Cherry pick changes into the release branch and check them off the list above.
    - [ ] Ensure that the cherry-picked commits don't depend on any commits that aren't already in the release branch.
- [ ] Push the release branch with your cherry-picked commit(s) and make sure CI passes.

## Release sourcegraph/server

- [ ] Create an annotated git tag using `git tag -a $VERSION -m $VERSION`, where `$VERSION` is the patch version number (e.g. `v3.1.1`).
- [ ] `git push origin $VERSION`. This publishes the new tag, triggering the Docker image for the new version to start building.
- [ ] Wait for the final Docker images to be available at https://hub.docker.com/r/sourcegraph/server/tags.

## Release Kubernetes deployment

In [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph):

- [ ] Wait for Renovate to open a PR to update the image tags and merge that PR ([example](https://github.com/sourcegraph/deploy-sourcegraph/pull/199)).
- [ ] Tag the `vMAJOR.MINOR.PATCH` release at this commit.

## Update the docs

- [ ] Update the documented version of Sourcegraph ([example](https://github.com/sourcegraph/sourcegraph/pull/2370/commits/701780fefa5809abb16669c9fb29738ec3bb2039)).
  ```
  find . -type f -name '*.md' -exec sed -i '' -E 's/sourcegraph\/server:[0-9\.]+/sourcegraph\/server:$NEWVERSION/g' {} +
  ```
- [ ] Update versions in docs.sourcegraph.com header ([example](https://github.com/sourcegraph/sourcegraph/pull/2701/commits/386e5ecb5225ab9c8ccc9791b489160ed7c984a2))
- [ ] Update `latestReleaseKubernetesBuild` and `latestReleaseDockerServerImageBuild` ([example](https://github.com/sourcegraph/sourcegraph/pull/2370/commits/15925f2769564225e37013acb52d9d0b30e1336c)).
- [ ] Create a new section for the patch version in the changelog. Verify that all changes that have been cherry picked onto the release branch have been moved to this section of the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/master/CHANGELOG.md) on `master`.
- [ ] Post a reply in the #dev-announce thread to say that the release is complete.
