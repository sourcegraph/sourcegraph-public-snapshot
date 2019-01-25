# Issues

We use issues (on GitHub) to track all of our work.

## Principles

- Each issue is something that needs to be worked on: a bug, feature, doc page, blog post, question, etc.
- For any given piece of work, the issue is the source of truth for the description and status of the work. If anyone asks a question like "What's the status of XYZ?" or "What are we doing for XYZ?", anyone else should be able to answer the question with a link to the issue for XYZ.
- The assignee of an issue is responsible for the issue. This includes keeping it up to date (by updating the description, labels, and milestone as needed) and doing the work necessary to close it.

## Triage

Triage is how we handle issues filed in our repositories. For example, if a user reports an unforeseen bug, we often want to fix it in the current release milestone instead of waiting to review it in the next [product planning meeting](product/index.md#meetings) for the following release. This may require delaying other work to make time for the fix.

We use an "eventually consistent" triage process to ensure newly filed issues are handled.

### Primary triage

The **primary triage process** is how most issues will be triaged:

1. Assignment. Anyone who sees an unassigned issue should assign it if they know the right assignee.
1. Prioritization. The assignee adds a milestone to the issue to indicate when it'll be closed, one of:
   - *The current release milestone:* if the assignee commits to closing it for the upcoming release.
   - *A future release milestone:* if the assignee, consulting the [roadmap](roadmap/index.md) (and the [product manager](product/index.md#product-manager) if necessary), thinks it's likely that it will/should be prioritized for that future release. The issue will be reviewed again in [product planning](product/index.md#planning) as a check.
   - Backlog: for all other issues.
1. Details: The assignee is responsible for obtaining the information necessary for them to fix the issue and updating the first issue comment to reflect/summarize the current state so readers don't have to scan the entire conversation of the issue to get caught up.
   - For issues in the current release milestone, each project's [tech lead](releases.md#tech-lead) is ultimately responsible for ensuring that these details are present in their project's issues.
   - Skip this step (and avoid wasting time) for backlog issues in most cases.
1. Labels (informal). Anyone can label an issue. The assignee is responsible for the issue having the right labels.

If you see an issue you can close quickly (e.g., within 5 minutes) and want to handle, just self-assign it and don't bother following these steps.

### Secondary triage

The **secondary triage process** (run by the [product manager](product/index.md#product-manager)) ensures nothing falls through the cracks:

- The PM reviews new unassigned issues a few times daily and assigns them.
- The PM reviews issues with no milestone or a recently changed milestone and ensures they have the correct milestone.

The PM is responsible for all issues being triaged within 24 hours (whether by the primary or secondary process). In general, we think it's good to see more issues being correctly handled by the primary triage process (vs. the secondary process) because that implies that devs have sufficient context and are closer to the customer.

## Multiple repositories

We use issues across multiple repositories (not just [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph)).

- Use the same milestone names for all repositories in the `sourcegraph` GitHub organization: `Backlog` and one milestone per release (such as `3.1`). This lets us filter issues by milestone across all repositories.
- Labels are different in each repository. If you think it'd help to have a standard set of labels in all repositories, update this document.

### Finding issues across all repositories

- Use a global issue search to find issues in all `sourcegraph`-organization repositories: [is:open is:issue user:sourcegraph](https://github.com/issues?page=3&q=is%3Aopen+is%3Aissue+milestone%3A3.0+user%3Asourcegraph).
- Add `assignee:$USER` to the query to find issues assigned to you (replace `$USER` with your GitHub username).
- Add `milestone:$MILESTONE` (e.g., `milestone:3.1`) to the query to filter by milestone.
- Monitor your [GitHub notifications](https://github.com/notifications) if you find that view useful.
