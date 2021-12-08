# Who can see your organization’s code on Sourcegraph Cloud
Note: Team access for Sourcegraph Cloud is in Private Beta. If you are interested in getting early-access, [join our waitlist](https://share.hsforms.com/14OQ3RoPpQTOXvZlUpgx6-A1n7ku)!

Sourcegraph Cloud protects your organization’s private code using repository permissions from GitHub.com and GitLab.com to determine who can see repositories you’ve added to your organization on Sourcegraph Cloud.

## Public repositories

If a repository is public on GitHub.com or GitLab.com, other users on Sourcegraph Cloud can view and search across that repository. The repository will appear in the global search context.

## Private repositories

If a repository your organization has added to Sourcegraph Cloud is private on GitHub.com or GitLab.com, only users who have permission to access that repository on the code host and are members of your organization on Sourcegraph Cloud can view and search that repository. The repository will not appear in search results for other users.

Note that if a user has personally added a private repository to Sourcegraph Cloud, they may see that private repository even if they are not a member of your organization, as long as they have permission to access to the repository on the code host. 

The Sourcegraph team and administrators on Sourcegraph Cloud cannot view private repositories. Metadata related to private repositories on Sourcegraph cloud is excluded from all analytics and plain-text data storage. You can read more in our privacy policy.
