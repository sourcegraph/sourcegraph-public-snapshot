# Notifications on Sourcegraph cloud

There are various kinds of notifications to indicate status of your repositories on Sourcegraph cloud:

1. **No repositories:** You have not added any repositories. Try [Adding repositories to Sourcegraph cloud](../code_search/how-to/adding_repositories_to_cloud.md).
1. **Repositories cloning:** Some or all of the repositories you've added to Sourcegraph are currently being cloned. At this stage, search results are computed on-the-fly over your repositories. Note that this may be slow, and you may not get search results across all of the repositories you've added to Sourcegraph.
1. **Repositories indexing:** Some or all of your repositories are currently being indexed. At this stage, search results are coming back much faster for your repositories.
1. **Everything is good:** All of the repositories you've added to Sourcegraph have been cloned and indexed. Search results are fast across all of your repositories.
1. **Something went wrong:** Something unexpected happened during repositories cloning stage, please refer to the [Troubleshooting](#Troubleshooting) section for possible solutions.

### Troubleshooting

1. **Sourcegraph.com OAuth application is revoked on the code host:** You may have accidentally revoked the Sourcegraph.com OAuth application on the code host, if you do not see the Sourcegraph.com OAuth application appear in your authorized OAuth applications list, try repeating [Adding repositories to Sourcegraph cloud](../code_search/how-to/adding_repositories_to_cloud.md).
1. **API rate limit exceeded:** The API rate limit quota of GitHub.com [shares among all OAuth applications authorized by that user, personal access tokens owned by that user, and requests authenticated with that user's username and password](https://docs.github.com/en/developers/apps/building-github-apps/rate-limits-for-github-apps#normal-user-to-server-rate-limits). Sourcegraph cloud consumes very low usage of API rate limit quota, this may be expected when you're having heavy workloads for the hour, and should be resolved automatically once the API rate limit quota is refilled in the next hour for GitHub.com.
