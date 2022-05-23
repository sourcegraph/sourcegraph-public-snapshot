# Developing code intelligence

This guide documents our approach to developing code intelligence-related features in our codebase. This includes the code intelligence logic included in the Sourcegraph instance as well as the [extensions](https://github.com/sourcegraph/code-intel-extensions) that provide code intelligence to the web UI, browser extension, and code host integrations.

Services:

- [precise-code-intel-worker](https://github.com/sourcegraph/sourcegraph/blob/main/enterprise/cmd/precise-code-intel-worker/README.md)

Code intelligence-specific code:

- [lib/codeintel](https://github.com/sourcegraph/sourcegraph/tree/main/lib/codeintel)
- [dev/codeintel-qa](https://github.com/sourcegraph/sourcegraph/tree/main/dev/codeintel-qa)
- [enterprise/internal/codeintel](https://github.com/sourcegraph/sourcegraph/tree/main/enterprise/internal/codeintel)
- [enterprise/cmd/worker/internal/codeintel](https://github.com/sourcegraph/sourcegraph/tree/main/enterprise/cmd/worker/internal/codeintel)
- [enterprise/cmd/frontend/internal/codeintel](https://github.com/sourcegraph/sourcegraph/tree/main/enterprise/cmd/frontend/internal/codeintel)
- [enterprise/cmd/frontend/internal/executorqueue/queues/codeintel](https://github.com/sourcegraph/sourcegraph/tree/main/enterprise/cmd/frontend/internal/executorqueue/queues/codeintel)

Docs:

- [Deployment documentation](deployment.md)
- [How LSIF indexes are processed](uploads.md)
- [How precise code intelligence queries are resolved](queries.md)
- [How code intelligence extensions resolve hovers](extensions.md)
- [How Sourcegraph auto-indexes source code](auto-indexing.md)
