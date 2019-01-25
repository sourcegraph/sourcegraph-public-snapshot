# Product

This document is about *how* we do product at Sourcegraph. For the *what*, see the [product roadmap](../roadmap/index.md).

## Goals

The goals of product at Sourcegraph are to make the following true:

- The team is working on the most important things (listed in the [product roadmap](../roadmap/index.md)) to achieve our company goals.
- Each teammate has the customer and product context needed (about customer problems, likely future priorities, possible solutions, etc.) to perform their work effectively.
- The product vision and roadmap are communicated well to teammates and everyone outside Sourcegraph.

## Planning

We plan each project with the project team at least through the [current release](../releases.md). The outcome of planning a project is that:

- The [product roadmap](../roadmap/index.md) describes what the project will ship (by linking to tracking issues). The tracking issue is the source of truth for the status of a project and should be the primary planning artifact. TODO!(sqs): add tracking issue template, possibly remove plans/0000-00-00-template.md
- The [issues](../issues.md) for the project are correctly prioritized, assigned, and documented.
- The project team has the context necessary to work effectively and to [triage issues](../issues.md#triage) that are filed between now and the next planning session.

Product is responsible for each project being planned at least through the [current release](../releases.md). How far out a project is planned depends on the project. Planning for a project can happen at any time, not just between releases. However, around the time we ship a release, product should review the plans for all projects in the next release for overall coherency.

### Meetings

Project planning is a continuous process, punctuated with meetings to ensure everyone is on the same page.

The agenda for a project planning kickoff or checkin is:

1. Review, update, and/or create the project's tracking issue on the [product roadmap](../roadmap/index.md).
   - Use the [**tracking issue template**](tracking_issue_template.md).
   - Are they the most important things to work on?
   - Are they realistic?
   - Are they accurate and up to date?
1. Review, update, and/or create the project's issues.
   - Are they the best way to accomplish what the roadmap tracking issues say?
   - Are they assigned to the right person?
   - Are they added to the right milestone?
   - Are there hidden blockers?
   - Are they accurate and up to date?

Projects should also hold [retrospectives](../retrospectives/index.md), and this may be combined with planning meetings.

#### Related reading

- [Scaling Engineering Teams via Writing Things Down and Sharing - aka RFCs](https://blog.pragmaticengineer.com/scaling-engineering-teams-via-writing-things-down-rfcs/)
- [Go proposal template](https://github.com/golang/proposal/blob/master/design/TEMPLATE.md)
- [Rust RFC template](https://github.com/rust-lang/rfcs/blob/master/0000-template.md)

## Saying "no"

We receive tons of feature request and bug reports, more than we can handle. This means we must frequently say "no" or prioritize things less urgently than some people would like. Our job is to find the most important things to work on.

### Product manager

The product manager is responsible for prioritization, which means ensuring the following is true:

> The issues for each release milestone are the most important things to work on.

This is a broader sense of the term "prioritization" that means the product manager must do 3 things:

- **Educate** teams so they can prioritize on their own.
- **Plan** with each team the set of issues to work on for each release.
- **Backstop:** triage unprioritized/misprioritized issues and file unfiled issues.

The right mix of these 3 things depends on the team and what they're working on (and should be agreed on with the team). For example:

- For some teams, the PM will educate and plan by meeting with the team once at the start of a release, and will have a very limited backstop role in between. The devs will be the first responders for triaging issues and will kick off the planning meeting with a well prioritized set of issues to work on.
- For other teams, the PM will be the first to respond to and prioritize issues that are filed on the team.

No matter the chosen mix, the PM is still responsible for prioritization. For example, if the PM doesn't effectively educate the devs so they can plan and triage well, then the PM needs to step up on planning and triage (as a backstop).

