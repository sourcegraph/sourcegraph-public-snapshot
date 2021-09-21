# API docs developer guide

This article is for engineers working at Sourcegraph, if you're using Sourcegraph:

* [Learn more about Sourcegraph's API docs feature](../../../../code_intelligence/apidocs/index.md)
* [Learn about using Sourcegraph's APIs](../../../../api/index.md)

**Table of contents**

- [API docs developer guide](#api-docs-developer-guide)
  - [State of affairs](#state-of-affairs)
  - [General architecture](#general-architecture)

## State of affairs

We've focused heavily on feature development and not breaking any other parts of Sourcegraph. At the same time, we've intentionally not spent time on e.g. testing which can make the code brittle if you don't understand what is going on. We're more focused on proving that this is a valuable feature than we are on productionisation. This is intentional.

## General architecture

API docs is primarily composed of:

* LSIF data generation (80% of the logic):
  * An extension to LSIF, [defined in `protocol/documentation.go`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:protocol/documentation.go&patternType=literal) which enables language indexers to emit documentation pages for workspaces.
  * An implementation in the lsif-go indexer, defined in [`internal/indexer/documentation.go`](https://github.com/sourcegraph/lsif-go/blob/master/internal/indexer/documentation.go), which implements the LSIF extension and does all the hard work of emitting documentation pages in a hierarchial node-based format.
* LSIF data consumption (10% of the logic):
  * [An implementation in our LSIF codeintel backend](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:codeintel+file:documentation+-file:protocol&patternType=literal) which consumes the LSIF data in this new extension format, processes it to be easier to work with, and stores the data in the codeintel database.
  * [The database tables](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:schema.codeintel.md+lsif_data_documentation&patternType=literal) which store the documentation pages, metadata to map between a section of documentation and e.g. a code file in a repository, and other information like trigram indexes for searching over API docs.
  * The GraphQL API exposed to clients for consuming the data.
* The frontend (10% of the logic):
  * [Various React components](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:documentation+file:%5C.tsx%24&patternType=literal) which fetch the API docs data via GraphQL and render it.
  * This is a bit of a "dumb client" in the fact that the backend has already precomputed what all of the documentation looks like, how it is annotated, etc. The client gets a tree of nodes that are just documentation labels/descriptions making up each section of a document, and the client merely renders that Markdown and adds some very minor functionality on top to make navigating it easier.
  * Most of what the frontend does is styling/layout, but even that is limited as the underlying data format is pretty opaque: a page, with sections of documentation each composed of a title/description and some metadata. The client cannot really choose to do much introspection on the data. There is no deep understanding of the code at this layer, it's just a display layer.
