# Notifications on Sourcegraph cloud

There are various kinds of notifications to indicate status of your repositories on Sourcegraph cloud:

1. **No repositories:** You have not added any repositories. Try [Adding repositories to Sourcegraph cloud](../code_search/how-to/adding_repositories_to_cloud.md).
1. **Repositories cloning:** Some or all of the repositories you've added to Sourcegraph are currently being cloned. At this stage, search results are computed on-the-fly over your repositories. Note that this may be slow, and you may not get search results across all of the repositories you've added to Sourcegraph.
1. **Repositories indexing:** Some or all of your repositories are currently being indexed. Search results will return across all of your repositories, and more quickly for indexed repositories.
1. **Everything is good:** All of the repositories you've added to Sourcegraph have been cloned and indexed. Search results will be fast across all of your repositories.
1. **Something went wrong:** Something unexpected happened while cloning the repositories you added to Sourcegraph. Please refer to the [Troubleshooting](#Troubleshooting) section for possible solutions.

### Troubleshooting

1. **Sourcegraph.com OAuth application is revoked on the code host:** Check whether the Sourcegraph.com OAuth application on the code host appears in your authorized OAuth applications list. If it does not appear in the list, it may have been revoked. Try repeating [Adding repositories to Sourcegraph cloud](../code_search/how-to/adding_repositories_to_cloud.md).
1. **API rate limit exceeded:** **API rate limit exceeded:** GitHub.com's API rate limit quota is [shared among all OAuth applications authorized by a user, personal access tokens owned by that user, and requests authenticated with that user's username and password](https://docs.github.com/en/developers/apps/building-github-apps/rate-limits-for-github-apps#normal-user-to-server-rate-limits). Sourcegraph cloud consumes very low amount of the API rate limit quota. If you have an unusually heavy workload during a given period of time, the quota may be exceeded. This error should be resolved automatically within the next hour once the GitHub.com API rate limit quota is renewed.
