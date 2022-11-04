# Configure code graph data retention policies

<style>
img.screenshot {
  display: block;
  margin: 1em auto;
  max-width: 600px;
  margin-bottom: 0.5em;
  border: 1px solid lightgrey;
  border-radius: 10px;
}
</style>

This guide shows how to configure the retention policies for code graph data. As code graph data ages, it gets less and less relevant. Users tend to need code navigation on newer commits, tagged commits, and in branches. The data providing intelligence for an old commit on the main development branch is very unlikely to be used, but takes up valuable space in the database.

Each policy has a number of configurable options, including:

- The set of Git branches or tags to which the policy applies
- How long to keep the associated code graph data
- For branches, whether or not to consider the _tip_ of the branch only, or all commits contained in that branch

Note that we also track cross-repository dependencies and will not delete any data that is referenced by another code graph data index. This ensures that we don't delete code graph data for dependencies pinned to older versions (or dependencies that have reached a steady state and no longer receives frequent updates).

All upload records will be periodically compared against global data retention policies as well as their target repository's data retention policies. Uploads on the tip of the default branch for a repository will never expire regardless of age.

## Applying data retention policies globally

Site admins may create data retention policies that are applied to _all repositories_ on your Sourcegraph instance. In order to view and edit these policies, navigate to the code graph configuration in the site-admin dashboard.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/rename/global-list.png" class="screenshot" alt="Global data retention policy configuration list page">

By default, there are three _protected_ policies which cannot be deleted or disabled. These policies refer to all tagged commits (associated data being kept for one year by default), the tip of all branches (associated data being kept for three months by default), and the HEAD of the default branch (associated data being kept forever by default) for all repositories. Protected policies cannot be deleted or disabled, but the retention length can be modified to suit your instance's usage patterns.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/rename/global-protected.png" class="screenshot" alt="Protected global data retention policy edit page">

New policies can also be created to apply to the HEAD of the default branch, or to apply to any arbitrary Git branch or tag pattern. For example, you may want to never expire any data for any major version tags in your organization.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/rename/retention-create.png" class="screenshot" alt="Global data retention policy configuration edit page">
<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/rename/retention-post-create.png" class="screenshot" alt="Global data retention policy configuration created confirmation">

New policies can be created to apply to a set of repositories that are matched by name. For example, you may want to change duration retention on branches that exist within a particular set of repositories (in this example, repositories in the same organization matching `scip-*`).

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/rename/retention-create-repo-list.png" class="screenshot" alt="Global data retention policy with repository patterns configuration edit page">
<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/rename/retention-post-create-repo-list.png" class="screenshot" alt="Global data retention policy with repository patterns configuration created confirmation">

## Applying data retention policies to a specific repository

Data retention policies can also be created on a per-repository basis as commit and merge workflows differ wildly from project to project. In order to view and edit repository-specific policies, navigate to the code graph settings in the target repository's index page.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/sg-3.33/repository-page.png" class="screenshot" alt="Repository index page">

The settings page will show all policies that apply to the given repository, including both repository-specific policies as well as global policies that match the repository.

In this example, we create the _`:bestcoder:` branch retention policy_ that ensures all commits visible to the tip of any branch matching the pattern `ef/*` will not be removed regardless of age.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/rename/retention-repo-create.png" class="screenshot" alt="Repository-specific data retention policy configuration edit page">
<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/rename/retention-repo-post-create.png" class="screenshot" alt="Repository-specific data retention policy configuration created confirmation">
