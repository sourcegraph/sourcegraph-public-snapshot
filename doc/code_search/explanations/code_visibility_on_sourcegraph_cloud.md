# Who can see your code on Sourcegraph Cloud

Sourcegraph Cloud protects your private code using repository permissions from GitHub.com and GitLab.com to determine who can see repositories you've [added to Sourcegraph Cloud](../how-to/adding_repositories_to_cloud.md).

## Public repositories
If a repository is public on GitHub.com or GitLab.com, other users on Sourcegraph Cloud can view and search across that repository. The repository will appear in the global search context.

## Private repositories (private beta)
If a repository is private on GitHub or GitLab, only users who have permission to access that repository on the code host **and** have added that repository to Sourcegraph Cloud can view and search that repository. The repository will not appear in search results for other users.

The Sourcegraph team and administrators on Sourcegraph Cloud cannot view private repositories. Metadata related to private repositories on Sourcegraph Cloud is excluded from all analytics and plain-text data storage.

> NOTE: Want to collaborate on private code with your team or organization on Sourcegraph Cloud? [Sign up for early access](https://share.hsforms.com/1copeCYh-R8uVYGCpq3s4nw1n7ku)!
