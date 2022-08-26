# Code host connections

Sourcegraph can sync repositories from code hosts and other similar services.

**Site admins** can configure the following code hosts:

- [GitHub](github.md)
- [GitLab](gitlab.md)
- [Bitbucket Cloud](bitbucket_cloud.md)
- [Bitbucket Server / Bitbucket Data Center](bitbucket_server.md) or Bitbucket Data Center
<!-- [Phabricator](phabricator.md) -->
<!-- [Gitolite](gitolite.md) -->
<!-- [AWS CodeCommit](aws_codecommit.md) -->
- [Other Git code hosts (using a Git URL)](other.md)
- [Non-Git code hosts](non-git.md)
  - [Perforce](../repo/perforce.md)
  - [JVM dependencies](jvm.md)
  - [Go dependencies](go.md)
  - [npm dependencies](npm.md)
  - [Python dependencies](python.md)

**Users** can configure the following public code hosts:

- [GitHub.com](github.md)
- [GitLab.com](gitlab.md)


## Rate limits
Sourcegraph makes our best effort to use the least amount of calls to your code host. However, it is possible for Sourcegraph 
to encounter rate limits in some scenarios. Please see the specific code host documentation for more information and how to 
mitigate these issues. 

### Rate limit syncing
Sourcegraph has a mechanism of syncing code host rate limits. When Sourcegraph is started, code host configurations of all
external services are checked for rate limits and these rate limits are stored and used.

When any of code host configurations is edited, rate limits are synchronized and updated if needed, this way Sourcegraph always 
knows how many requests to which code host can be sent at a given point of time.

### Current rate limit settings
Current rate limit settings can be viewed by site admins on the following page: `Site Admin -> Instrumentation -> Repo Updater -> Rate Limiter State`.
This page includes rate limit settings for all external services configured in Sourcegraph. 

Here is an example of one external service, including information about external service name,  maximum allowed burst of requests,
maximum allowed requests per second and whether the limiter is infinite (there is no rate limiting):
```json
{
  "extsvc:github:4": {
    "Burst": 10,
    "Limit": 1.3888888888888888,
    "Infinite": false
  }
}
```

### Increasing code host rate limits
Customers should avoid creating additional **free** accounts for the purpose of circumventing code-host rate limits. 
Some code hosts have higher rate limits for **paid** accounts and allow the creation of additional **paid** accounts which 
Sourcegraph can leverage.

Please contact support@sourcegraph.com if you encounter rate limits.
