# Sourcegraph Architecture Overview

This is a high level overview of Sourcegraph's architecture so you can understand how our systems fit together.
You can click on each component to jump to its respective code repository or subtree.

<object data="/dev/architecture/architecture.svg" type="image/svg+xml" style="width:100%; height: 100%">
</object>

To re-generate this diagram from the `architecture.dot` file with Graphviz, run: `dot -Tsvg -o architecture.svg architecture.dot`.

## Clients

We maintain multiple Sourcegraph clients:

- [Web application](https://github.com/sourcegraph/sourcegraph/tree/master/web)
- [Browser extensions](https://github.com/sourcegraph/sourcegraph/tree/master/browser)
- [src-cli](https://github.com/sourcegraph/src-cli)
- [Editor integrations](https://docs.sourcegraph.com/integration/editor)
  - [Visual Studio Code](https://github.com/sourcegraph/sourcegraph-vscode)
  - [Atom](https://github.com/sourcegraph/sourcegraph-atom)
  - [IntelliJ](https://github.com/sourcegraph/sourcegraph-jetbrains)
  - [Sublime](https://github.com/sourcegraph/sourcegraph-sublime)

These clients generally communicate with a Sourcegraph instance (either https://sourcegraph.com or a private customer instance) through our [GraphQL API](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/cmd/frontend/graphqlbackend/schema.graphql). There are also a small number of REST endpoints for specific use-cases.

## Services

Our backend is composed of multiple services:

- Most are Go services found in the [cmd](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/cmd) folder.
- [Syntect server](https://sourcegraph.com/github.com/sourcegraph/syntect_server) is our syntax highlighting service written in Rust. It is not horizontally scalable so only 1 replica is supported.
- [LSIF server](https://github.com/sourcegraph/sourcegraph/tree/master/lsif) provide precise code intelligence based on the LSIF data format. It is written in TypeScript.
- [zoekt-indexserver](https://sourcegraph.com/github.com/sourcegraph/zoekt/-/tree/cmd/zoekt-sourcegraph-indexserver) and [zoekt-webserver](https://sourcegraph.com/github.com/sourcegraph/zoekt/-/tree/cmd/zoekt-webserver) provide indexed search. It is written in Go.

## Infrastructure

- [sourcegraph/infrastructure](https://sourcegraph.com/github.com/sourcegraph/infrastructure) contains Terraform configurations for Cloudflare DNS and Site 24x7 monitoring, as well as build steps for various docker images. Only private docker images should be added here, public ones belong in the main repository.
- [sourcegraph/deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) contains YAML that can be used by customers to deploy Sourcegraph to a Kubernetes cluster.
- [sourcegraph/deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker) contains a pure-Docker cluster deployment reference that some one-off customers use to deploy Sourcegraph to a non-Kubernetes cluster.
  - [sourcegraph/deploy-sourcegraph-dot-com](https://github.com/sourcegraph/deploy-sourcegraph-dot-com) is a fork of the above that is used to deploy to the Kubernetes cluster that serves https://sourcegraph.com.

## Guides

Here are some guides to help you understand how multiple systems fit together:

- [Life of a search query](life-of-a-search-query.md)
- [Search pagination](search-pagination.md)
- Future topics we will cover here:
  - Life of a repository (i.e. how does code end up on gitserver?)
  - Sourcegraph extension architecture
  - Web app and browser extension architecture
