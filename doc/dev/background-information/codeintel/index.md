# Developing code intelligence

This guide documents our approach to developing code intelligence-related features in our codebase. This includes the code intelligence [subsystems](https://github.com/sourcegraph/sourcegraph/tree/main/enterprise/cmd) and the [extensions](https://github.com/sourcegraph/code-intel-extensions) that provide code intelligence to the web UI, browser extension, and code host integrations.

Services:

- [executor-queue](https://github.com/sourcegraph/sourcegraph/blob/main/enterprise/cmd/executor-queue/README.md)
- [executor](https://github.com/sourcegraph/sourcegraph/blob/main/enterprise/cmd/executor/README.md)
- [precise-code-intel-worker](https://github.com/sourcegraph/sourcegraph/blob/main/enterprise/cmd/precise-code-intel-worker/README.md)

Docs:

- [Deployment documentation](deployment.md)
- [How LSIF indexes are processed](uploads.md)
- [How precise code intelligence queries are resolved](queries.md)
- [How code intelligence extensions resolve hovers](extensions.md)
- [How Sourcegraph auto-indexes source code](auto-indexing.md)

