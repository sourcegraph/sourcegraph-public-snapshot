# npm dependency integration with Sourcegraph

You can use Sourcegraph with npm packages from any npm registry, including open source code from npmjs.com or a private registry such as Verdaccio.
This integration makes it possible to search and navigate through the source code of published JavaScript or TypeScript packages (for example, [`@types/gzip-js@0.3.3`](https://sourcegraph.com/npm/types/gzip-js@v0.3.3/-/blob/index.d.ts)).

Feature | Supported?
------- | ----------
[Repository syncing](#repository-syncing) | ✅
[Credentials](#credentials) | ✅
[Repository permissions](#repository-syncing) | ❌
[Multiple npm dependency code hosts](#multiple-npm-dependency-code-hosts) | ❌

## Repository syncing

There are two ways to sync npm dependency repositories.

* **LSIF** (recommended): run [`lsif-node`](https://github.com/sourcegraph/lsif-node) against your JS/TS codebase and upload the generated index to Sourcegraph using the  [src-cli](https://github.com/sourcegraph/src-cli) command `src lsif upload`. Sourcegraph automatically synchronizes npm dependency repositories based on the dependencies that are discovered by `lsif-node`.
* **Code host configuration**: manually list dependencies in the `"dependencies"` section of the JSON configuration when creating the npm dependency code host. This method can be useful to verify that the credentials are picked up correctly without having to upload LSIF.

## Credentials

Use the `"credentials"` section of the JSON configuration to provide an access token for your private npm registry. See the [official npm documentation](https://docs.npmjs.com/about-access-tokens) for more details about how to create, list and view npm access tokens.


## Repository permissions

⚠️ npm dependency repositories are visible by all users of the Sourcegraph instance.

## Multiple npm dependency code hosts

⚠️ It's only possible to create one npm dependency code host for each Sourcegraph instance.
See the issue [sourcegraph#32499](https://github.com/sourcegraph/sourcegraph/issues/32499) for more details about this limitation. In most situations, it's possible to work around this limitation by configurating a single private npm registry to proxy multiple underlying registries.
