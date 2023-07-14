# Embeddings Policies

Embedding policies define which repositories are automatically scheduled for embedding.

## How to create an embeddings policy

Policies are created by administrators from the _Site Admin_ page.
Open the _Site Admin_ page and select **Cody > Embedding Policies** from the left-hand navigation menu.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/embeddings/embedding-policies.png" class="screenshot" alt="Embedding policies page">

The page shows a list of all existing embedding policies.
Click the **Create new global policy** button to create a new policy.
Provide a descriptive name for the policy and click on **Add repository pattern** to define a pattern.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/embeddings/new-policy-create.png" class="screenshot" alt="Create new global policy">

The pattern matches exactly, unless you use an asterisk _*_, which matches any sequence of zero or more characters.

### Example 1

- Original: "github.com/sourcegraph/sourcegraph"
- Pattern: `github.com/sourcegraph/*`
- Result: The pattern matches the original and all other repositories under the "github.com/sourcegraph/" organization.

### Example 2

- Original: "github.com/sourcegraph/sourcegraph"
- Pattern: `github.com/sourcegraph/sourcegraph`
- Result: The pattern matches only the original.

### Example 3

- Original: "github.com/sourcegraph/sourcegraph"
- Pattern: `*sourcegraph*`
- Result: The pattern matches the original and any repository, from any code host, with the word "sourcegraph" in it.

The policy will be applied to all repositories that match the pattern.
If you choose not to define a pattern, the policy will be applied to up to [embeddings.policyRepositoryMatchLimit](./code_graph_context.md#configuring-the-global-policy-match-limit) repositories.

We recommend embedding only repositories that are actively developed or highly relied upon,
as embedding all repositories without discrimination can consume significant resources without necessarily providing better context for Cody.

Finally, click on **Create policy**.
The new policy will be shown in the list of policies.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/embeddings/new-policy-saved.png" class="screenshot" alt="policy-saved">

## Lifecycle of an embeddings policy

Once a policy has been created, it is active.
To deactivate a policy, simply delete it.
A worker process periodically checks the embeddings policies and resolves them into a list of repositories to index.
Another worker then creates a new index job for each repository and queues it for processing.
A repository cannot be queued for processing if

- it is already queued or being processed
- a job for the same repository and the same revision already completed successfully
- if another job for the same repository has been queued for processing within the [embeddings.MinimumInterval](./code_graph_context.md#adjust-the-minimum-time-interval-between-automatically-scheduled-embeddings) time window

All embeddings jobs are listed under **Cody > Embeddings Jobs**

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/embeddings/embeddings-jobs.png" class="screenshot" alt="policy-saved">
