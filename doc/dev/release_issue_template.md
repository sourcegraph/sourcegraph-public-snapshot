<!--
This template is used for our monthly major/minor releases of Sourcegraph.
It is not used for patch releases.
-->

# MAJOR.MINOR Release (YYYY-MM-DD)

## No later than 5 working days before release (YYYY-MM-DD)

- [ ] Choose dates/times for the steps in this release process and update this issue template accordingly. Note that this template references _working days_, which do not include weekends or holidays observed by Sourcegraph.
- [ ] Add events to the shared Release Schedule calendar in Google and invite team@sourcegraph.com.
    - [ ] Creating the release branch.
    - [ ] Tagging the final release.
    - [ ] Publishing the blog post.
- [ ] Send message to #dev-announce with a link to this tracking issue to notify the team of the release schedule.
- [ ] Private message each teammate who has open issues in the milestone and ask them to remove any issues that won't be done three working days before the release.

## 3 working days before release (YYYY-MM-DD)

- [ ] **HH:MM AM/PM PT** Add a new section header for this version to the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/master/CHANGELOG.md#unreleased) immediately under the `## Unreleased changes` heading and add new empty sections under `## Unreleased changes` (e.g. [#2323](https://github.com/sourcegraph/sourcegraph/pull/2323)).
- [ ] Create the `MAJOR.MINOR` branch for this release off of the changelog commit that you created in the previous step.
- [ ] Tag the first release candidate `MAJOR.MINOR.0-rc.1`.
- [ ] Send a message to #dev-announce to announce the release candidate.
- [ ] Run Sourcegraph Docker image with no previous data.
    - [ ] Run the new version of Sourcegraph.
        ```
        CLEAN=true IMAGE=sourcegraph/server:$NEWVERSION ./dev/run-server-image.sh
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
        CLEAN=false IMAGE=sourcegraph/server:$NEWVERSION ./dev/run-server-image.sh
        ```
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that basic code intelligence works on Go or TypeScript.
- [ ] Run the new version of Sourcegraph on a clean Kubernetes cluster with no previous data.
    - [ ] Initialize the site by creating an admin account.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that basic code intelligence works on Go or TypeScript.
    - [ ] Tear down this Kubernetes cluster.
- [ ] Upgrade Sourcegraph on a Kubernetes cluster.
    - [ ] Setup a Kubernetes cluster that is running the previous release.
    - [ ] Initialize the site by creating an admin account.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
    - [ ] Upgrade the cluster to the new release.
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that basic code intelligence works on Go or TypeScript.
    - [ ] Tear down this Kubernetes cluster.
- [ ] Send a message to #dev-announce to report whether any [release blocking issues](releases.md#blocking) were found.
- [ ] Add any [release blocking issues](releases.md#blocking) as checklist items here and start working to resolve them.

## As necessary

- `git cherry-pick` bugfix (not feature!) commits from `master` into the release branch.
- Re-test any flows that might have been impacted by commits that have been cherry picked into the release branch.
- Tag additional release candidates.

## 1 working day before the release (YYYY-MM-DD)

- [ ] **HH:MM AM/PM PT** Tag the final release.
- [ ] Send a message to #dev-announce to announce the final release.
- [ ] Verify that all changes that have been cherry picked onto the release branch have been moved to the approriate section of the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/master/CHANGELOG.md) on `master`.
- [ ] Wait for the final Docker images to be available at https://hub.docker.com/r/sourcegraph/server/tags.
- [ ] [Update the image tags](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/README.dev.md#updating-docker-image-tags) in [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) (e.g. [d0bb80](https://github.com/sourcegraph/deploy-sourcegraph/commit/d0bb80f559e7e9ef3b1915ddba72f4beff32276c))
- [ ] Wait for Renovate to pin the hashes in deploy-sourcegraph (e.g. [#190](https://github.com/sourcegraph/deploy-sourcegraph/pull/190/files)), or pin them yourself.
- [ ] In deploy-sourcegraph, tag the release (e.g. `v3.0.0`) and create the release branch off of master at this commit (e.g. `v3.0`).

## On or before release day (YYYY-MM-DD)

- [ ] **HH:MM AM/PM PT** Merge the blog post. <!-- TODO(nick): example -->
- [ ] Update the documented version of Sourcegraph in `master` (e.g. https://github.com/sourcegraph/sourcegraph/pull/2210/files <!-- TODO(nick): example that doesn't include latestReleaseDockerServerImageBuild -->).
    ```
    find . -type f -name '*.md' -exec sed -i '' -E 's/sourcegraph\/server:[0-9\.]+/sourcegraph\/server:$NEWVERSION/g' {} +
    ```
- [ ] Update versions in docs.sourcegraph.com header ([example](https://github.com/sourcegraph/docs.sourcegraph.com/commit/d445c460c2da54fba4f56f647d656ca3311decf5))
- [ ] Update `latestReleaseKubernetesBuild` and `latestReleaseDockerServerImageBuild` in `master`. <!-- TODO(nick): example -->
- [ ] Review all issues in the release milestone. Backlog things that didn't make it into the release and ping issues that still need to be done for the release (e.g. Tweets, marketing).
- [ ] Close this issue.
