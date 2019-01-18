# Releases

This document describes how we plan and ship releases at Sourcegraph.

## Releases are monthly

We ship a release on the 20th day of each month. (Except in February 2019, where we will ship on the 4th and 20th.) Why the 20th? Because it's a day number that we can hit each month (any other number would result in a December release that comes out too close to Christmas, or a January release that comes out too soon after New Year's Day).

The release always ships on time, even if it's missing features or bug fixes we hoped to get in ([why?](https://about.gitlab.com/2015/12/07/why-we-shift-objectives-and-not-release-dates-at-gitlab/)).

## Milestones and issues

Each release has a GitHub milestone ([in all repositories](issues.md#multiple-repositories)) whose name is the version number, such as `3.1`:

- A release milestone's [issues](issues.md) constitute the release plan (i.e., the canonical definition of the work for the release).
- A release is ready when its milestone has zero open issues ([across all repositories](issues.md#multiple-repositories)).

The issues for a release come from two sources:

- [product planning](product/index.md#planning) (work planned for the release ahead of time)
- [issue triage](issues.md#triage) (issues filed at any time)

## Planning

The release planning process is how we prepare the issues in a release milestone so that the release delivers on our [roadmap](roadmap/index.md). It occurs before all other work on the release, around the time when the previous release ships.

The goal of the release planning process is to make the following true:

> The issues for the current release:
>
> - are the most important things to work on
> - are completed (closed) by the release
> - describe the problem/solution

### Process

Planning is an iterative process because it surfaces new information that causes us to reconsider the decisions we made at previous steps (which is the point!). The process is:

1. Repeat until convergence to the best of our abilities given current information:
   1. Ensure the [roadmap](roadmap/index.md) is prioritized to best achieve our company goals and is achievable.
   1. For each project in this release on the roadmap, follow the [per-project planning process](#per-project-planning-process).
1. While release work is underway: [triage issues](issues.md#triage), updating the release plan if necessary.
1. When all release milestone issues are closed: release! <!-- TODO(sqs): Document release process. -->
1. Immediately after release: [start a retrospective](retrospectives/index.md#how-to-lead-a-retrospective).


### Per-project planning process

Each project in this release on the [roadmap](roadmap/index.md) must be planned. This involves the the [product manager](#product-manager), engineering manager, and the team responsible for the project.

For each project:

1. Define the team of people responsible for the project.
1. TODO(sqs): examine how gitlab kicks off projects with input from product

- the right team of people owns it;
- it has a good [tracking issue](plans/0000-00-00-template.md)
      - 

## Roles


### Tech lead

The tech lead is responsible for making sure the following two statements are true:

> The issues for each release milestone are completed (closed) by the release.
>
> The issues for each release milestone describe the problem/solution.

The following sections give more details.

#### The issues for each release milestone are completed (closed) by the release.

This means the tech lead needs to do 2 things:

- **Estimate** at planning time how much can get done in the time alotted for a release milestone.
- **Reschedule** issues continuously into/from the milestone depending on the pace.

It's impossible to estimate perfectly, so the tech lead needs to continuously monitor the team's progress and make adjustments accordingly.

If it looks like the team won't be able to complete the issues by the release, the tech lead needs to do some combination of:

- helping unblock or accelerate the work of individual teammates
- descoping issues in this milestone
- deferring issues to a later milestone
- getting help from other people on certain issues

The product manager can help here, especially with product and priority questions for descoping or deferring work.

#### The issues for each release milestone describe the problem/solution.

This means the tech lead needs to do 2 things:

<!-- TODO rough from here on -->

- Ensure the issues for feature work have enough information for the devs who are implementing the feature.
- Gather information for bug reports (including documenting/automating how other people can supply the info needed to diagnose/fix the bug).

## We don't release continuously

Although [Sourcegraph.com](https://sourcegraph.com) is continuously deployed (from sourcegraph/sourcegraph@`master`), the version of Sourcegraph that customers use is not continuously released or updated. This is because:

- We don't think customers would be comfortable with a continuously updated service running on their own infrastructure, for security and stability reasons.
- We haven't built the automated testing and update infrastructure to make continuous customer releases reliable.

In the future, we may introduce continuous releases if these issues become surmountable.
