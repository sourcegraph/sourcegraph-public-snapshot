# Life of a code intelligence query

This document describes how our backend systems serve code intelligence results to clients. There are multiple kinds of code intelligence queries:

- _Hover queries_ retrieve the hover text (documentation) associated with a symbol;
- _Definitions queries_ retrieve a list of definitions (generally one) of a symbol, possibly defined in a different repository; and
- _References queries_ retrieve a list of uses of a symbol, possibly defined across multiple repositories.

The results of each query can be _precise_ or _fuzzy_, depending on the quality of data available. This document will detail the conditions required in order for results to be precise.

## Clients

There are a few ways to perform a code intelligence query with Sourcegraph:

1. Using the GraphQL API served by our [frontend](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/cmd/frontend) service. The API only serves _precise_ code intelligence queries.
2. Using the [basic-code-intel](https://github.com/sourcegraph/sourcegraph-basic-code-intel) extension in the Sourcegraph UI. The extension attempts to serve _precise_ code intelligence via the GraphQL API, falling back to _fuzzy_ code intelligence based on search queries when no precise results are available.
3. Using a browser extension on a codehost such as GitHub or Bitbucket. The (browser) extension will use the basic-code-intel (Sourcegraph) extension to retrieve results.

These clients are discussed in turn.

### GraphQL API

All GraphQL queries for precise code intelligence must first resolve a `GitTree`, which is a specific path or directory in a repository at a specific commit. All code intelligence operations (definitions, references, and hovers) are nested under an `lsif` field, which resolves to a null value when there is no LSIF upload present for the git tree.

For information about how LSIF data is uploaded and processed, see [life of an LSIF upload](life-of-an-lsif-upload.md).

```graphql
query {
    repository(name: "github.com/foo/bar") {
        commit(rev: "0123456789012345678901234567890123456789") {
            blob(path: "/baz/bonk.go") {
                lsif {
                    definitions(line: 10, character: 42) {
                        ...
                    }
                }
            }
        }
    }
}
```

The example above resolves the [git tree](https://git-scm.com/book/en/v2/Git-Internals-Git-Objects) (via `blob`) for the file `/baz/bonk.go`, then asks for the definitions under the cursor position `10:42`. The resolver for the `lsif` field can be found [here](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%22func+%28r+*gitTreeEntryResolver%29+LSIF%28%22).

The resolver for definitions, references, and hover fields within the `lsif` field are defined in the enterprise codeintel package. The definition resolver is [a method of `lsifQueryResolver`](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+lsifQueryResolver%29+definitions+file:codeintel+&patternType=literal), for example. These resolvers are very basic and simply call a method on the [lsif-server client](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%22%29+Definitions%28%22+file:lsifserver/.*.go) that simply makes an HTTP request to the precise-code-intel-api-server (discussed below).

It may be the case that the `lsif` field is not null, but there is no known definition for a particular source location. This happens in particular when a symbol is defined in a repository that does not also have (properly correlated) LSIF data.

Code intelligence resolvers exist only in the enterprise version of the product. The OSS version will return a canned message for all LSIF requests.

### Basic code intel

The [basic-code-intel](https://github.com/sourcegraph/sourcegraph-basic-code-intel) repository automatically generates extensions for most languages. The [Go extension](https://github.com/sourcegraph/sourcegraph-go) and the [TypeScript extension](https://github.com/sourcegraph/sourcegraph-typescript) require the basic code intel package as a dependency, but are published to the extension registry separately (due to special support for language servers).

Code intelligence extensions register _providers_ which are called from the UI or browser extensions to get hover text for a symbol or for the locations of its definitions or references. For example, the definition provider is defined [here](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph-basic-code-intel%24+registerDefinitionProvider). This provider will first query the GraphQL API for LSIF data. If the query returns a result, it is returned. Otherwise, either the position is not defined in that LSIF upload, or there is no LSIF upload that can provide intelligence that commit and path. In this case, the extension falls back to _fuzzy_ code intelligence. A [search query](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph-basic-code-intel%24+async+definition%28+file:handler.ts) is constructed with the symbol name and search results are filtered for obvious non-relevance (e.g. the target module name not matching a source import statement). The hover and references provider are not dissimilar.

Code intelligence queries are displayed in a [hover overlay](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%3CWebHoverOverlay+file:Blob.tsx) over the code blob in the UI. Overlays are shown after the user hovers over a particular token on a line, which will immediately fire a hover and definition (preload) request to the basic code intel extension. If _Go to Definition_ is clicked, the user is navigated to the preloaded location. If _Find References_ is clicked, a subsequent references request is made to the basic code intel extension, and a file match results panel is populated.

### Browser extensions

The browser extension will display the same [hover overlay](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+class+HoverOverlayContainer) as the UI does.

## LSIF API Server

The [precise-code-intel-api-server](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/precise-code-intel) accepts code intelligence queries via HTTP requests. The payload of each query is a repository ID, a commit hash, a file path, and a position in the source file. The definition endpoint is defined [here](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+definitions+file:precise-code-intel/.*/routes), for example.

Each query attempts to load a open LISF upload file (formatted as a SQLite database) from disk. Each LSIF upload is associated with a repository, a commit, and a _root_. The target LSIF upload is the upload with the same repository, commit, and a root that is a prefix of the request file path. If there is no LSIF upload for that exact commit, we try to load the _closest_ database by traversing ancestor and descendent commits until we find an upload with a matching repository and root.

This SQLite file is then queried and returns the hover text or the set of locations associated with that source position. This is generally enough for hover text, but not enough or definitions and reference queries in the presence of multiple mutually-referential repositories.

After all _local_ location results are found in the SQLite file, the server will query Postgres with the repository and commit to find the set of packages that it defines (for remote references) or for the set of dependencies that it has (for remote definitions). This will return a list of additional repository and commit pairs, which are queried in a similar way.
