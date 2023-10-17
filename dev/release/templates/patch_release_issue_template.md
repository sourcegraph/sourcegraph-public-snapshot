<!--
DO NOT COPY THIS ISSUE TEMPLATE MANUALLY. Use `pnpm run release tracking:issues` in the `sourcegraph/sourcegraph` repository.

Arguments:
- $MAJOR
- $MINOR
- $PATCH
- $RELEASE_DATE
- $ONE_WORKING_DAY_AFTER_RELEASE
-->

# $MAJOR.$MINOR.$PATCH patch release

This release is scheduled for **$RELEASE_DATE**.

> [!WARNING]
> To get your commits in `main` included in this patch release, add the `backport-$MAJOR.$MINOR` to the PR to `main`.

## Setup

<!-- Keep in sync with release_issue_template's "Setup" section -->

- [ ] Ensure you have the latest version of the release tooling and configuration by checking out and updating `sourcegraph@main`.
- [ ] Ensure release configuration in [`dev/release/release-config.jsonc`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/release/release-config.jsonc) on `main` has version $MAJOR.$MINOR.$PATCH selected by using the command:

```shell
pnpm run release release:activate-release
```

## Prepare release

- [ ] Ensure that all [backported PRs](https://github.com/sourcegraph/sourcegraph/pulls?q=is%3Apr+is%3Aopen+base%3A$MAJOR.MINOR) have been merged.

Create and test the first release candidate:

> [!NOTE]
> Ensure that you've pulled both main and release branches before running this command.

- [ ] Push a new release candidate tag. This command will automatically detect the appropriate release candidate number. This command can be executed as many times as required, and will increment the release candidate number for each subsequent build: :

  ```sh
  pnpm run release release:create-candidate
  ```

- [ ] Ensure that the following Buildkite pipelines all pass for the `v$MAJOR.$MINOR.$PATCH-rc.1` tag:
  - [ ] [Sourcegraph pipeline](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=v$MAJOR.$MINOR.$PATCH-rc.1)
- [ ] File any failures and regressions in the pipelines as `release-blocker` issues and assign the appropriate teams.

> [!NOTE]
> You will need to re-check the above pipelines for any subsequent release candidates. You can see the Buildkite logs by tweaking the "branch" query parameter in the URLs to point to the desired release candidate. In general, the URL scheme looks like the following (replacing `N` in the URL): `https://buildkite.com/sourcegraph/sourcegraph/builds?branch=v$MAJOR.$MINOR.$PATCH-rc.N`

Once there is a release candidate available:

- [ ] Create a [Security release approval](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=andreeleuterio%2C+evict%2C+willdollman%2C+mohammadualam&labels=release-blocker&projects=&template=security-release-approval.md&title=$MAJOR.$MINOR.$PATCH+Security+approval) issue and post a message in the [#discuss-security](https://sourcegraph.slack.com/archives/C1JH2BEHZ) channel tagging `@security-support`.

## Stage release

<!-- Keep in sync with release_issue_template's "Stage release" section -->

- [ ] Verify the **$MAJOR.$MINOR.$PATCH** section of [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/main/CHANGELOG.md) on the `main` is accurate.
- [ ] Ensure security has approved the [Security release approval](https://github.com/sourcegraph/sourcegraph/issues?q=label%3Arelease-blocker+Security+approval+is%3Aopen) issue you created.
- [ ] Promote a release candidate to the final release build. You will need to provide the tag of the release candidate which you would like to promote as an argument. To get a list of available release candidates, you can use:
  ```shell
  pnpm run release release:check-candidate
  ```
  To promote the candidate, use the command:
  ```sh
  pnpm run release release:promote-candidate <tag>
  ```
- [ ] Ensure that the pipeline for the `v$MAJOR.$MINOR.$PATCH` tag has passed: [Sourcegraph pipeline](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=v$MAJOR.$MINOR.$PATCH)
- [ ] Wait for the `$MAJOR.$MINOR.$PATCH` release Docker images to be available in [Docker Hub](https://hub.docker.com/r/sourcegraph/server/tags)
- [ ] Open PRs that publish the new release and address any action items required to finalize draft PRs (track PR status via the [generated release batch change](https://sourcegraph.sourcegraph.com/organizations/sourcegraph/batch-changes/release-sourcegraph-$MAJOR.$MINOR.$PATCH)):
  ```sh
  pnpm run release release:stage
  ```

## Finalize release

<!-- Keep in sync with release_issue_template's "Finalize release" section, except no blog post -->

- [ ] From the [release batch change](https://sourcegraph.sourcegraph.com/organizations/sourcegraph/batch-changes/release-sourcegraph-$MAJOR.$MINOR.$PATCH), merge the release-publishing PRs created previously. Note: some PRs require certain actions performed before merging.
- [ ] **After all the PRs are merged**, perform following checks/actions
  - For [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph)
    - [ ] Ensure the [release tag](https://github.com/sourcegraph/deploy-sourcegraph/tags) has been created
  - For [deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker)
    - [ ] Ensure the [release tag](https://github.com/sourcegraph/deploy-sourcegraph-docker/tags) has been created
  - For [deploy-sourcegraph-helm](https://github.com/sourcegraph/deploy-sourcegraph-helm), also:
    - [ ] Update the [changelog](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/CHANGELOG.md) to include changes from the patch
    - [ ] Cherry-pick the release-publishing PR from the release branch into `main`
- [ ] Announce that the release is live:
  ```sh
  pnpm run release release:announce
  ```

## Post-release

- [ ] Close the release:

```shell
  pnpm run release release:close
```

- [ ] Open a PR to update [`dev/release/release-config.jsonc`](https://github.com/sourcegraph/sourcegraph/edit/main/dev/release/release-config.jsonc) after the auto-generated changes above if any.
- [ ] Update the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/main/CHANGELOG.md) by opening and merging a PR into `main` (**not** the release branch), making the following changes:
  - [ ] Move the released changes into the $MAJOR.$MINOR.$PATCH version section
  - [ ] Add a new section for the [upcoming patch release](https://handbook.sourcegraph.com/departments/engineering/dev/process/releases/#current-patch-schedule) if this is not the last planned patch release for version $MAJOR.$MINOR

> [!NOTE]
> If another patch release is requested after the release, ask that a [patch request issue](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=team%2Fdistribution&template=request_patch_release.md) be filled out and approved first.
