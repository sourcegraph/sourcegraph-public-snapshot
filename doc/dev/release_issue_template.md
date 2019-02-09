<!--
This template is used for our monthly major/minor releases of Sourcegraph.
It is not used for patch releases.
-->

# MAJOR.MINOR Release (YYYY-MM-DD)

## No later than 5 working days before release (YYYY-MM-DD)

- [ ] Choose dates/times for the steps in this release process and update this issue template accordingly. Note that this template references _working days_, which do not include weekends or holidays observed by Sourcegraph.
- [ ] Send message to #dev-announce with a link to this tracking issue to notify the team of the release schedule.
- [ ] Private message each teammate who has open issues in the milestone and ask them to remove any issues that won't be done three working days before the release.

## 3 working days before release (YYYY-MM-DD)

- [ ] **HH:MM AM/PM PT** Create the branch for this release off of master (e.q. `3.0`) and tag the first release candidate (e.g. `v3.0.0-rc.1`).
- [ ] Add a new section header for this version to the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/master/CHANGELOG.md#unreleased) immediately under the `## Unreleased changes` heading. <!-- TODO(nick): link to example change --> Commit and `git push` this change directly to upstream `master`.
- [ ] Send a message to #dev-announce to announce the release candidate.
- [ ] Run Sourcegraph Docker image with no previous data.
    - [ ] `CLEAN=true IMAGE=sourcegraph/server:3.0.2$NEWVERSION ./dev/run-server-image.sh`
    - [ ] Initialize the site by creating an admin account.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that code intelligence works as you expect.
        - [ ] Go to definition works in Go.
        - [ ] Go to definition works in TypeScript.
        - [ ] Find references works in Go.
        - [ ] Find references works in TypeScript.
- [ ] Upgrade Sourcegraph Docker image from previous released version.
    - [ ] `CLEAN=true IMAGE=sourcegraph/server:3.0.2$OLDVERSION ./dev/run-server-image.sh`
    - [ ] Initialize the site by creating an admin account.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
    - [ ] `CLEAN=false IMAGE=sourcegraph/server:3.0.2$NEWVERSION ./dev/run-server-image.sh`
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that code intelligence works as you expect.
        - [ ] Go to definition works in Go.
        - [ ] Go to definition works in TypeScript.
        - [ ] Find references works in Go.
        - [ ] Find references works in TypeScript.
- [ ] Run Sourcegraph on a clean Kubernetes cluster with no previous data.
    - [ ] Initialize the site by creating an admin account.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that code intelligence works as you expect.
        - [ ] Go to definition works in Go.
        - [ ] Go to definition works in TypeScript.
        - [ ] Find references works in Go.
        - [ ] Find references works in TypeScript.
- [ ] Upgrade Sourcegraph on a Kubernetes cluster.
    - [ ] Initialize the site by creating an admin account.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
    - [ ] Verify that code search returns results as you expect (depending on the repositories that you added).
    - [ ] Verify that code intelligence works as you expect.
        - [ ] Go to definition works in Go.
        - [ ] Go to definition works in TypeScript.
        - [ ] Find references works in Go.
        - [ ] Find references works in TypeScript.
- [ ] Add any [release blocking](releases.md#blocking) issues as checklist items and start working to resolve them.

_Cherry pick commits from `master` into the release branch and tag new release candidates as necessary._

## 1 working day before the release (YYYY-MM-DD)

- [ ] Re-test any flows that might have been impacted by commits that have been `git cherry-picked` into the release branch.
- [ ] **HH:MM AM/PM PT** Tag the final release.
- [ ] Send a message to #dev-announce to announce the final release.
- [ ] Verify that all changes that have been cherry picked onto the release branch have been moved to the approriate section of the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/master/CHANGELOG.md) on `master`.
- [ ] Open a PR that updates the documented version of Sourcegraph (e.g. `find . -type f -name '*.md' -exec sed -i '' -E 's/sourcegraph\/server:[0-9\.]+/sourcegraph\/server:3.0.2/g' {} +`) and `latestReleaseDockerServerImageBuild` (e.g. https://github.com/sourcegraph/sourcegraph/pull/2210/files).
- [ ] Merge the above PR above after the final Docker images are available at https://hub.docker.com/r/sourcegraph/server/tags.
- [ ] [Update the image tags](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/README.dev.md#updating-docker-image-tags) in [deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) (e.g. [d0bb80](https://github.com/sourcegraph/deploy-sourcegraph/commit/d0bb80f559e7e9ef3b1915ddba72f4beff32276c))
- [ ] Wait for Renovate to pin the hashes in deploy-sourcegraph (e.g. [#190](https://github.com/sourcegraph/deploy-sourcegraph/pull/190/files)), or pin them yourself.
- [ ] Tag the release in deploy-sourcegraph (e.g. `v3.0.0`).
- [ ] Create a release branch off of master at this commit (e.g. `v3.0`).
- [ ] Update `latestReleaseKubernetesBuild` in `master`. <!-- TODO(nick): example -->

## On or before release day (YYYY-MM-DD)

- [ ] **HH:MM AM/PM PT** Merge the blog post.
- [ ] Review all issues in the release milestone. Backlog things that didn't make it into the release and ping issues that still need to be done for the release (e.g. Tweets, marketing).
- [ ] Close this issue.
