# Sourcegraph Cloud

[Sourcegraph Cloud](https://sourcegraph.com/search) lets you search your public code on GitHub or GitLab, and across millions of open-source project.

Note that you can search across a maximum of 2,000 repositories at once using Sourcegraph Cloud. To search across more than 2,000 repositories at once or to search code hosted in an on-prem environment, [run your own Sourcegraph instance](../../../admin/install/index.md).

Note that **Organizations** feature is disabled in Sourcegraph Cloud. If you want to use it, you also need to [run your own Sourcegraph instance](../../../admin/install/index.md).


## Explanations and how-tos

- [Adding repositories to Sourcegraph Cloud](../how-to/adding_repositories_to_cloud.md)
- [Searching across repositories you’ve added to Sourcegraph Cloud with search contexts](../how-to/searching_with_search_contexts.md)
- [Who can see your code on Sourcegraph Cloud](./code_visibility_on_sourcegraph_cloud.md)

## FAQ
### What is Sourcegraph Cloud? 

### What are the differences between Sourcegraph cloud and Sourcegraph on-prem/Enterprise? 

Sourcegraph cloud is fundamentally similar to Sourcegraph on-prem. Both have the same search capabilities that help developers understand big code. Sourcegraph cloud is still under active development and missing several key organizational features. See a [full breakdown between cloud, on-prem, and enterprise](../../cloud/cloud_ent_on-prem_comparison.md). 

### What if I want to use Sourcegraph for my organization? 

Sourcegraph cloud only supports individual users today. This means that any user can sign up for Sourcegraph.com, connect public or private repositories hosted on Github.com or Gitlab.com, and leverage the powerful code search of Sourcegraph. Organizations are supported in Sourcegraph on-prem. Learn how to [run your own Sourcegraph instance](../../../admin/install/index.md).

### What if my code is not stored on Github.com or Gitlab.com? 
Today, only Github.com or Gitlab.com are supported on Sourcegraph cloud, though many other code-hosts are supported in our self-hosted version of Sourcegraph. Learn how to [run your own Sourcegraph instance](../../../admin/install/index.md).

## Search contexts

>NOTE: This feature is still in active development.

Search contexts help you search the code you care about on Sourcegraph cloud. A search context represents a set of repositories on Sourcegraph cloud that will be targeted by search queries by default.

Sourcegraph Cloud supports two search contexts: 

- Your personal context, `context:@username`, which automatically includes [all repositories you add to Sourcegraph](../how-to/adding_repositories_to_cloud.md).
- The global context, `context:global`, which includes all repositories on Sourcegraph cloud.

You can also define your own search contexts to include the subset of repositories you search most often. See more about [search contexts here](../how-to/search_contexts.md)
