<!--
This template is used for our monthly major/minor releases of Sourcegraph.
It is not used for patch releases. See [patch_release_issue_template.md](patch_release_issue_template.md)
for the patch release checklist.
-->

# MAJOR.MINOR Release (YYYY-MM-20)

## At the start of the month (YYYY-MM-01)

- [ ] Choose dates/times for the steps in this release process and update this issue template accordingly. Note that this template references _working days_, which do not include weekends or holidays observed by Sourcegraph.
- [ ] Add events to the shared Release Schedule calendar in Google and invite team@sourcegraph.com.
    - [ ] Creating the release branch.
    - [ ] Tagging the final release.
    - [ ] Publishing the blog post.
- [ ] Send message to #dev-announce with a link to this tracking issue to notify the team of the release schedule.
- [ ] Create the [retrospective document](retrospectives/index.md) and schedule the retrospective meeting within a few days _after_ the release (send calendar invites to team@sourcegraph.com).
- [ ] Create reminders for yourself to preform the remaining sections in this checklist at appropriate times.

## 5 working days before release (YYYY-MM-DD)

- [ ] Private message each teammate who has open issues in the milestone and ask them to remove any issues that won't be done by the time that the release branch is scheduled to be created.
- [ ] Verify that there is a draft of the blog post and that it will be ready to be merged on time.

## 3 working days before release (YYYY-MM-DD)

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
        CLEAN=true IMAGE=sourcegraph/server:vMAJOR.minor.0-rc.1 ./dev/run-server-image.sh
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
        CLEAN=true IMAGE=sourcegraph/server:vMAJOR.minor.0-rc.1 ./dev/run-server-image.sh
        ```
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that basic code intelligence works on Go or TypeScript.
- [ ] Run the new version of Sourcegraph on a clean Kubernetes cluster with no previous data.
    - [ ] Log in to https://console.cloud.google.com, select `sourcegraph-dev` project and create the new Kubernetes cluster.
    - [ ] Set up the new version of Sourcegraph on that cluster using our docs.
    - [ ] Initialize the site by creating an admin account.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that basic code intelligence works on Go or TypeScript.
    - [ ] Tear down this Kubernetes cluster.
- [ ] Upgrade Sourcegraph on a Kubernetes cluster.
    - [ ] Log in to https://console.cloud.google.com, select `sourcegraph-dev` project and create the new Kubernetes cluster.
    - [ ] Set up an old version of Sourcegraph on that cluster using our docs.
    - [ ] Initialize the site by creating an admin account.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
    - [ ] Upgrade the cluster to the new release.
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that basic code intelligence works on Go or TypeScript.
    - [ ] Tear down this Kubernetes cluster.
- [ ] Send a message to #dev-announce to report whether any [release blocking issues](releases.md#blocking) were found.
- [ ] Add any [release blocking issues](releases.md#blocking) as checklist items here and start working to resolve them.
- [ ] Review all open issues in the release milestone that aren't blocking and ask assignees to triage them to a different milestone (backlog preferred).
- [ ] Remind the team that they should submit [retrospective feedback](retrospectives/index.md) 24 hours before the scheduled retrospective meeting.

## As necessary

- `git cherry-pick` bugfix (not feature!) commits from `master` into the release branch.
- Re-test any flows that might have been impacted by commits that have been cherry picked into the release branch.
- Tag additional release candidates.

## 1 working day before release (YYYY-MM-DD)

- [ ] **HH:MM AM/PM PT** Tag the final release.
    ```
    VERSION='vMAJOR.MINOR.0' bash -c 'git tag -a "$VERSION" -m "$VERSION" && git push origin "$VERSION"'
    ```
- [ ] Send a message to #dev-announce to announce the final release.
- [ ] Verify that all changes that have been cherry picked onto the release branch have been moved to the approriate section of the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/master/CHANGELOG.md) on `master`.
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
    find . -type f -name '*.md' -exec sed -i '' -E 's/sourcegraph\/server:[0-9\.]+/sourcegraph\/server:$NEWVERSION/g' {} +
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
- [ ] Notify the next release captain that they are on duty for the next release.
