# Connecting your code host

Sourcegraph supports [every popular Git code host](#code-host-integrations) with flexible hosting options for [other git repository hosts](../admin/external_service/other.md) and [non-Git code hosts](../admin/external_service/non-git.md). Sourcegraph also  index and search across repositories from different code hosts, e.g., GitHub.com and GitLab CE, on a single instance.

Sourcegraph can also search across every branch, every commit diff, and even commit messages for every code host. This is why regardless of how many repositories developers may sync to their local machines, Sourcegraph's search capabilities outweigh those of any editor, IDE, and existing code search tool. See our [detailed feature comparison chart](https://about.sourcegraph.com/workflow) for more details.

## Code host integration beyond repository search

Code host integration may also include:

- **User authentication**<br/>
Log-in via your code host (requires oAuth application or similar)<br/>

- **Repository permission syncing**<br/>
Only display search results for repositories you have access to (requires code host authentication).<br/>

- **Code intelligence**<br/>
Decorate code views with IDE quality code intelligence, either through a native application/integration (BitBucket Server, GitLab, Phabricator) or browser extension (GitHub)

## Code host integrations

Select your code host below for configuration instructions:

- [GitHub](github.md)
- [GitLab](gitlab.md)
- [Bitbucket Cloud](../integration/bitbucket_cloud.md)
- [Bitbucket Server](../integration/bitbucket_server.md)
- [Phabricator](../integration/phabricator.md)
- [AWS CodeCommit](../integration/aws_codecommit.md)
- [Gitolite](../integration/gitolite.md)

---

[**» Next: Introduction to Universal Code Search**](intro_universal_code_search.md)
