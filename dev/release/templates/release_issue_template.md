<!--
DO NOTE COPY THIS ISSUE TEMPLATE MANUALLY. Use `pnpm run release tracking:issues` in the `sourcegraph/sourcegraph` repository.

Arguments:
- $MAJOR
- $MINOR
- $PATCH
- $RELEASE_DATE
- $ONE_WORKING_WEEK_BEFORE_RELEASE
- $THREE_WORKING_DAY_BEFORE_RELEASE
- $ONE_WORKING_DAY_AFTER_RELEASE
-->

# $MAJOR.$MINOR release

This release is scheduled for **$RELEASE_DATE**.

---

## Setup

- [ ] Ensure release configuration in [`dev/release/release-config.jsonc`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/release/release-config.jsonc) on `main` is up to date with the parameters for the current release.
- [ ] Ensure you have the latest version of the release tooling and configuration by checking out and updating `sourcegraph@main`.

## Security review (one week before release - $ONE_WORKING_WEEK_BEFORE_RELEASE)

- [ ] Create a [new issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose) using the **Security release approval** template and post a message in the [#security](https://sourcegraph.slack.com/archives/C1JH2BEHZ) channel tagging `@security-support`.

## Cut release (three days before release - $THREE_WORKING_DAY_BEFORE_RELEASE)

Perform these steps three days before the release date to generate a stable release candidate.

### Prepare release

- [ ] Post a release status update to Slack - [review all release-blocking issues](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Arelease-blocker), and ensure someone is resolving each.

  ```sh
  pnpm run release release:status
  ```

Do the [branch cut](./index.md#release-branches) for the release:

- [ ] Update the changelog and create pull requests:

  ```sh
  pnpm run release changelog:cut
  ```

- [ ] Manually review the pull requests created in the previous step and merge.
- [ ] Create the `$MAJOR.$MINOR` branch off the CHANGELOG commit in the previous step:

  ```sh
  pnpm run release release:branch-cut
  ```

- [ ] To support the multi-version upgrade utility, update [the `maxVersionString` constant](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24%40main+file:data/cmd/generator/consts.go+const+maxVersionString&patternType=lucky) to `$MAJOR.$MINOR.0` on the `main` branch, then cherry-pick this change into the `$MAJOR.$MINOR` branch. Bumping this version will require the `$MAJOR.$MINOR` branch to exist, and `go generate` will need to be invoked ([example](https://github.com/sourcegraph/sourcegraph/pull/43152)).

Upon branch cut, create and test release candidates:

- [ ] Tag the first release candidate:

  ```sh
  pnpm run release release:create-candidate
  ```

- [ ] Ensure that the following Buildkite pipelines all pass for the `v$MAJOR.$MINOR.$PATCH-rc.N` tag:
  - [ ] [Sourcegraph pipeline](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=v$MAJOR.$MINOR.$PATCH-rc.1)
- [ ] Cross check all reported CVEs are in the accepted list (`https://handbook.sourcegraph.com/departments/security/tooling/trivy/$MAJOR-$MINOR-$PATCH`). You can use the utility command `sg release cve-check` to help with this step. Otherwise, alert `@security-support` in the [#release-guild](https://sourcegraph.slack.com/archives/C032Z79NZQC) channel ASAP.
- [ ] File any failures and regressions in the pipelines as `release-blocker` issues and assign the appropriate teams.

Revert or disable features that may cause delays. As necessary, `git cherry-pick` bugfix (not feature!) commits from `main` into the release branch. Continue to create new release candidates as necessary, until no more `release-blocker` issues remain.

**Note**: You will need to re-check the above pipelines for any subsequent release candidates. You can see the Buildkite logs by tweaking the "branch" query parameter in the URLs to point to the desired release candidate. In general, the URL scheme looks like the following (replacing `N` in the URL):

- Sourcegraph: `https://buildkite.com/sourcegraph/sourcegraph/builds?branch=v$MAJOR.$MINOR.$PATCH-rc.N`

- [ ] Post a release status update to Slack:

  ```sh
  pnpm run release release:status
  ```

- [ ] Edit the following message to reflect the correct release candidate number, and post the message to the #cloud channel asking for the release candidate to be deployed to a test managed instance. You're good to go once the instance is up and running:

  ```
  Hey team, I'm the release captain for the $MAJOR.$MINOR release, posting here for asking for a release candidate (v$MAJOR.$MINOR.$PATCH-rc.N) to be deployed to a test managed instance. Could someone help here? :ty:
  ```

## Release day ($RELEASE_DATE)

### Stage release

<!-- Keep in sync with patch_release_issue's "Stage release" section -->

On the day of the release, confirm there are no more release-blocking issues (as reported by the `release:status` command), then proceed with creating the final release:

- [ ] Make sure [CHANGELOG entries](https://github.com/sourcegraph/sourcegraph/blob/main/CHANGELOG.md) have been moved from **Unreleased** to **$MAJOR.$MINOR.$PATCH**, but exluding the ones that merged to `main` after the branch cut (whose changes are not in the `$MAJOR.$MINOR` branch).
- [ ] Make sure [deploy-sourcegraph-helm CHANGELOG entries](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/CHANGELOG.md) have been moved from **Unreleased** to **$MAJOR.$MINOR.$PATCH**, but exluding the ones that merged to `main` after the branch cut (whose changes are not in the `$MAJOR.$MINOR` branch).
- [ ] Tag the final release:

  ```sh
  pnpm run release release:create-candidate final
  ```

- [ ] Ensure that the following pipelines all pass for the `v$MAJOR.$MINOR.$PATCH` tag:
  - [ ] [Sourcegraph pipeline](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=v$MAJOR.$MINOR.$PATCH)
- [ ] Wait for the `v$MAJOR.$MINOR.$PATCH` release Docker images to be available in [Docker Hub](https://hub.docker.com/r/sourcegraph/server/tags)
- [ ] Open PRs that publish the new release and address any action items required to finalize draft PRs (track PR status via the [generated release batch change](https://k8s.sgdev.org/organizations/sourcegraph/batch-changes)):

  ```sh
  pnpm run release release:stage
  ```

### Finalize release

- [ ] From the [release batch change](https://k8s.sgdev.org/organizations/sourcegraph/batch-changes), merge the release-publishing PRs created previously.
  - For [sourcegraph](https://github.com/sourcegraph/sourcegraph)
    - [ ] Cherry pick the release-publishing PR from `sourcegraph/sourcegraph@main` into the release branch.
  - For [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph)
    - [ ] Ensure the [release tag `v$MAJOR.$MINOR.$PATCH`](https://github.com/sourcegraph/deploy-sourcegraph/tags) has been created
  - For [deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker)
    - [ ] Ensure the [release tag `v$MAJOR.$MINOR.$PATCH`](https://github.com/sourcegraph/deploy-sourcegraph-docker/tags) has been created
  - For [deploy-sourcegraph-helm](https://github.com/sourcegraph/deploy-sourcegraph-helm)
    - [ ] Cherry pick the release-publishing PR from the release branch into main
- [ ] Alert the marketing team in [#release-post](https://sourcegraph.slack.com/archives/C022Y5VUSBU) that they can merge the release post.
- [ ] Finalize and announce that the release is live:

  ```sh
  pnpm run release release:announce
  ```

### Post-release

- [ ] Notify the next release captain that they are on duty for the next release.
- [ ] Open a PR to update [`dev/release/release-config.jsonc`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/release/release-config.jsonc) with the parameters for the next release.
  - [ ] Change `upcomingRelease` to the current patch release
  - [ ] Change `previousRelease` to the previous patch release version
  - [ ] Change `releaseDate` to the current date (time is optional) along with `oneWorkingDayAfterRelease` and `threeWorkingDaysBeforeRelease`
  - [ ] Change `captainSlackUsername` and `captainGitHubUsername` accordingly
- [ ] Ensure you have the latest version of the release tooling and configuration by checking out and updating `sourcegraph@main`.
- [ ] Create release calendar events, tracking issue, and announcement for next release:

  ```sh
  pnpm run release tracking:issues
  pnpm run release tracking:timeline
  ```

- [ ] Close the release.

  ```sh
  pnpm run release release:close
  ```

**Note:** If a patch release is requested after the release, ask that a [patch request issue](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=team%2Fdistribution&template=request_patch_release.md&title=$MAJOR.$MINOR.1%3A+) be filled out and approved first.
