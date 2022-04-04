# <a href="https://sourcegraph.com"><picture><source srcset="https://raw.githubusercontent.com/sourcegraph/sourcegraph/main/ui/assets/img/sourcegraph-head-logo.svg" media="(prefers-color-scheme: dark)"/><img alt="Sourcegraph" src="https://raw.githubusercontent.com/sourcegraph/sourcegraph/main/ui/assets/img/sourcegraph-logo-light.svg" height="48px" /></picture></a>

[![build](https://badge.buildkite.com/00bbe6fa9986c78b8e8591cffeb0b0f2e8c4bb610d7e339ff6.svg?branch=main)](https://buildkite.com/sourcegraph/sourcegraph)

[Sourcegraph](https://about.sourcegraph.com/) is a fast and featureful code search and navigation engine.

![sourcegraph com_github com_golang_go_-_blob_src_net_http_request go_L855_6](https://user-images.githubusercontent.com/989826/126650657-cef98203-1505-4848-aab6-57acda1ec35f.png)

**Features**

- Fast global code search with a hybrid backend that combines a trigram index with in-memory streaming.
- Code intelligence for many languages via the [Language Server Index Format](https://lsif.dev/).
- Enhances GitHub, GitLab, Phabricator, and other code hosts and code review tools via the [Sourcegraph browser extension](https://docs.sourcegraph.com/integration/browser_extension).
- Integration with third-party developer tools via the [Sourcegraph extension API](https://docs.sourcegraph.com/extensions).

## Try it now

Sourcegraph has three editions:

- [Sourcegraph Cloud](https://sourcegraph.com) lets you search over the open-source ecosystem plus your private code. [Search strings](https://sourcegraph.com/search?q=context:global+failed+to+ensure+HEAD+exists&patternType=literal), [search patterns](https://sourcegraph.com/search?q=context:global+lang:python+range%28len%28:%5B1%5D%29%29&patternType=structural), [search symbols](https://sourcegraph.com/search?q=context:global+type:symbol+lang:typescript+%28OA%7Coa%7COa%29uth+%5BHh%5Dandler+-file:%28%5E%7C/%29node_modules/+&patternType=regexp&case=yes) and [find references](https://sourcegraph.com/github.com/spf13/cobra@a684a6d7f5e37385d954dd3b5a14fc6912c6ab9d/-/blob/command.go?L221:19&subtree=true#tab=references) across your entire codebase and the open-source world.
- [Sourcegraph Enterprise](https://docs.sourcegraph.com/#getting-started) lets you run your own Sourcegraph instance in your own environment.
- [Sourcegraph OSS](#sourcegraph-oss) is an open-source version of Sourcegraph that provides the core functionality of Sourcegraph (code search, code browsing, basic code navigation), but lacks more advanced features (enterprise authentication, repository permissions, admin controls, advanced code navigation, etc.)

> Source code for all three editions is contained in this repository. See the [License section](#license) for more details.

More:

- Install the open-source [browser extension](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en) to add Sourcegraph-like powers to your code review and code host.
- File feature requests and bug reports in [our issue tracker](https://github.com/sourcegraph/sourcegraph/issues).
- Visit [about.sourcegraph.com](https://about.sourcegraph.com) for more information about features, use cases, and organizations that use Sourcegraph.

## Self-hosted installation

### Sourcegraph Enterprise (free up to 10 users)

The fastest way to run Sourcegraph self-hosted is with the Docker container. See the [quickstart installation guide](https://docs.sourcegraph.com/#getting-started). There are also several additional ways of running a [production instance](https://docs.sourcegraph.com/admin/install).

### Sourcegraph OSS

1. Go through [Quickstart](https://docs.sourcegraph.com/dev/setup/quickstart) to install `sg` and dependencies
1. Start the development environment in OSS mode:
   ```sh
   sg start oss
   ```

Sourcegraph should now be running at https://sourcegraph.test:3443.

For detailed instructions and troubleshooting, see the [local development documentation](https://docs.sourcegraph.com/dev).

## Development

Refer to the [Developing Sourcegraph guide](https://docs.sourcegraph.com/dev) to get started.

### Documentation

The `doc` directory has additional documentation for developing and understanding Sourcegraph:

- [Project FAQ](./doc/admin/faq.md)
- [Architecture](./doc/dev/background-information/architecture/index.md): high-level architecture
- [Database setup](./doc/dev/background-information/postgresql.md): database best practices
- [Go style guide](./doc/dev/background-information/languages/go.md)
- [Documentation style guide](https://handbook.sourcegraph.com/engineering/product_documentation)
- [GraphQL API](./doc/api/graphql/index.md): useful tips when modifying the GraphQL API
- [Contributing](./CONTRIBUTING.md)

## License

This repository contains both OSS-licensed and non-OSS-licensed files. We maintain one repository rather than two separate repositories mainly for development convenience.

All files in the `enterprise` and `client/web/src/enterprise` fall under [LICENSE.enterprise](LICENSE.enterprise).

The remaining files fall under the [Apache 2 license](LICENSE.apache). Sourcegraph OSS is built only from the Apache-licensed files in this repository.
