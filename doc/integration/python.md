# Python dependencies integration with Sourcegraph

You can use Sourcegraph with Python packages from any Python package mirror, including open source code from pypi.org or a private mirror such as [Nexus](https://www.sonatype.com/products/nexus-repository).
This integration makes it possible to search and navigate through the source code of published Python packages (for example, [`numpy@v1.19.5`](https://sourcegraph.com/python/numpy@v1.19.5)).

Feature | Supported?
------- | ----------
[Repository syncing](#repository-syncing) | ✅
[Credentials](#credentials) | ✅
[Rate limiting](#rate-limiting) | ✅
Repository permissions | ❌

## Repository syncing

There are currently two ways to sync Python dependency repositories.

* **Dependencies search**: Sourcegraph automatically syncs Python dependency repos that are found in some lockfiles files during a [dependencies search](../code_search/how-to/dependencies_search.md).
* **Code host configuration**: manually list dependencies in the `"dependencies"` section of the JSON configuration when creating the Python dependency code host. This method can be useful to verify that the credentials are picked up correctly without having to run a dependencies search.

Sourcegraph tries to find each dependency repository in all configured `"urls"` until it's found. This means you can configure a public mirror first and fallback to a private one second (e.g. `"urls": ["https://pypi.org", "https://admin:foobar@nexus.yourcorp.com"]`).

## Credentials

Each entry in the `"urls"` array can contain basic auth if needed (e.g. `https://user:password@nexus.yourcorp.com`).

## Rate limiting

By default, requests to the Python package mirrors will be rate-limited based on a default internal limit. ([source](https://github.com/sourcegraph/sourcegraph/blob/main/schema/python-packages.schema.json))

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

⚠️ Python dependency repositories are visible by all users of the Sourcegraph instance.
