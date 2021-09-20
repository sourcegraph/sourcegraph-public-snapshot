# Configure precise code intelligence data retention policies

This guide shows how to configure the retention policies for precise code intelligence data. As code intelligence data ages, it gets less and less relevant. Users tend to need code intelligence on newer commits, tagged commits, and in branches. The data providing intelligence for an old commit on the main development branch is very unlikely to be used, but takes up valuable space in the database.

Each policy has a number of configurable options, including:

- The set of Git commits, branches, or tags to which the policy applies
- How long to keep the associated code intelligence data
- For branches, whether or not to consider the _tip_ of the branch only, or all commits contained in that branch

Note that we also track cross-repository dependencies and will not delete any data that is referenced by another precise code intelligence index. This ensures that we don't delete code intelligence for dependencies pinned to older versions (or dependencies that have reached a steady state and no longer receives frequent updates).

All upload records will be periodically compared against global and their target repository's data retention policies with the exception of uploads that provide code intelligence for the tip of the default branch. These uploads will never expire due to age.

<style>
img.screenshot {
  display: block;
  margin: 1em auto;
  max-width: 600px;
  margin-bottom: 0.5em;
  border: 1px solid lightgrey;
  border-radius: 10px;
}

img.screenshot.thin-screenshot {
  max-width: 200px;
}
</style>

## Applying data retention policies globally

Data retention policies can be applied to _all repositories_ on your Sourcegraph instance. In order to view and edit these policies, navigate to the code intelligence configuration in the site-admin dashboard.

<img src="https://sourcegraphstatic.com/docs/images/code-intelligence/retention-config-sidebar.png" class="screenshot thin-screenshot">

By default, there are two _protected_ policies which cannot be deleted or disabled. These policies refer to all tagged commits (associated data being kept for one year by default), and the tip of all branches (associated data being kept for three months by default). 

<img src="https://sourcegraphstatic.com/docs/images/code-intelligence/retention-config-global-list.png" class="screenshot">

The retention length for protected policies can be modified to suit your instance's usage patterns.

<img src="https://sourcegraphstatic.com/docs/images/code-intelligence/retention-config-global-detail.png" class="screenshot">

New policies can also be created to apply to any arbitrary Git commit, branch, or tag patterns. For example, you may want to **never** expire any data on the tip of any `main` or `master` branch in your organization.

## Applying data retention policies to a specific repository

Data retention policies can also be created per-repository basis as commit and merge workflows differ wildly from project to project. In order to view and edit repository-specific policies, navigate to the code intelligence settings in the target repository's index page.

<img src="https://sourcegraphstatic.com/docs/images/code-intelligence/retention-config-repo.png" class="screenshot">

Global policies continue to apply to repositories that define additional policies.

<img src="https://sourcegraphstatic.com/docs/images/code-intelligence/retention-config-repo-list.png" class="screenshot">

In this example, we create a policy that ensures all commits visible to the tip of any branch matching the pattern `ef/*` will not be removed for at least one year.

<img src="https://sourcegraphstatic.com/docs/images/code-intelligence/retention-config-repo-detail.png" class="screenshot">
