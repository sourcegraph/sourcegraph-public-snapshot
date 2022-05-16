# Who can see your organization’s code on Sourcegraph Cloud

Sourcegraph Cloud protects your organization’s private code using repository permissions from GitHub.com and GitLab.com to determine who can see repositories you’ve added to your organization on Sourcegraph Cloud.

## Public repositories

If a repository is public on GitHub.com or GitLab.com, other users on Sourcegraph Cloud can view and search across that repository. The repository will appear in the global search context. If a public repository is made private on the code host, that repository will no longer appear in the global search context. 

## Private repositories

TL;DR: Users on Sourcegraph can only see a private repository if they have permissions to see that repository on your code host and if they are a member of your organization on Sourcegraph. 

If a repository your organization has added to Sourcegraph Cloud is private on GitHub.com or GitLab.com, only users who have permission to access that repository on the code host and are members of your organization on Sourcegraph Cloud can view and search that repository. The repository will not appear in search results for other users.

Note that if a user has personally added a private repository to Sourcegraph Cloud, they may see that private repository even if they are not a member of your organization, as long as they have permission to access to the repository on the code host. 

The Sourcegraph team and administrators on Sourcegraph Cloud cannot view private repositories. You can read more in our privacy policy.
