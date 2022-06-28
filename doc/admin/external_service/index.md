# Code host connections

Sourcegraph can sync repositories from code hosts and other similar services.

**Site admins** can configure the following code hosts:

- [GitHub](github.md)
- [GitLab](gitlab.md)
- [Bitbucket Cloud](bitbucket_cloud.md)
- [Bitbucket Server / Bitbucket Data Center](bitbucket_server.md) or Bitbucket Data Center
- [Phabricator](phabricator.md)
- [Gitolite](gitolite.md)
- [AWS CodeCommit](aws_codecommit.md)
- [Other Git code hosts (using a Git URL)](other.md)
- [Non-Git code hosts](non-git.md)
  - [Perforce](../repo/perforce.md)

**Users** can configure the following public code hosts:

- [GitHub.com](github.md)
- [GitLab.com](gitlab.md)


## Rate limits
Sourcegraph makes our best effort to use the least amount of calls to your code host. However, it is possible for Sourcegraph 
to encounter rate limits in some scenarios. Please see the specific code host documentation for more information and how to 
mitigate these issues. 

### Increasing code host rate limits
Customers should avoid creating additional **free** accounts for the purpose of circumventing code-host rate limits. 
Some code hosts have higher rate limits for **paid** accounts and allow the creation of additional **paid** accounts which 
Sourcegraph can leverage.

Please contact support@sourcegraph.com if you encounter rate limits.
