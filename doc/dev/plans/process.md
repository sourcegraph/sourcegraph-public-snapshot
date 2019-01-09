# Planning process

This document defines the process for planning engineering projects.

## Definition

A plan
- Is written down.
- Solves a problem that is a priority.
- Has an owner.
- Has a concrete set of steps in the form of a checklist.
- Has clear success metrics.
- Has a done date that the owner has committed to.

## Goal

The goal of the planning process is to produce plans that accurately reflect what the engineering team is working on.

## Benefits

- Plans can be reviewed asynchronously by our distributed team.
- Plans serve as documentation for why decisions were made.
- Plans can be referenced by teammates and customers.
- Changes to plans are transparent.
- Teammates have a record of the work that they have accomplished.

## Process

1. Product or engineering identifies a problem that engineering needs to solve.
2. The problem is discussed with relevant teammates to determine if the problem is a priority to solve and to identify an owner.
3. If the problem is a priority and an owner is identified, the owner opens a PR to add a new document to this directory that follows the [template](0000-00-00-template.md). The file name has the following components:
    - The date that the plan was first authored (YYYY-MM-DD).
    - The dash (`-`) separated title.
    - The markdown (`.md`) suffix.
4. The owner requests a review from relevant stakeholders, including the owner's manager.
    - Reviewers are expected to respond to new and updated proposals within 2 business days.
    - If a proposal is not approved within 5 business days, the owner's manager will make a final judgement call whether a project is approved as written or rejected.
5. If a proposal is approved, the owner copies the implementation checklist into a new issue that links to the proposal's PR and links to this issue from the proposal. Further changes to the checklist are reflected in the issue and not the proposal.
6. The owner merges the proposal.
7. The owner completes the checklist in the issue, updating it as necessary throughout the project.
8. After the project is done, the owner fills out the retrospective portion of the plan.
