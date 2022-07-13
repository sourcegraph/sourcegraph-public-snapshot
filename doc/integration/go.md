# Go dependencies integration with Sourcegraph

You can use Sourcegraph with Go modules from any Go module proxy, including open source code from proxy.golang.org or a private proxy such as [Athens](https://github.com/gomods/athens).
This integration makes it possible to search and navigate through the source code of published Go modules (for example, [`gorilla/mux@v1.8.0`](https://sourcegraph.com/go/github.com/gorilla/mux@v1.8.0)).

Feature | Supported?
------- | ----------
[Repository syncing](#repository-syncing) | ✅
[Credentials](#credentials) | ✅
[Rate limiting](#rate-limiting) | ✅
[Repository permissions](#repository-syncing) | ❌

## Setup

To add Go dependencies to Sourcegraph you need to setup a Go dependencies code host:

1. As *site admin*: go to **Site admin > Manage code hosts**
1. Select **JVM Dependencies**.
1. Configure the connection by following the instructions above the text field. Additional fields can be added using <kbd>Cmd/Ctrl+Space</kbd> for auto-completion. See the [configuration documentation below](#configuration).
1. Press **Add repositories**.

## Repository syncing

There are currently two ways to sync Go dependency repositories.

* **Dependencies search**: Sourcegraph automatically syncs Go dependency repos that are found in `go.mod` files during a [dependencies search](../code_search/how-to/dependencies_search.md).
* **Code host configuration**: manually list dependencies in the `"dependencies"` section of the JSON configuration when creating the Go dependency code host. This method can be useful to verify that the credentials are picked up correctly without having to run a dependencies search.

Sourcegraph tries to find each dependency repository in all configured `"urls"` until it's found. This means you can configure a public proxy first and fallback to a private one second (e.g. `"urls": ["https://proxy.golang.org", "https://admin:foobar@athens.yourcorp.com"]`).

## Credentials

Each entry in the `"urls"` array can contain basic auth if needed (e.g. `https://user:password@athens.yourcorp.com`).

## Rate limiting

By default, requests to the Go module proxies will be rate-limited
based on a default internal limit. ([source](https://github.com/sourcegraph/sourcegraph/blob/main/schema/go-modules.schema.json))

```json
"rateLimit": {
  "enabled": true,
  "requestsPerHour": 57600.0
}
```
where the `requestsPerHour` field is set based on your requirements.

**Not recommended**: Rate-limiting can be turned off entirely as well.
This increases the risk of overloading the proxy.

```json
"rateLimit": {
  "enabled": false
}
```

## Repository permissions

⚠️ Go dependency repositories are visible by all users of the Sourcegraph instance.

## Configuration

Go dependencies code host connections support the following configuration options, which are specified in the JSON editor in the site admin "Manage code hosts" area.

<div markdown-func=jsonschemadoc jsonschemadoc:path="integration/go-modules.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/integration/go) to see rendered content.</div>
