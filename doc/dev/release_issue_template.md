<!--
This template is used for our monthly major/minor releases of Sourcegraph.
It is not used for patch releases. See [patch_release_issue_template.md](patch_release_issue_template.md)
for the patch release checklist.

Run a find replace on:
- $VERSION
- $RELEASE_DATE
- $FIVE_WORKING_DAYS_BEFORE_RELEASE
- $FOUR_WORKING_DAYS_BEFORE_RELEASE
- $THREE_WORKING_DAYS_BEFORE_RELEASE
- $ONE_WORKING_DAY_BEFORE_RELEASE
-->

# $VERSION Release ($RELEASE_DATE)

**Note:** All `yarn run release ...` commands should be run from `dev/release`.

## At the start of the iteration

- [ ] Create the retrospective doc for the next iteration by copying the previous one.
- [ ] Update `dev/release/config.json` with the parameters for the current release.
- [ ] Add calendar events and reminders for key dates in the release cycle: `yarn run release add-timeline-to-calendar`
- [ ] Create the release tracking issue (i.e., this issue): `yarn run release tracking-issue:create`
- [ ] Post link to tracking to #dev-announce: `yarn run release tracking-issue:announce`
- [ ] Create a new test grid for MAJOR.MINOR by cloning the previous [release testing grid on Monday.com](https://sourcegraph-team.monday.com) and renaming it to "MAJOR.MINOR Release test grid".
    - [ ] Reset all tested cells to "To test", unless the "Automated" column is marked as "Done". See [this article for how to update multiple values in Monday.com](https://support.monday.com/hc/en-us/articles/115005335049-Batch-Actions-Edit-multiple-items-in-one-click).
    - [ ] Assign rows in the release testing grid to engineers from the team that owns the row.

## 5 working days before release ($FIVE_WORKING_DAYS_BEFORE_RELEASE)

- [ ] Use `./dev/release-ping.sh` to message teammates who have open issues / PRs in the milestone and ask them to remove those that won't be done by the time that the release branch is scheduled to be created.
- [ ] Verify that there is a draft of the blog post and that it will be ready to be merged on time.
- [ ] Ping each team, and ask them to identify which of the optional rows that they own on the release testing grid should be tested this iteration.
- [ ] Ping the @distribution team to determine which environments each row on the release testing grid should be tested in.

## 4 working days before release ($FOUR_WORKING_DAYS_BEFORE_RELEASE)

- [ ] **HH:MM AM/PM PT** Add a new section header for this version to the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/master/CHANGELOG.md#unreleased) immediately under the `## Unreleased changes` heading and add new empty sections under `## Unreleased changes` ([example](https://github.com/sourcegraph/sourcegraph/pull/2323)).
- [ ] Create the `MAJOR.MINOR` branch for this release off of the changelog commit that you created in the previous step.
- [ ] Tag the first release candidate `vMAJOR.MINOR.0-rc.1`:
    ```
    VERSION='vMAJOR.MINOR.0-rc.1' bash -c 'git tag -a "$VERSION" -m "$VERSION" && git push origin "$VERSION"'
    ```
- [ ] Send a message to #dev-announce to announce the release candidate.
- [ ] Run Sourcegraph Docker image with no previous data.
    - [ ] Run the new version of Sourcegraph.
        ```
        CLEAN=true IMAGE=sourcegraph/server:MAJOR.MINOR.0-rc.1 ./dev/run-server-image.sh
        ```
    - [ ] Initialize the site by creating an admin account.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that basic code intelligence works on Go or TypeScript.
- [ ] Upgrade Sourcegraph Docker image from previous released version.
    - [ ] Run the previous version of Sourcegraph.
        ```
        CLEAN=true IMAGE=sourcegraph/server:$OLDVERSION ./dev/run-server-image.sh
        ```
    - [ ] Initialize the site by creating an admin account.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
    - [ ] Stop the previous version of Sourcegraph and run the new version of Sourcegraph with the same data.
        ```
        CLEAN=false IMAGE=sourcegraph/server:MAJOR.MINOR.0-rc.1 ./dev/run-server-image.sh
        ```
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that basic code intelligence works on Go or TypeScript.
- [ ] Run the new version of Sourcegraph on a clean Kubernetes cluster with no previous data.
    - [ ] Create a new Kubernetes cluster using https://github.com/sourcegraph/deploy-k8s-helper.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that basic code intelligence works on Go or TypeScript.
    - [ ] Tear down this Kubernetes cluster.
- [ ] Delete entries from section 15 (CHANGELOG) of the testing grid, or move them into permanent sections above. Add new CHANGELOG items for this release into section 15. Assign (1) the feature owner and (2) a person who did not work on the feature as the testers for each row.
- [ ] Send a message to #dev-announce to kick off testing day.
  - [ ] Include a link to the testing grid.
  - [ ] Include the command to run the latest release candidate:
    ```
    IMAGE=sourcegraph/server:MAJOR.MINOR.0-rc.1 ./dev/run-server-image.sh
    ```
  - [ ] Mention that testing is the top priority, it is expected to take the whole day, and that known or suspected regressions should be tagged as release blockers. Mention that, for other issues found, high-level details like customer impact should be added to help Product determine whether it's a release blocker.

## 3 working days before release ($THREE_WORKING_DAYS_BEFORE_RELEASE)

- [ ] Send a message to #dev-announce to report whether any [release blocking issues](releases.md#blocking) were found.
- [ ] Add any [release blocking issues](releases.md#blocking) as checklist items here and start working to resolve them.
- [ ] Review all open issues in the release milestone that aren't blocking and ask assignees to triage them to a different milestone (backlog preferred).

## As necessary

- `git cherry-pick` bugfix (not feature!) commits from `master` into the release branch.
- Aggressively revert features that may cause delays.
- Re-test any flows that might have been impacted by commits that have been cherry picked into the release branch.
- Tag additional release candidates.

## 1 working day before release ($ONE_WORKING_DAY_BEFORE_RELEASE)

- [ ] **HH:MM AM/PM PT** Tag the final release.
    ```
    VERSION='vMAJOR.MINOR.0' bash -c 'git tag -a "$VERSION" -m "$VERSION" && git push origin "$VERSION"'
    ```
- [ ] Send a message to #dev-announce to announce the final release.
- [ ] Verify that all changes that have been cherry picked onto the release branch have been moved to the appropriate section of the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/master/CHANGELOG.md) on `master`.
- [ ] Wait for the final Docker images to be available at https://hub.docker.com/r/sourcegraph/server/tags.
- [ ] In [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph):
    - [ ] Wait for Renovate to open a PR to update the image tags and merge that PR ([example](https://github.com/sourcegraph/deploy-sourcegraph/pull/199)).
    - [ ] Create the `MAJOR.MINOR` release branch from this commit.
    - [ ] Tag the `vMAJOR.MINOR.0` release at this commit.
        ```
        VERSION='vMAJOR.MINOR.0' bash -c 'git tag -a "$VERSION" -m "$VERSION" && git push origin "$VERSION"'
        ```
- [ ] Open (but do not merge) PRs that do the following:
    - [ ] Update the documented version of Sourcegraph ([example](https://github.com/sourcegraph/sourcegraph/pull/2370/commits/701780fefa5809abb16669c9fb29738ec3bb2039)).
    ```
    find . -type f -name '*.md' -exec sed -i '' -E 's/sourcegraph\/server:[0-9\.]+/sourcegraph\/server:MAJOR.MINOR.0/g' {} +
    # Or use ruplacer
    ruplacer --go -t md 'sourcegraph/server:[0-9\.]+' 'sourcegraph/server:MAJOR.MINOR.0'
    ```
    - [ ] Update `latestReleaseKubernetesBuild` and `latestReleaseDockerServerImageBuild` ([example](https://github.com/sourcegraph/sourcegraph/pull/2370/commits/15925f2769564225e37013acb52d9d0b30e1336c)).
    - [ ] [Update deploy-aws version](https://github.com/sourcegraph/deploy-sourcegraph-aws/edit/master/ec2/resources/user-data.sh#L3)
    - [ ] [Update deploy-digitalocean version ](https://github.com/sourcegraph/deploy-sourcegraph-digitalocean/edit/master/resources/user-data.sh#L3)
    - [ ] Message @slimsag on Slack: `MAJOR.MINOR.PATCH has been released, update deploy-sourcegraph-docker as needed`
    - [ ] Update versions in docs.sourcegraph.com header ([example](https://github.com/sourcegraph/sourcegraph/pull/2701/commits/386e5ecb5225ab9c8ccc9791b489160ed7c984a2))
- [ ] Review all issues in the release milestone. Backlog things that didn't make it into the release and ping issues that still need to be done for the release (e.g. Tweets, marketing).
- [ ] Verify that the blog post is ready to be merged.

## By 10am PT on the 20th

- [ ] Merge the docs PRs created in the previous step.
- [ ] Merge the blog post ([example](https://github.com/sourcegraph/about/pull/83)).
- [ ] Close this issue.
- [ ] Close the milestone.
- [ ] Notify the next release captain that they are on duty for the next release. Include a link to this release issue template.
- [ ] Remind the team that they should submit [retrospective feedback](retrospectives/index.md) 24 hours before the scheduled retrospective meeting.

## After the Retrospective

- [ ] Scrub the retrospective Google doc for any priviledged customer data.
- [ ] [Convert to Markdown](https://gsuite.google.com/marketplace/app/docs_to_markdown/700168918607).
- [ ] Add a new retrospective page in `sourcegraph/doc../../team/product-dev/retrospectives`.
- [ ] Add a link to it on `retrospectives/index.md` and in the left nav (`sourcegraph/doc/_resources/templates/doc.html`)
