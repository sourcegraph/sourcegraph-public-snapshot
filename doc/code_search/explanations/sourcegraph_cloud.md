# Sourcegraph Cloud

[Sourcegraph Cloud](https://sourcegraph.com/search) lets you search your public code on GitHub or GitLab, and across millions of open-source project.

Note that you can search across a maximum of 2,000 repositories at once using Sourcegraph Cloud. To search across more than 2,000 repositories at once or to search code hosted in an on-prem enviornment, [run your own Sourcegraph instance](../../../admin/install/index.md).

## Search contexts

>NOTE: This feature is still in active development.

Search contexts help you search the code you care about on Sourcegraph Cloud. A search context represents a set of repositories on Sourcegraph Cloud that will be targeted by search queries by default.

Sourcegraph Cloud supports two search contexts: 

- Your personal context, `context:@username`, which automatically includes [all repositories you add to Sourcegraph](../how-to/adding_repositories_to_cloud.md).
- The global context, `context:global`, which includes all repositories on Sourcegraph Cloud.

You can also define your own search contexts to include the subset of repositories you search most often. See more about [search contexts here](../how-to/search_contexts.md)
