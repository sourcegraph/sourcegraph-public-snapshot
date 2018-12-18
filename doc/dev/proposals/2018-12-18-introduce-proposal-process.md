# Introduce proposal process

This is a proposal to define a proposal process.

## Background

Our engineering planning process has evolved over time.

From 2017-08 to 2018-09, we checked in proposals to [sourcegraph/docs-private/](https://github.com/sourcegraph/docs-private/).
- We were on a monthly planning and release cycle (1 week planning, 2 weeks build, 1 week endgame and testing).
- It was hard to plan N projects in one week and get them reviewed and approved by the smaller number of appropriate owners/reviewers.
- 1 week wasn't enough time to fully spec out certain projects so the proposal documents had underspecified design that needed to be figured out during the course of the project (but was never retroactively reflected in the proposal).

Around October 2018, we made some changes to our planning process:
- We decided to decouple planning from our release cycle.
    - This allowed natually longer projects to not be interrupted by our release cycle.
    - This naturally staggered project planning throughout the month.
    - Planning for a project could take as long (or as short) as necessary.
- We decided to capture plans in [GitHub issues](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Aroadmap) and require more explicit checklists (which could be dynamically updated to communicate progress).

In December 2018 we onboarded two teammates and realized three pain points:
1. GitHub issues are not the ideal medium to iterate on a design.
    - If you continuously update the original issue description to reflect the best known design, then it makes it hard to follow the subsequent conversation on the issue.
    - It is possible to see the previous states of the issue description in the GitHub UI, but it isn't obvious and it isn't coupled with the conversation that motivated those changes.
    - [Example](https://github.com/sourcegraph/sourcegraph/issues/1467)
2. GitHub issues are not an ideal store of important design decisions because it is mixed with other things like bugs.
3. We have out of date documentation for services. It is generally hard to keep separate documentation up to date, so as a first step we should focus on documenting changes.

Other reading:
- [Scaling Engineering Teams via Writing Things Down and Sharing - aka RFCs](https://blog.pragmaticengineer.com/scaling-engineering-teams-via-writing-things-down-rfcs/)
- [Go proposal template](https://github.com/golang/proposal/blob/master/design/TEMPLATE.md)
- [Rust RFC template](https://github.com/rust-lang/rfcs/blob/master/0000-template.md)

## Proposal

Adopt the proposal process documented in [README.md](README.md).

## Rationale

This is the next step in the evolution of our planning process.
- We want to re-capture the benefits of the old process (written, searchable, record of markdown files for why we did things) now that we no longer have the same constraints that caused us to switch away from the old docs-private process (i.e. get all planning for the month done in 1 week).
- We want to maintain the benefits of the current process:
  - Explicit checklist for action items serves as a proof of understanding what needs to happen.
  - Checklist is dyncamically updated to communicate the current state of the project.

## Implementation

- [ ] Document the proposal process in [README.md](README.md)
