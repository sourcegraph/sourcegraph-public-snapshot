# Integrating your code host with Sourcegraph (WIP)

Sourcegraph Universal Code Search means we integrate with every major Git code host with flexible hosting options for generic and non-Git code hosts. It also means Sourcegraph can index and search across repositories from different code hosts, e.g., GitHub.com and GitLab CE, on the same instance.

Sourcegraph can also search across every branch, every commit diff, and even commit messages for every code host. This is why regardless of how many repositories developers may sync to their local machines, Sourcegraph's search capabilities outweigh those of any editor, IDE, and existing code search tool. See our [detailed feature comparison chart](https://about.sourcegraph.com/workflow) for more details.

## Code host integration beyond repository search

Code host integration may also include:

- **User authentication**<br/>
Log-in via your code host (requires oAuth application or similar)<br/><br/>
- **Repository permission syncing**<br/>
Only display search results for repositories you have access to (requires code host authentication).<br/><br/>
- **Code intelligence**<br/>
Decorate code views with IDE quality code intelligence, either through a native application/integration (BitBucket Server, GitLab, Phabricator) or browser extension (GitHub)

## Code host integrations

- [GitHub](github.md)
- [GitLab](gitlab.md)
- [Bitbucket Cloud](integration/bitbucket_cloud.md)
- [Bitbucket Server](integration/bitbucket_server.md)
- [Phabricator](integration/phabricator.md)
- [AWS CodeCommit](integration/aws_codecommit.md)
- [Gitolite](integration/gitolite.md.md)

---

[**» Next: Universal Code Search**](universal_code_search.md)
