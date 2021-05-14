# Sourcegraph Cloud

[Sourcegraph Cloud](https://sourcegraph.com/search) lets you search inside your public code on GitHub or GitLab, and inside any open-source project on GitHub.

Note that you can search across a maximum of 50 repositories at once using Sourcegraph Cloud. To search across more than 50 repositories at once or to search your organization's internal code, [run your own Sourcegraph instance](../../../admin/install/index.md).

## Search contexts <span class="badge badge-primary">experimental</span>

>NOTE: This feature is still in active development.

Search contexts help you search the code you care about on Sourcegraph Cloud. A search context represents a set of repositories on Sourcegraph Cloud that will be targeted by search queries by default.

Sourcegraph Cloud supports two search contexts: 

- Your personal context, `context:@username`, which automatically includes [all repositories you add to Sourcegraph](../how-to/adding_repositories_to_cloud.md).
- The global context, `context:global`, which includes all repositories on Sourcegraph Cloud.

Coming soon: create your own search contexts that include the repositories you choose. Want early access to custom search contexts? [Let us know](mailto:feedback@sourcegraph.com).
