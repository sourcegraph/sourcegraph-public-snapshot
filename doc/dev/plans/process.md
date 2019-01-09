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
3. The owner opens a work-in-progress PR (e.g. "wip: " prefix) to add a new document to this directory that follows the [template](0000-00-00-template.md).
    - The file name has the following components:
        - The date that the plan was first authored (YYYY-MM-DD).
        - The dash (`-`) separated title.
        - The markdown (`.md`) suffix.
    - It is ok if the first draft of the plan isn't complete. It can be used as a starting point for discussion.
4. The owner does whatever work is necessary to flesh out the details of the plan (e.g. discussions with teammates, experimental coding, etc.).
    - The goal is to complete this step in no longer than a week. If this step takes more than a week, it might be a signal that the project is too big or there is too much uncertainty. The owner should consider descoping the project in this case.
5. When planning document is done, the owner removes the "wip: " prefix and requests approval from relevant stakeholders, including the owner's manager.
5. Once a plan is approved, the owner copies the implementation checklist from the plan into a new issue.
6. The owner updates the plan to include a link to the new issue and then merges the plan.
7. The owner completes the checklist in the issue, updating it as necessary throughout the project.
8. After the project is done, the owner schedules a retrospective with the everyone involved in the project, including the owner's manager, and fills out the retrospective portion of the plan.
    - [How Atlassian does retrospectives](https://www.atlassian.com/team-playbook/plays/retrospective).