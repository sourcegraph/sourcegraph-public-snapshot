# Product

This document is about *how* we do product at Sourcegraph. For the *what*, see the [product roadmap](../roadmap/index.md).

## Goals

The goals of product at Sourcegraph are to make the following true:

- The team is working on the most important things (listed in the [product roadmap](../roadmap/index.md)) to achieve our company goals.
- Each teammate has the customer and product context needed (about customer problems, likely future priorities, possible solutions, etc.) to perform their work effectively.
- The product vision and roadmap are communicated well to teammates and everyone outside Sourcegraph.

## Planning

The [product manager](#product-manager) and project team plan each project at least through the [current release](../releases.md). The outcome of planning a project is that:

- The project has a [tracking issue](tracking_issue_template.md) in each release milestone that ships something related to the project. The tracking issue describes what a project will ship in a specific release. It is the source of truth for the description and status of a project.
- The [product roadmap](../roadmap/index.md) links to the project's [tracking issues](tracking_issue_template.md).
- The [issues](../issues.md) for the project are correctly prioritized, assigned, and documented.
- The project team has the context necessary to work effectively and to [triage issues](../issues.md#triage) that are filed between now and the next planning session.

The [product manager](#product-manager) is responsible for each project being planned at least through the [current release](../releases.md). How far out a project is planned depends on the project.

### Product planning is continuous

Planning for a project is a continuous process and can happen at any time, not just between releases. However, the [product manager](#product-manager) should meet with the project team to check in within 1-2 weeks before each release. The product manager should also review the plans for all projects in the next release for overall coherency.

This ensures that projects can work on the schedules that make the most sense for them (subject to the constraint of needing to ship some kind of milestone monthly). A particularly long-term project can have many months of visibility into requirements and plans, and shorter or more experimental projects can be planned with shorter time horizons.

### Meetings

Project planning is a continuous process, punctuated with meetings to ensure everyone is on the same page.

The agenda for a project planning kickoff or checkin is:

1. Review, update, and/or create the project's tracking issue on the [product roadmap](../roadmap/index.md) for the next release.
   - Use the [**tracking issue template**](tracking_issue_template.md).
   - Are the tasks the most important things to work on?
   - Are they realistic?
   - Are they accurate and up to date?
1. Review, update, and/or create the project's issues in the next release milestone.
   - Are the tasks the best way to accomplish what the roadmap tracking issues say?
   - Are they assigned to the right person?
   - Are they added to the right milestone?
   - Are there hidden blockers?
   - Are they accurate and up to date?

If a project wants to do a project-specific [retrospective](../retrospectives/index.md) (in addition to the full-team release retrospective), they can combine it with the planning meeting for the next cycle.

### Related reading

- [Scaling Engineering Teams via Writing Things Down and Sharing - aka RFCs](https://blog.pragmaticengineer.com/scaling-engineering-teams-via-writing-things-down-rfcs/)
- [Go proposal template](https://github.com/golang/proposal/blob/master/design/TEMPLATE.md)
- [Rust RFC template](https://github.com/rust-lang/rfcs/blob/master/0000-template.md)

## Release early, release often

Each project, no matter how long-running, needs to plan to ship *something* in each release. The "something" depends on the project. We strongly prefer for it to be a minimal viable feature that is enabled by default. The next best thing is to ship something that is feature-flagged off by default.

The reason for this is to avoid going for too long without customer feedback (from customers trying it) or even technical/product feedback (from performing the diligent work of polishing it to be ready to release). Lacking these critical checks means we will end up building something that doesn't solve people's problems or that is over-built.

When we have relaxed this in the past, the results have been bad and the overwhelming feedback from retrospectives has been to release regularly.

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

