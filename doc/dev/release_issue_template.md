<!--
This template is used for our monthly major/minor releases of Sourcegraph.
It is not used for patch releases.
-->

Release version: MAJOR.MINOR
Release date: YYYY-MM-DD

No later than 5 working days before release (YYYY-MM-DD)
- [ ] Choose dates/times for the steps in this release process and update this issue template accordingly. Note that this template references _working days_, which do not include weekends or holidays observed by Sourcegraph.
- [ ] Send message to #dev-announce with a link to this tracking issue to notify the team of the release schedule.
- [ ] Private message each teammate who has open issues in the milestone and ask them to remove any issues that won't be done three working days before the release.

3 working days before release (YYYY-MM-DD)
- [ ] **HH:MM AM/PM PT** Create the branch for this release off of master and tag the first release candidate (e.g. `v3.0.0-rc.1`).
- [ ] Send a message to #dev-announce to announce the release candidate.
- [ ] Run Sourcegraph Docker image with no previous data.
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
    - [ ] Initialize the site by creating an admin account.
    - [ ] Add a public repository (i.e. https://github.com/sourcegraph/sourcegraph).
    - [ ] Add a private repository (i.e. https://github.com/sourcegraph/infrastructure).
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

1 working day before the release (YYYY-MM-DD)
- [ ] **HH:MM AM/PM PT** Tag a final release candidate. Barring a new [release blocking](releases.md#blocking) issue, this will become the final release.
- [ ] Send a message to #dev-announce to announce the release candidate.
- [ ] Re-test any flows that might have been impacted by commits that have been `git cherry-picked` into the release branch.

On or before release day (YYYY-MM-DD)
- [ ] **HH:MM AM/PM PT** Tag the final release and wait for the Docker images to get pushed (make sure you do this in advance of when you want to merge/publish the blog post).
- [ ] **HH:MM AM/PM PT** Merge the blog post.
- [ ] Review all issues in the release milestone. Backlog things that didn't make it into the release and ping issues that still need to be done for the release (e.g. Tweets, marketing).
- [ ] Close this issue.
