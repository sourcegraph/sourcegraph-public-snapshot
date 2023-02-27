# Python dependencies integration with Sourcegraph

You can use Sourcegraph with Python packages from any Python package mirror, including open source code from pypi.org or a private mirror such as [Nexus](https://www.sonatype.com/products/nexus-repository).
This integration makes it possible to search and navigate through the source code of published Python packages (for example, [`numpy@v1.19.5`](https://sourcegraph.com/python/numpy@v1.19.5)).

Feature | Supported?
------- | ----------
[Repository syncing](#repository-syncing) | ✅
[Repository permissions](#repository-permissions) | ❌

## Setup

See the "[Python dependencies](../admin/external_service/python.md)" documentation.

## Repository syncing

Site admins can [add Python packages to Sourcegraph](../admin/external_service/python.md#repository-syncing).

## Repository permissions

⚠️ Python dependency repositories are visible by all users of the Sourcegraph instance.
