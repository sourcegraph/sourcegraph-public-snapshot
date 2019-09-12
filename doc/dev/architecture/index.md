# Sourcegraph Architecture Overview

This is a high level overview of Sourcegraph's architecture so you can understand how our systems fit together.

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

These clients generally communicate with a Sourcegraph instance (either https://sourcegraph.com or a private customer instance) through our [GraphQL API](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/cmd/frontend/graphqlbackend/schema.graphql).

## Services


