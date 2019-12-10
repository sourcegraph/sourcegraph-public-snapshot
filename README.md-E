# <a href="https://sourcegraph.com"><img alt="Sourcegraph" src="https://storage.googleapis.com/sourcegraph-assets/sourcegraph-logo.png" height="32px" /></a>

[![build](https://badge.buildkite.com/00bbe6fa9986c78b8e8591cffeb0b0f2e8c4bb610d7e339ff6.svg?branch=master)](https://buildkite.com/sourcegraph/sourcegraph)
[![apache license](https://img.shields.io/badge/license-Apache-blue.svg)](LICENSE)

[Sourcegraph](https://about.sourcegraph.com/) OSS edition is a fast, open-source, fully-featured code search and navigation engine. [Enterprise editions](https://about.sourcegraph.com/pricing) are available.

![Screenshot](https://user-images.githubusercontent.com/1646931/46309383-09ba9800-c571-11e8-8ee4-1a2ec32072f2.png)

**Features**

- Fast global code search with a hybrid backend that combines a trigram index with in-memory streaming
- Code intelligence for many languages via the [Language Server Protocol](https://langserver.org/)
- Enhances GitHub, GitLab, Phabricator, and other code hosts and code review tools via the [Sourcegraph browser extension](https://docs.sourcegraph.com/integration/browser_extension)
- Integration with third-party developer tools via the [Sourcegraph extension API](https://docs.sourcegraph.com/extensions)

## Try it yourself

- Try out the public instance on any open-source repository at [sourcegraph.com](https://sourcegraph.com/github.com/golang/go/-/blob/src/net/http/httptest/httptest.go#L41:6&tab=references).
- Install the free and open-source [browser extension](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en).
- Spin up your own instance with the [quickstart installation guide](https://docs.sourcegraph.com/#quickstart).
- File feature requests and bug reports in [our issue tracker](https://github.com/sourcegraph/sourcegraph/issues).
- Visit [about.sourcegraph.com](https://about.sourcegraph.com) for more information about product features.

## Development

### Prerequisites

- Git
- Go (1.13 or later)
- Docker
- PostgreSQL (v11 or higher)
- Node.js (version 8 or 10)
- Redis
- Yarn
- Nginx

For a detailed guide to installing prerequisites, see [these
instructions](doc/dev/local_development.md#step-1-install-dependencies).

### Installation

> Prebuilt Docker images are the fastest way to use Sourcegraph Enterprise. See the [quickstart installation guide](https://docs.sourcegraph.com/#quickstart).

To use Sourcegraph OSS:

1.  [Ensure Docker is running](doc/dev/local_development.md#step-3-macos-start-docker)
1.  [Initialize the PostgreSQL database](doc/dev/local_development.md#step-2-initialize-your-database)
1.  Start the development server

    ```
    ./dev/start.sh
    ```

Sourcegraph should now be running at http://localhost:3080.

For detailed instructions and troubleshooting, see the [local development documentation](./doc/dev/local_development.md).

### Documentation

The `docs` folder has additional documentation for developing and understanding Sourcegraph:

- [Project FAQ](./doc/admin/faq.md)
- [Architecture](./doc/dev/architecture/index.md): high-level architecture
- [Database setup](./doc/dev/postgresql.md): database setup and best practices
- [General style guide](./doc/team/style_guide.md)
- [Go style guide](./doc/dev/go_style_guide.md)
- [Documentation style guide](./team/product-dev/documentation/style_guide.md)
- [GraphQL API](./doc/dev/graphql_api.md): useful tips when modifying the GraphQL API
- [Contributing](./CONTRIBUTING.md)

### License

Sourcegraph OSS is available freely under the [Apache 2 license](LICENSE.apache). Sourcegraph OSS comprises all files in this repository except those in the `enterprise/` and `web/src/enterprise` directories.

All files in the `enterprise/` and `web/src/enterprise/` directories are subject to the [Sourcegraph Enterprise license](LICENSE.enterprise).
