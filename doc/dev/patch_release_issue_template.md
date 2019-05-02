<!--
This template is used for patch releases.
It is not used for our monthly major/minor releases of Sourcegraph.
See [release_issue_template.md](release_issue_template.md) for the monthly release checklist.
-->

# MAJOR.MINOR.PATCH Release

- [ ] Create a checklist of the changes that you want to release (i.e. open or merged PRs).
- [ ] Communicate your intentions by sending a message to #dev-announce that includes a link to this issue.
- [ ] Cherry pick changes into the existing release branch (this exists already as `MAJOR.MINOR`, do not create a new branch) and check them off the list above.
    - [ ] Ensure that the cherry-picked commits don't depend on any commits that aren't already in the release branch.
- [ ] Push the release branch with your cherry-picked commit(s) and make sure CI passes.

## Release sourcegraph/server

- [ ] Create an annotated git tag and push it (this triggers CI to build the Docker images for the new version). For example:
    ```
    VERSION='v3.2.1-rc.1' bash -c 'git tag -a "$VERSION" -m "$VERSION" && git push origin "$VERSION"'
    ```

- [ ] Wait for the final Docker images to be available at https://hub.docker.com/r/sourcegraph/server/tags.

## Release Kubernetes deployment

In [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph):

- [ ] Wait for Renovate to open a PR to update the image tags and merge that PR ([example](https://github.com/sourcegraph/deploy-sourcegraph/pull/199)).
- [ ] Follow the ["Cutting a new patch version" instructions in https://github.com/sourcegraph/deploy-sourcegraph/blob/master/README.dev.md](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/README.dev.md#cutting-a-new-patch-version-eg-v2121)

## Update the docs

- [ ] Update the version (major.minor.patch) of Sourcegraph in the docs ([example](https://github.com/sourcegraph/sourcegraph/pull/2841)) by running the following
  ```
  find . -type f -name '*.md' -exec sed -i '' -E 's/sourcegraph\/server:[0-9\.]+/sourcegraph\/server:$NEW_VERSION/g' {} +
  ```
- [ ] [Update deploy-aws version](https://github.com/sourcegraph/deploy-sourcegraph-aws/blob/master/ec2/resources/user-data.sh#L3)
- [ ] [Update deploy-digitalocean version ](https://github.com/sourcegraph/deploy-sourcegraph-digitalocean/blob/master/resources/user-data.sh#L3)
- [ ] Update versions in docs.sourcegraph.com template ([example](https://github.com/sourcegraph/sourcegraph/pull/2841/files#diff-3d0e70da24a04f44a1fdc404b7242b89))
- [ ] Update `latestReleaseKubernetesBuild` and `latestReleaseDockerServerImageBuild` ([example](https://github.com/sourcegraph/sourcegraph/pull/2370/commits/15925f2769564225e37013acb52d9d0b30e1336c)).
- [ ] [Update deploy-aws version](https://github.com/sourcegraph/deploy-sourcegraph-aws/edit/master/ec2/resources/user-data.sh#L3)
- [ ] [Update deploy-digitalocean version ](https://github.com/sourcegraph/deploy-sourcegraph-digitalocean/edit/master/resources/user-data.sh#L3)
- [ ] Message @slimsag on Slack: `MAJOR.MINOR.PATCH has been released, update deploy-sourcegraph-docker as needed`
- [ ] Create a new section for the patch version in the changelog. Verify that all changes that have been cherry picked onto the release branch have been moved to this section of the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/master/CHANGELOG.md) on `master`.
- [ ] Post a reply in the #dev-announce thread to say that the release is complete.
