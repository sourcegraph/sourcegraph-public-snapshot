# Code host connections

Sourcegraph can sync repositories from code hosts and other similar services.

**Site admins** can configure the following code hosts:

- [GitHub](github.md)
- [GitLab](gitlab.md)
- [Bitbucket Cloud](bitbucket_cloud.md)
- [Bitbucket Server](bitbucket_server.md)
- [Phabricator](phabricator.md)
- [Gitolite](gitolite.md)
- [AWS CodeCommit](aws_codecommit.md)
- [Other Git code hosts (using a Git URL)](other.md)
- [Non-Git code hosts](non-git.md)
  - [Perforce](../repo/perforce.md)

**Users** can configure the following public code hosts:

- [GitHub](github.md)
- [GitLab](gitlab.md)
- [Bitbucket Cloud](bitbucket_cloud.md)

The feature is currently in beta and can be enabled by adding the `AllowUserExternalServicePublic` tag to the user in the Sourcegraph database. An example query to enable this for the user with username `bob` is:

```sql
UPDATE users SET tags = array_append(tags, 'AllowUserExternalServicePublic') where username='bob';
```
