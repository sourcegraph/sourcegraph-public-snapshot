# Product

This document is about *how* we plan product changes at Sourcegraph. For the *what*, see the [product roadmap](../roadmap/index.md).

## Goals

The goals of product at Sourcegraph are to make the following true:

- The team is working on the most important things (listed in the [product roadmap](../roadmap/index.md)) to achieve our company goals.
- Each teammate has the customer and product context needed (about customer problems, likely future priorities, possible solutions, etc.) to perform their work effectively.
- The product vision and roadmap are communicated well to teammates and everyone outside Sourcegraph.

## Planning

Product planning has 3 parts:

- Long-term/high-level product planning: described here, done [continuously](index.md#product-planning-is-continuous)
- Planning for a release: described here, results in a tracking issue per project
- [Triaging issues](../issues.md)

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
- [Personas](./personas.md)

## Release early, release often

Each project, no matter how long-running, needs to plan to ship *something* in each release. The "something" depends on the project. We strongly prefer for it to be a minimal viable feature that is enabled by default. The next best thing is to ship something that is feature-flagged off by default.

The reason for this is to avoid going for too long without customer feedback (from customers trying it) or even technical/product feedback (from performing the diligent work of polishing it to be ready to release). Lacking these critical checks means we will end up building something that doesn't solve people's problems or that is over-built.

When we have relaxed this in the past, the results have been bad and the overwhelming feedback from retrospectives has been to release regularly.

## Saying "no"

We receive tons of feature request and bug reports, more than we can handle. This means we must frequently say "no" or prioritize things less urgently than some people would like. Our job is to find the most important things to work on.

## Roles

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

### Tech lead

Each project has one tech lead. A tech lead serves in that role for one or more release cycles.

The tech lead is responsible for making sure the following two statements are true:

> The issues for the current release milestone describe the problem/solution.
>
> The issues for the current release milestone are completed (closed) before the release.

The following sections give more details.

#### The issues for each release milestone are completed (closed) before the release.

This means the tech lead needs to do 2 things:

- **Estimate** at planning time what can get done in the time alotted for a release milestone.
- **Reschedule** issues continuously into/from the milestone depending on the pace and [triage](../issues.md#triage).

It's impossible to estimate perfectly, so the tech lead needs to continuously monitor the team's progress and make adjustments accordingly.

If it looks like the team won't be able to complete all planned issues by the release date, the tech lead **must** communicate this in the #product channel and follow up with a combination of:

- helping unblock or accelerate the work of individual teammates
- descoping issues in this milestone
- deferring issues to a later milestone
- getting help from other people on certain issues

The [product manager](#product-manager) can help here, especially with product and priority questions for descoping or deferring work.

#### The issues for each release milestone describe the problem/solution.

This means the tech lead needs to do 2 things:

- Ensure the issues for feature work have enough information for the devs who are implementing the feature.
- Gather information for bug reports (including documenting/automating how other people can supply the info needed to diagnose/fix the bug).
