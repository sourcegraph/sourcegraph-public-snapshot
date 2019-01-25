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

### Process

A release cycle is [1 month](#releases-are-monthly). Each release cycle goes as follows (where T is the start of the release cycle, beginning the moment when the previous release ships):

- Between T-2 weeks and T: for all projects, [product planning](product/index.md#planning) for the release is done; tracking issues and all other foreseen work issues exist.
  - Because [product planning is continuous](product/index.md#product-planning-is-continuous), these may have been ready even earlier for certain projects.
- Between T and T+1 month:
  - Each project team works on "making the tracking issue come true" by closing all project issues in the release milestone.
  - [Triage issues](issues.md#triage) as they come in (which may require updates to the project plan for the current release because they supply us with new information).
  - TODO: document testing and release preparation
- T+1 month:
  - Release! (TODO: document process and also how we do it if the 20th is on a non-work day)
  - [Start a retrospective.](retrospectives/index.md#how-to-lead-a-retrospective)

## Roles

### Tech lead

Each project has one tech lead. A tech lead serves in that role for one or more release cycles.

The tech lead is responsible for making sure the following two statements are true:

> The issues for the current release milestone are completed (closed) by the release.
>
> The issues for the current release milestone describe the problem/solution.

The following sections give more details.

#### The issues for each release milestone are completed (closed) by the release.

This means the tech lead needs to do 2 things:

- **Estimate** at planning time how much can get done in the time alotted for a release milestone.
- **Reschedule** issues continuously into/from the milestone depending on the pace and [triage](issues.md#triage).

It's impossible to estimate perfectly, so the tech lead needs to continuously monitor the team's progress and make adjustments accordingly.

If it looks like the team won't be able to complete the issues by the release, the tech lead **must** mention this in the #product channel and needs to do some combination of:

- helping unblock or accelerate the work of individual teammates
- descoping issues in this milestone
- deferring issues to a later milestone
- getting help from other people on certain issues

The [product manager](product/index.md#product-manager) can help here, especially with product and priority questions for descoping or deferring work.

#### The issues for each release milestone describe the problem/solution.

This means the tech lead needs to do 2 things:

- Ensure the issues for feature work have enough information for the devs who are implementing the feature.
- Gather information for bug reports (including documenting/automating how other people can supply the info needed to diagnose/fix the bug).

## We don't release continuously

Although [Sourcegraph.com](https://sourcegraph.com) is continuously deployed (from sourcegraph/sourcegraph@`master`), the version of Sourcegraph that customers use is not continuously released or updated. This is because:

- We don't think customers would be comfortable with a continuously updated service running on their own infrastructure, for security and stability reasons.
- We haven't built the automated testing and update infrastructure to make continuous customer releases reliable.

In the future, we may introduce continuous releases if these issues become surmountable.
