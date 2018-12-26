# Define planning process

This is a proposal to define the planning process for product and engineering.

## Background

Our product and engineering planning process has evolved over time.

From 2017-08 to 2018-09, we committed plans to [sourcegraph/docs-private/](https://github.com/sourcegraph/docs-private/).

- We were on a monthly planning and release cycle (1 week planning, 2 weeks build, 1 week endgame and testing).
- It was hard to plan N projects in one week and get them reviewed and approved by the smaller number of appropriate owners/reviewers.
- 1 week wasn't enough time to fully spec out certain projects so the plan documents had underspecified design that needed to be figured out during the course of the project (but was never retroactively reflected in the proposal).

Around October 2018, we made some changes to our planning process:

- We decided to decouple planning from our release cycle.
    - This allowed natually longer projects to not be interrupted by our release cycle.
    - This naturally staggered project planning throughout the month.
    - Planning for a project could take as long (or as short) as necessary.
- We decided to capture plans in [GitHub issues](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Aroadmap) and require more explicit checklists (which could be dynamically updated to communicate progress).

Since then, we have felt new pain points:

1. Most projects were behind schedule because the work had not been adequately planned.
2. GitHub issues are not the ideal medium to iterate on a design.
    - If you continuously update the original issue description to reflect the best known design, then it makes it hard to follow the subsequent conversation on the issue.
    - It is possible to see the previous states of the issue description in the GitHub UI, but it isn't obvious and it isn't coupled with the conversation that motivated those changes.
    - [Example](https://github.com/sourcegraph/sourcegraph/issues/1467)
3. GitHub issues are not an ideal store of important design decisions because they are intermixed with other things like bugs.
4. We have out of date documentation for services. It is generally hard to keep separate documentation up to date, so as a first step we should focus on documenting changes.

Other reading:

- [Scaling Engineering Teams via Writing Things Down and Sharing - aka RFCs](https://blog.pragmaticengineer.com/scaling-engineering-teams-via-writing-things-down-rfcs/)
- [Go proposal template](https://github.com/golang/proposal/blob/master/design/TEMPLATE.md)
- [Rust RFC template](https://github.com/rust-lang/rfcs/blob/master/0000-template.md)

## Plan

Adopt the planning process documented in [README.md](README.md).

### Test plan

The new process will be tested by using it and iterating on it based on our experience.

### Release plan

All new projects will use this planning process.

### Success metrics

This project is successful if/when every engineer's primary project is planned using this process.

### Company goals

This will help us plan and prioritize all of the features that will help us achieve our [Q1 goal to increase the number of companies that get from 1 to 20 daily users](company-goals.md#2018-Q1).

## Rationale

This is the next incremental step in the evolution of our planning process.

- We want to re-capture the benefits of the old process (written, searchable, record of markdown files for why we did things) now that we no longer have the same constraints that caused us to switch away from the old docs-private process (i.e. get all planning for the month done in 1 week).
- We want to maintain the benefits of the current process:
  - Explicit checklist for action items serves as a proof of understanding what needs to happen.
  - Checklist is dynamically updated to communicate the current state of the project.

## Checklist 

- [ ] Merge https://github.com/sourcegraph/sourcegraph/pull/1492 by Dec 21, 2018 to unblock the creation of new proposals.
- [ ] Discuss proposal process during team meeting on Jan 7, 2019
- [ ] Get written confirmation for each engineer that they have read https://github.com/sourcegraph/sourcegraph/pull/1492 and have raised any concerns by Jan 14, 2019.
- [ ] Have open or merged proposal PRs for all projects in the 3.1 milestone by Jan 22, 2019.

Live checklist: https://github.com/sourcegraph/sourcegraph/issues/1534

## Done date

Jan 22, 2019

## Retrospective

[This section is completed after the project is completely done (i.e. the checklist is complete).]

### Actual checklist

[What is the actual checklist the you completed (i.e. paste the final checklist from the issue here)? Explain any differences from the original checklist in the proposal.]

### Actual done date

[What is the date that the project was actually finished? Explain why this is earlier or later than originally planned or explain why the project was not completed.]