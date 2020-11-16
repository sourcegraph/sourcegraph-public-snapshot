---
name: Request patch release
about: Sourcegraph teams, use this issue to request the Distribution team perform a patch release or include your changes in a patch release..
title: ''
labels: 'team/distribution,patch-release-request'
assignees: ''

---

@sourcegraph/distribution I am requesting the following commits be included in a patch release. They are already merged into `main`:

The intent of the questions below is to ensure we keep Sourcegraph high quality and [only create patch releases based on a strict criteria.](https://about.sourcegraph.com/handbook/engineering/releases#when-are-patch-releases-performed) If you can answer yes to many or most of these questions, we will be happy to create the patch release.

- <!-- LINK TO EXACT MERGED COMMITS HERE -->

I have read [when and why we perform patch releases](https://about.sourcegraph.com/handbook/engineering/releases#when-are-patch-releases-performed) and answer the questions as follows:

> Are users/customers actively asking us for these changes and cannot wait until the next full release?

<!-- ANSWER THIS, include links to customer issue tracker -->

> Are the changes extremely minimal, well-tested, and low risk such that not testing as we do in a full release is OK?

<!-- **ANSWER THIS, explain in detail** -->

> Is there some functionality completely broken that warrants redacting the prior release of Sourcegraph and advising users wait for the patch release?

<!-- **ANSWER THIS, explain if needed** -->

> This will interrupt our regular planned work and release cycle, taking 3-6 hours of our time, and will take up all of our site admin's valuable time by asking them to upgrade or producing noise for them if they don't need to upgrade.
>
> Do you believe the changes are important enough to warrant this?

<!-- **ANSWER THIS, yes/no** -->

> Patch releases are a signal we can do something better to improve the quality of Sourcegraph. Have you already scheduled a call (or created a google doc) to perform a [retrospective](https://about.sourcegraph.com/retrospectives) and identify ways we can improve?

<!-- **ANSWER THIS** -->

---

**For the [release captain](https://about.sourcegraph.com/handbook/engineering/releases#release-captain)** - after reviewing and approving this request:

- If there is [already an upcoming patch release](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aissue+label%3Arelease-tracking+), add the listed commits alongside a link to this issue
- If there is no upcoming patch release, create a new one:
  - Update [`dev/release/release-config.jsonc`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/release/release-config.jsonc) with the patch release in `upcomingRelease` (and open a PR to `main` to update it)
  - `cd dev/release && yarn build && yarn run release tracking:patch-issue`
  - Add the listed commits alongside a link to this issue to the generated [release tracking issue](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aissue+label%3Arelease-tracking+)

Comment and close this issue once the relevant commit(s) have been cherry-picked into the release branch.
