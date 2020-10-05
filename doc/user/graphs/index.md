# Graphs

Graphs help make code search and code intelligence results more relevant to you, by excluding code in repositories you don't care about.

## About graphs

A graph defines the scope for code search and code intelligence. When you use Sourcegraph, you always have exactly one currently selected graph. The currently selected graph affects the results and other information you see in many parts of Sourcegraph.

You can define your own graphs that include only the code you care about for various tasks. Here are some examples:

- For everyday use: "The 3 repositories I mainly work in, plus all code of their dependencies"
- For service owners: "The directory of our monorepo that contains the service I own, plus all other projects that call my service's API"
- For debugging issues in past releases: "All projects and dependencies, at the various exact historical versions, that went into our June 2020 release"
- For finding high-quality usage examples: "All open-source projects using [Bootstrap](https://getbootstrap.com/) v4.5+ that have 1,000+ stars"

## Selecting a graph

You always have exactly one graph selected when using Sourcegraph. You can see and change it in the top navigation bar, on the left side:

![Screenshot of Sourcegraph with a selected graph named "lsp-lsif-repos"](https://user-images.githubusercontent.com/1976/94392257-a6868080-010c-11eb-9fac-731b39c3c131.png)

In the screenshot above, the selected graph is named `lsp-lsif-repos`.

To select a different graph:

1. Click the graph selector (**<em>Name of the currently selected graph</em> ▼**) to open a list of other graphs.
1. Click another graph to select it.

## Creating a graph

## Specifying a graph in the GraphQL API

Some queries and mutations in the [Sourcegraph GraphQL API](../../api/index.md) accept a graph as an input argument. API consumers must explicitly supply the graph as an argument if they want it to apply. Although Sourcegraph tracks a user's last-selected graph in the user interface, the GraphQL API never implicitly uses the user's last-selected graph.
