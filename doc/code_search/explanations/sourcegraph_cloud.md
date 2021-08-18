# Sourcegraph cloud

[Sourcegraph cloud](https://sourcegraph.com/search) lets you search across your code from GitHub.com or GitLab.com, and across any open-source project on GitHub.com or Gitlab.com.

## Explanations and how-tos

- [Adding repositories to Sourcegraph Cloud](../how-to/adding_repositories_to_cloud.md)
- [Who can see your code on Sourcegraph Cloud](./code_visibility_on_sourcegraph_cloud.md)
- [Searching across repositories you’ve added to Sourcegraph Cloud with search contexts](../how-to/searching_with_search_contexts.md)

## Limitations

- **Adding repositories**: You can add a maximum of 2,000 repositories to Sourcegraph cloud. To add more than 2,000 repositories or to search code hosted in environments other than GitHub.com or GitLab.com, [run your own Sourcegraph instance](../../../admin/install/index.md).
- **Searching code**: You can search across a maximum of 50 repositories at once with a `type:diff` or `type:commit` search using Sourcegraph cloud. To search across more than 50 repositories at once, [run your own Sourcegraph instance](../../../admin/install/index.md).
- **Organizations and collaboration**: Sourcegraph Cloud currently only supports individual use of Sourcegraph Cloud. To create and manage an organization with Sourcegraph with team-oriented functionality, get started with the [self-hosted deployment](../../../admin/install/index.md) in less than a minute.

## FAQ

### What are the differences between Sourcegraph cloud and Sourcegraph self-hosted / Enterprise?

Both Sourcegraph cloud and self-hosted Sourcegraph instances power the same search experience relied on by developers around the world. The Sourcegraph team is working on bringing Sourcegraph cloud to feature parity with our self-hosted Sourcegraph solution. See a [full breakdown between Sourcegraph cloud, self-hosted, and enterprise](../../cloud/cloud_self_hosted_comparison.md).

### How do I use Sourcegraph cloud for my organization?

Sourcegraph cloud only supports individual use today. This means that anyone can sign up for Sourcegraph.com, connect public or private repositories hosted on Github.com or Gitlab.com, and leverage the powerful code search of Sourcegraph. To create and manage an organization with Sourcegraph with team-oriented functionality, get started with the [self-hosted deployment](../../../admin/install/index.md).

### What if my code is not hosted on Github.com or Gitlab.com?

Today, only Github.com or Gitlab.com are supported on Sourcegraph cloud. To search your code hosted on other code hosts, get started with the [self-hosted version of Sourcegraph](../../../admin/install/index.md).
