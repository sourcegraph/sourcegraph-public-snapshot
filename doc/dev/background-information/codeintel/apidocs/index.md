# API docs developer guide

This article is for engineers working at Sourcegraph, if you're using Sourcegraph:

* [Learn more about Sourcegraph's API docs feature](../../../../code_intelligence/apidocs/index.md)
* [Learn about using Sourcegraph's APIs](../../../../api/index.md)

**Table of contents**

- [API docs developer guide](#api-docs-developer-guide)
  - [State of affairs](#state-of-affairs)
  - [General architecture](#general-architecture)
  - [Search architecture](#search-architecture)
    - [Tradeoffs](#tradeoffs)
    - [What is indexed?](#what-is-indexed)
    - [Scaling estimation](#scaling-estimation)
    - [Limiting the search index size](#limiting-the-search-index-size)
    - [Public vs. private repositories](#public-vs-private-repositories)

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

## Search architecture

### Tradeoffs

API docs search architecture is built around Postgres pg_trgm extension. This is in contrast to every other part of Sourcegraph's search backend, which can be described as:

* Live searches (git grep, live symbols search, live git diff/commit search, etc.)
* Indexed Zoekt searches (text search, indexed symbol search)

At present, Sourcegraph's search architecture is not set up to facilitate the straightforward addition of more indexed content: there is no practical, straightforward, easy way to chuck API docs content into our Zoekt backend and ask Sourcegraph to index it and expose that. It would be a fairly substantial undertaking, perhaps on the order of a few months, and for a feature that is not really proven out yet.

For this reason, and with the search team's advice, and because API docs content size is relatively small compared to e.g. what Zoekt is regularly indexing, we've chosen to go the route of storing API docs in the codeintel Postgres database (which is conveniently nearby) and use pg_trgm to provide trigram based search indexing over the content.

It's much more important that we prove out API docs search is a useful feature before investing more into it, than it is to implement that functionality properly into our existing search backend (which, would also be a desirable time to think about how LSIF data in general fits into our search backend -- a very non-trivial architecture decision with many stakeholders.) For this reason, API docs search isn't properly implemented into our search backend - it's the "what can we get to that is reasonable, very fast to build, can scale for a year or two reasonably, and won't cause any harm to the rest of Sourcegraph?" option.

In the future, we may to revisit these architecture decisions entirely.

### What is indexed?

The most recent LSIF bundle upload, iff it is for the default Git branch, is indexed. In practice this generally means the most-recent default-branch commit.

When a new LSIF upload is created, the older commit is removed from the search index (for the same bundle upload root / indexer only.)

### Scaling estimation

To estimate the table & index sizes, four repositories were used:

* github.com/sourcegraph-testing/titan
* github.com/sourcegraph-testing/etcd
* github.com/sourcegraph-testing/zap
* github.com/sourcegraph-testing/tidb

These repositories include 43MB, 3,043 files, 1.14 million lines, and 109,318 symbols of Go code.

The `lsif_data_documentation_search_*` tables for this data ends up as:

* Total table size (sum of both): 51 MB
* Total size of content in trigram indexes: 14MB

We can compare this to some other similar tables in Sourcegraph's DB with the same repos to get a sense of how much data this is in comparison:

* `lsif_data_references`: 89 MB table size
* `lsif_data_definitions`: 35 MB table size
* `lsif_data_documents`: 216 MB table size

This implies that the `lsif_data_documentation_search_*` tables is relatively small in comparison to other tables. However, it will still be the largest Postgres _trigram_ index in Sourcegraph for the foreseeable future - and Postgres trigram indexes are primarily CPU bound.

In experiments outside Sourcegraph, I have empirical measurements that approx. 82 GiB of data in a Postgres trigram index is doable on a Macbook (2.3 GHz 8-Core Intel Core i9 + 16 GB 2667 MHz DDR4). If we take the average repo's trigram content size (14 MB / 4), we can ballpark estimate how much resources Postgres would need for a given number of repos:

* 25k repos / 683 million Go symbols: 1+ 2.3 Ghz i9 CPU
* 50k repos / 1.3 billion Go symbols: 2+ CPUs
* 75k repos / 2 billion Go symbols: 3+ CPUs

Memory consumption does not scale linearly with corpus size, and it seems likely that just 16-64 GB would do for potentially hundreds of thousands of repos.

As we begin to scale, search latency times may prove prohibitive and will require we move to a table-splitting approach to benefit from more CPU parallelism from Postgres on its trigram indexes. There is more information about this approach and other general Postgres trigram indexing concerns in [this blog post of mine which I created outside of work.](https://devlog.hexops.com/2021/postgres-regex-search-over-10000-github-repositories)

On Sourcegraph.com, only a few thousand repos have Go LSIF data (as of Sept 15, 2021 - although this is quickly growing and so may be outdated).

### Limiting the search index size

We make it easy to limit the amount of resources going to API docs search as a feature, since it is desirable to both prevent unbounded growth issues on e.g. Sourcegraph.com and prevent any unexpected resource consumption on enterprise instances (e.g. if someone out there has a Postgres instance provisioned well today, but has hundreds of thousands of Go repositories with LSIF indexing, adding this table may increase resource usage.)

In specific, a site configuration option `"apidocs.search-index-limit-factor": 1.0` enables limiting the index size. The value `1.0` is a multiple of 250 million symbols, i.e., `1.0` indicates 250 million symbols (approx 12.5k Go repos) can be in the public and private search indexes independently (500 million total), `2.0` indicates 500 million symbols (approx 50k Go repos), and so on.

We implement this by merely requesting an estimate number of rows in the table:

```
SELECT reltuples AS estimate FROM pg_class where relname = 'lsif_data_documentation_search_public';
```

And then deleting any rows which exceed the limit:

```
DELETE FROM lsif_data_documentation_search_public ORDER BY dump_id LIMIT estimate_rows-search_index_limit;
```

This removes entries that come from older LSIF uploads in general, and so typically the most actively committed to repositories will stay in the index more.

### Public vs. private repositories

Private repositories, and the API docs that come from them, require extensive and expensive ACL checks. We break search data into two tables currently:

* `lsif_data_documentation_search_public`
* `lsif_data_documentation_search_private`

The results from `_public` are always from repositories in the `repo` table with `private=false`, and hence can be returned to any user with access to Sourcegraph. This means ACL checks can be skipped.

In contrast, results from `_private` are from private repositories. It is only possible to return these if users can pass ACL checks, and as there may be a large number of results in here that users do not have access to at all, global search across this table generally does not make sense. Instead, a list of private repositories the user has access to is composed and the search includes a `WHERE repo_id IN (...)` conditional.

The breakdown between public and private search results is also nice because we can e.g. prioritize results from private repositories over those from public repositories easily.

Note that, due to the way Postgres trigram indexes work under the hood, this split of data (compared to say, using a `repo_is_private` boolean field on the same table) reduces the amount of data that must be considered by the trigram index and can substantially reduce query times (e.g. 50% faster if 50% of code is private.)
