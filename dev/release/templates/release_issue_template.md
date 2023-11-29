<!--
DO NOTE COPY THIS ISSUE TEMPLATE MANUALLY. Use `pnpm run release tracking:issues` in the `sourcegraph/sourcegraph` repository.

Arguments:
- $MAJOR
- $MINOR
- $PATCH
- $RELEASE_DATE
- $SECURITY_REVIEW_DATE
- $CODE_FREEZE_DATE
-->

# $MAJOR.$MINOR release

This release is scheduled for **$RELEASE_DATE**.

---

## Setup

<!-- Keep in sync with patch_release_issue_template's "Setup" section -->

- [ ] Ensure you have the latest version of the release tooling and configuration by checking out and updating `sourcegraph@main`.
- [ ] Ensure release configuration in [`dev/release/release-config.jsonc`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/release/release-config.jsonc) on `main` has version $MAJOR.$MINOR.$PATCH selected by using the command:

```shell
pnpm run release release:activate-release
```

## Security review ($SECURITY_REVIEW_DATE)

- [ ] Create a [Security release approval](https://github.com/sourcegraph/sourcegraph/issues/new/choose#:~:text=Security%20release%20approval) issue and post a message in the [#discuss-security](https://sourcegraph.slack.com/archives/C1JH2BEHZ) channel tagging `@security-support`.

## Cut release ($CODE_FREEZE_DATE)

Perform these steps three days before the release date to generate a stable release candidate.

### Prepare release

- [ ] Post a release status update to Slack - [review all release-blocking issues](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Arelease-blocker), and ensure someone is resolving each.

  ```sh
  pnpm run release release:status
  ```

Do the [branch cut](https://handbook.sourcegraph.com/departments/engineering/dev/process/releases/#release-branches) for the release:

- [ ] Update the changelog and create pull requests:

  ```sh
  pnpm run release changelog:cut
  ```

- [ ] Manually review the pull requests created in the previous step and merge.
- [ ] Wait for CI of the commit on `main` to pass.
- [ ] Request Admin permissions of `sourcegraph/sourcegraph` repository through [Entitle](https://app.entitle.io/request?targetType=resource&duration=1800&justification=Temporarily%20disable%20the%20%22Require%20linear%20history%22%20rule%20for%20release%20branches%20to%20create%20a%20new%20release%20branch.&integrationId=032680b6-f13d-42aa-9837-38097b45f0fe&resourceId=cd16ad0f-0e7e-4f20-8a8c-b3c57751dafd&roleId=5151f2f3-40a3-4697-99a2-b5e756e43f5b&grantMethodId=5151f2f3-40a3-4697-99a2-b5e756e43f5b) in order to disable the [**Require linear history** protection rule for release branches](https://github.com/sourcegraph/sourcegraph/settings/branch_protection_rules/34536616#:~:text=Require%20linear%20history).
- [ ] Enable the [`release-protector` GitHub Action](https://github.com/sourcegraph/sourcegraph/blob/main/.github/workflows/release-protector.yml) in the `sourcegraph/sourcegraph` repository.

- [ ] Create the `$MAJOR.$MINOR` branch off the CHANGELOG commit in the previous step:

  ```sh
  pnpm run release release:branch-cut
  ```

- [ ] Re-enable the [**Require linear history** protection rule for release branches](https://github.com/sourcegraph/sourcegraph/settings/branch_protection_rules/34536616#:~:text=Require%20linear%20history).

- [ ] Push a new release candidate tag. This command will automatically detect the appropriate release candidate number. This command can be executed as many times as required, and will increment the release candidate number for each subsequent build: :

  ```sh
  pnpm run release release:create-candidate
  ```

- [ ] Ensure that the following Buildkite pipelines all pass for the `v$MAJOR.$MINOR.$PATCH-rc.N` tag:
- [ ] [Sourcegraph pipeline](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=v$MAJOR.$MINOR.$PATCH-rc.1)
- [ ] Cross check all reported CVEs are in the accepted list (`https://handbook.sourcegraph.com/departments/security/tooling/trivy/$MAJOR-$MINOR-$PATCH`). You can use the utility command `sg release cve-check` to help with this step. Otherwise, alert `@security-support` in the [#release-guild](https://sourcegraph.slack.com/archives/C032Z79NZQC) channel ASAP.
- [ ] File any failures and regressions in the pipelines as `release-blocker` issues and assign the appropriate teams.

Revert or disable features that may cause delays. As necessary, `git cherry-pick` bugfix (not feature!) commits from `main` into the release branch. Continue to create new release candidates as necessary, until no more `release-blocker` issues remain.

- [ ] Update the [target branch of the RC test instance](https://github.com/sourcegraph/cloud/blob/main/.github/workflows/mi_upgrade_rctest.yml#L51) to the new release branch `$MAJOR.$MINOR`.
- [ ] Trigger a [manual run of the GitHub Action](https://github.com/sourcegraph/cloud/actions/workflows/mi_upgrade_rctest.yml) to upgrade the RC test instance. It should complete without an error, otherwise there might be a database migration problem that warrants a `release-blocker` issue.

> [!important]
> You will need to re-check the above pipelines for any subsequent release candidates. You can see the Buildkite logs by tweaking the "branch" query parameter in the URLs to point to the desired release candidate. In general, the URL scheme looks like the following (replacing `N` in the URL):

- Sourcegraph: `https://buildkite.com/sourcegraph/sourcegraph/builds?branch=v$MAJOR.$MINOR.$PATCH-rc.N`

- [ ] Post a release status update to Slack:

  ```sh
  pnpm run release release:status
  ```

## Code Freeze

Create candidates as necessary

```shell
pnpm run release release:create-candidate
```

Monitor the release branch, and backports. Ensure the branch remains healthy.

## Release day ($RELEASE_DATE)

### Stage release

<!-- Keep in sync with patch_release_issue's "Stage release" section -->

On the day of the release, confirm there are no more release-blocking issues (as reported by the `release:status` command), then proceed with creating the final release:

- [ ] Bake constants and other static values into the release branch (and also update main) This requires the release branch exists (should be automated above).
  ```shell
  pnpm run release release:bake-content
  ```
- [ ] Merge the resulting pull requests for the content bake generated by the command above
- [ ] Release a new version of src-cli, terraform-google-executors, aws-executors
  ```shell
  pnpm run release release:create-tags
  ```
- [ ] Ensure the latest version of src-cli is available in all sources. You may need to run this command a few times in the background.
  ```shell
    pnpm run release release:verify-releases
  ```
- [ ] Make another release candidate with the baked content
- [ ] Make sure [CHANGELOG entries](https://github.com/sourcegraph/sourcegraph/blob/main/CHANGELOG.md) have been moved from **Unreleased** to **$MAJOR.$MINOR.$PATCH**, but exluding the ones that merged to `main` after the branch cut (whose changes are not in the `$MAJOR.$MINOR` branch).
- [ ] Ensure security has approved the [Security release approval](https://github.com/sourcegraph/sourcegraph/issues?q=label%3Arelease-blocker+Security+approval+is%3Aopen) issue you created.
- [ ] Make sure [deploy-sourcegraph-helm CHANGELOG entries](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/CHANGELOG.md) have been moved from **Unreleased** to **$MAJOR.$MINOR.$PATCH**, but exluding the ones that merged to `main` after the branch cut (whose changes are not in the `$MAJOR.$MINOR` branch).
- [ ] Promote a release candidate to the final release build. You will need to provide the tag of the release candidate which you would like to promote as an argument. To get a list of available release candidates, you can use:

  ```shell
  pnpm run release release:check-candidate
  ```

  To promote the candidate, use the command:

  ```sh
  pnpm run release release:promote-candidate <tag>
  ```

- [ ] Ensure that the following pipelines all pass for the `v$MAJOR.$MINOR.$PATCH` tag:
  - [ ] [Sourcegraph pipeline](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=v$MAJOR.$MINOR.$PATCH)
- [ ] Wait for the `v$MAJOR.$MINOR.$PATCH` release Docker images to be available in [Docker Hub](https://hub.docker.com/r/sourcegraph/server/tags)
- [ ] Open PRs that publish the new release and address any action items required to finalize draft PRs (track PR status via the [generated release batch change](https://sourcegraph.sourcegraph.com/organizations/sourcegraph/batch-changes)):

  ```sh
  pnpm run release release:stage
  ```

### Finalize release

- [ ] From the [release batch change](https://sourcegraph.sourcegraph.com/organizations/sourcegraph/batch-changes), merge the release-publishing PRs created previously.
  - For [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph)
    - [ ] Ensure the [release tag `v$MAJOR.$MINOR.$PATCH`](https://github.com/sourcegraph/deploy-sourcegraph/tags) has been created
  - For [deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker)
    - [ ] Ensure the [release tag `v$MAJOR.$MINOR.$PATCH`](https://github.com/sourcegraph/deploy-sourcegraph-docker/tags) has been created
  - For [deploy-sourcegraph-helm](https://github.com/sourcegraph/deploy-sourcegraph-helm)
    - [ ] Cherry pick the release-publishing PR from the release branch into main
- [ ] Alert the marketing team in [#release-post](https://sourcegraph.slack.com/archives/C022Y5VUSBU) that they can merge the release post.
- [ ] Announce that the release is live:
  ```sh
  pnpm run release release:announce
  ```
- [ ] Disable the `release-protector` github action in sourcegraph/sourcegraph. This may require you to request admin permissions using Entitle.

### Post-release

- [ ] Create release calendar events, tracking issue, and announcement for next release (note: these commands will prompt for user input to generate the definition for the next release):
  ```sh
  pnpm run release release:prepare
  pnpm run release tracking:issues
  pnpm run release tracking:timeline
  ```
- [ ] Close the release.
  ```sh
  pnpm run release release:close
  ```
- [ ] Open a PR to update [`dev/release/release-config.jsonc`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/release/release-config.jsonc) with the auto-generated changes from above.

**Note:** If a patch release is requested after the release, ask that a [patch request issue](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=team%2Fdistribution&template=request_patch_release.md&title=$MAJOR.$MINOR.1%3A+) be filled out and approved first.
