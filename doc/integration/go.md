# Go dependencies integration with Sourcegraph

You can use Sourcegraph with Go modules from any Go module proxy, including open source code from proxy.golang.org or a private proxy such as [Athens](https://github.com/gomods/athens).
This integration makes it possible to search and navigate through the source code of published Go modules (for example, [`gorilla/mux@v1.8.0`](https://sourcegraph.com/go/github.com/gorilla/mux@v1.8.0)).

Feature | Supported?
------- | ----------
[Repository syncing](#repository-syncing) | ✅
[Repository permissions](#repository-syncing) | ❌

## Setup

See the "[Go dependencies](../admin/external_service/go.md)" documentation.

## Repository syncing

Site admins can [add Go packages to Sourcegraph](../admin/external_service/go.md#repository-syncing).

## Repository permissions

⚠️ Go dependency repositories are visible by all users of the Sourcegraph instance.
