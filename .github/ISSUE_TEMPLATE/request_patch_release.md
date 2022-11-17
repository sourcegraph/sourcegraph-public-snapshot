---
name: Request patch release
about: Sourcegraph teams, use this issue to propose a patch release or include your changes in a patch release.
title: '$MAJOR.$MINOR.$PATCH patch release request'
labels: 'release-guild,patch-release-request'
assignees: ''

---

<!-- Update the title according to the relevant patch release -->

@sourcegraph/release-guild I am requesting the following commits be included in a patch release. They are already merged into `main`:

- <!-- LINK TO EXACT MERGED COMMITS HERE -->

---

The intent of the questions below is to ensure we keep Sourcegraph high quality and [only create patch releases based on a strict criteria.](https://handbook.sourcegraph.com/engineering/releases#patch-releases) If you can answer yes to many or most of these questions, we will be happy to create the patch release.

I have read [when and why we perform patch releases](https://handbook.sourcegraph.com/engineering/releases#patch-releases) and answer the questions as follows:

> Are users/customers actively asking us for these changes and cannot wait until the next full release?

<!-- DO NOT MENTION CUSTOMERS BY NAME – this is a public repo. Instead find the customer here and link to that issue https://github.com/sourcegraph/accounts/issues -->
<!-- ANSWER THIS, include links to customer issue tracker -->

> Are the changes extremely minimal, well-tested, and low risk such that not testing as we do in a full release is OK?

<!-- **ANSWER THIS, explain in detail** -->

> Is there some functionality completely broken that warrants redacting the prior release of Sourcegraph and advising users wait for the patch release?

<!-- **ANSWER THIS, explain if needed** -->

> This will interrupt our regular planned work and release cycle, taking one full working day of our time, and will take up all of our site admin's valuable time by asking them to upgrade or producing noise for them if they don't need to upgrade.
>
> Do you believe the changes are important enough to warrant this?

<!-- **ANSWER THIS, yes/no** -->

> Patch releases are a signal we can do something better to improve the quality of Sourcegraph. Have you already scheduled a call (or created a google doc) to perform a [retrospective](https://about.sourcegraph.com/retrospectives) and identify ways we can improve?

<!-- **ANSWER THIS** -->

---

**For the [Release Captain]** - after reviewing this request:

- [ ] **Comment on this issue** with a decision regarding the request.
- [ ] If you are a first-time [Release Captain], please review the high-level overview of the [patch release process].
- [ ] If approved, **add it to a patch release**:
  - If there is [already an upcoming patch release](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aissue+label%3Arelease-tracking+), add the listed commits alongside a link to this issue.
  - If there is no upcoming patch release, create a new one:
    - [ ] Update [`dev/release/release-config.jsonc`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/release/release-config.jsonc), open and merge a PR to `main` to update it.
      - [ ] Change `upcomingRelease` to the current patch release
      - [ ] Change `previousRelease` to the previous patch release version
      - [ ] Change `releaseDate` to the current date (time is optional) along with `oneWorkingDayAfterRelease` and `threeWorkingDaysBeforeRelease`
      - [ ] Change `captainSlackUsername` and `captainGitHubUsername` to the patch captain's
    - [ ] Run `yarn release tracking:issues` on `main`
    - [ ] Add the listed commits alongside a link to this issue to the generated [release tracking issue](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aissue+label%3Arelease-tracking+)
- [ ] **Comment and close this issue once the relevant commit(s) have been cherry-picked into the release branch**.

[release captain]: https://handbook.sourcegraph.com/engineering/releases#release-captain
[patch release process]: https://handbook.sourcegraph.com/departments/product-engineering/engineering/process/releases#patch-release-process
