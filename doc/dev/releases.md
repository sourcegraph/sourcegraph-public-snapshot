# Releases

This document describes how we release Sourcegraph.

## Goal

The goal of our release process is to make releases boring, regular, and eventually, automated.

## Releases are monthly

We release Sourcegraph **by** 10am PT on the 20th day of each month.

"Release" means:

- The Docker images are available for download.
- The blog post is published.
- The release is documented on docs.sourcegraph.com.

The release always ships on time, even if it's missing features or bug fixes we hoped to get in ([why?](https://about.gitlab.com/2015/12/07/why-we-shift-objectives-and-not-release-dates-at-gitlab/)).

### Why the 20th?

We don't want to ship a release too late in December because the Sourcegraph team has a scheduled break December 24 through January 1.

### Why aren't releases continuous?

Although [Sourcegraph.com](https://sourcegraph.com) is continuously deployed (from sourcegraph/sourcegraph@`master`), the version of Sourcegraph that customers use is not continuously released or updated. This is because:

- We don't think customers would be comfortable with a continuously updated service running on their own infrastructure, for security and stability reasons.
- We haven't built the automated testing and update infrastructure to make continuous customer releases reliable.

In the future, we may introduce continuous releases if these issues become surmountable.

## Versioning

[Monthly releases](#releases-are-monthly) of Sourcegraph increase the minor version number (e.g. 3.1 -> 3.2). These releases **never** require any manual migration steps.

Patch releases (e.g. 3.0.0 -> 3.0.1) are released on an as-needed basis to fix bugs and security issues. These releases **never** require any manual migration steps. To create a patch release, create a tracking issue using the [patch release issue template](patch_release_issue_template.md) and complete all listed steps.

On rare occasions we may decide to increase the major version number (e.g. 2.13 -> 3.0). These releases **may** require manual migration steps.

## Release process

What is the process we follow to release?

### Release captain

The release captain is _responsible_ for managing the release process and ensuring that the release happens on time. The release captain may _delegate_ work to other teammates, but such delegation does not absolve the release captain of their responsibility to ensure that delegated work gets done.

The release captain should create a tracking issue using the [release issue template](release_issue_template.md) at the beginning of the release cycle.

### Release templates

- [Release issue template](release_issue_template.md)
- [Patch release issue template](patch_release_issue_template.md)
- [DevRel release issue template](devrel_release_issue_template.md)

### Schedule

| Version | Captain | Release Date |
|---------|---------|--------------|
| 3.0 | @nicksnyder | 2019-02-07 |
| 3.1 | @nicksnyder | 2019-02-20 |
| 3.2 | @nicksnyder | 2019-03-20 |
| 3.3 | @slimsag | 2019-04-20 |
| 3.4 | @beyang | 2019-05-20 |
| 3.5 | @ggilmore | 2019-06-20 |
| 3.6 | @keegancsmith | 2019-07-20 |
| 3.7 | @attfarhan | 2019-08-20 |
| 3.8 | @lguychard | 2019-09-20 |
| 3.9 | @tsenart| 2019-10-20 |
| 3.10 | @beyang | 2019-11-20 |
| 3.11 | @chrismwendt | 2019-12-20 |
| 3.12 | @uwedeportivo | 2020-01-20 |
| 3.13 | @felixfbecker | 2020-02-20 |
| 3.14 | @efritz | 2020-03-20 |
| 3.15 | @unknwon | 2020-04-20 |
| 3.16 | @rvantonder | 2020-05-20 |
| 3.17 | @hadrian-git | 2020-06-20 |

Release captains may trade rotations with each other by updating this schedule.

If a release captain is unexpectedly unavailable and did not arrange a replacement, the release captain's manager is responsible for identifying a new release captain.

### Release branches

Each major and minor release of [Sourcegraph](https://github.com/sourcegraph/sourcegraph) has a long lived release branch (e.g. `3.0`, `3.1`). Individual releases are tagged from these release branches (e.g. `v3.0.0-rc.1`, `v3.0.0`, `v3.0.1-rc.1`, and `v3.0.1` would be tagged from the `3.0` release branch).

To avoid confusion between tags and branches:

- Tags are always the full semantic version with a leading `v` (e.g. `v2.10.0`)
- Branches are always the dot-separated major/minor versions with no leading `v` (e.g. `2.10`).

Development always happens on `master` and changes are cherry picked onto release branch as necessary **with the approval of the release captain**.

#### Example

Here is an example git commit history:

1. The release captain creates the `3.0` release branch at commit `B`.
1. The release captain tags the release candidate `v3.0.0-rc.1` at commit `B`.
1. A feature is committed to `master` in commit `C`. It will not ship in `3.0`.
1. An issue is found in the release candidate and a fix is committed to `master` in commit `D`.
1. The release captain cherry picks `D` from `master` into `3.0`.
1. The release captain tags `v3.0.0` on the `3.0` release branch.
1. Development continues on master with commits `E`, `F`, `G`, `H`.
1. Commit `F` fixes a critical bug that impacts 3.0, so it is cherry picked onto the `3.0` release branch and `v3.0.1` is tagged.
1. The release captain (different person) for 3.1 creates the `3.1` release branch at commit `H` and a new release cycle begins.
1. Commit `J` fixes a critical bug that impacts both 3.0 and 3.1, so it is cherry picked into both `3.0` and `3.1` release branches and new releases are tagged (`v3.0.2`, `v3.1.2`).

```
A---B---C---D---E---F---G---H---I---J---K---L (master branch)
     \                       \
      \                       `---v3.1.0-rc.1---I'---v3.1.0---J'---v3.1.2 (3.1 release branch)
       \
        `---v3.0.0-rc.1---D'---v3.0.0---F'---v3.0.1---J'---v3.0.2 (3.0 release branch)
```

### Issues

How do we deal with issues that are found during the release process?

#### Blocking

The release always ships on time, even if it's missing features or bug fixes we hoped to get in ([why?](https://about.gitlab.com/2015/12/07/why-we-shift-objectives-and-not-release-dates-at-gitlab/)).

There are only three kinds of issues that are eligible to block a release:

1. Issues that literally prevent us from tagging a release (i.e. our CI logic to produce builds from git tags is broken).
2. Issues that fundamentally break our product for a _majority_ of our customers and don't have acceptable workarounds.
3. Critical security _regressions_ from the previous release.

Only the release captain can label something as release blocking.

The release captain has unlimited power to make changes to the release branch to resolve release blocking issues. As soon as a release blocking issue is identified, the release captain should decide the least risky way to resolve the issue as soon as possible. A good default action is to identify and revert offending commits from the release branch. In the worst case, this could involved recreating the release branch from an earlier commit on master. Project owners can work on master to fix the issue, and if the issue is resolved in time, revert the revert and cherry-pick the fix on the release branch.

#### Non-blocking

Most issues are non-blocking. Fixes to non-blocking issues can be fixed in `master` by the code owner who can then `git cherry-pick` those commits into the release branch with the approval of the release captain. Alternatively, broken features can be reverted out of the release branch or disabled via feature flags if they aren't ready or are too buggy.
