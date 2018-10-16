# Sourcegraph architecture overview

This is a high level overview of our architecture at Sourcegraph so you can understand how our services fit together.

## Diagram

To view this diagram in its rendered form on GitHub, install the Sourcegraph extension and enable Mermaid.js rendering in the experimental options menu.

```mermaid
graph LR
    Frontend-- HTTP -->gitserver
    searcher-- HTTP -->gitserver

    query-runner-- HTTP -->Frontend
    query-runner-- Graphql -->Frontend
    repo-updater-- HTTP -->github-proxy
    github-proxy-- HTTP -->github[github.com]

    repo-updater-- HTTP -->codehosts[Code hosts: GitHub Enterprise, BitBucket, etc.]
    repo-updater-->redis-cache

    Frontend-- HTTP -->query-runner
    Frontend-->redis-cache["Redis (cache)"]
    Frontend-- SQL -->db[Postgresql Database]
    Frontend-->redis["Redis (session data)"]
    indexer-->lsp-proxy
    Frontend-- HTTP -->searcher
    Frontend-- LSP over TCP -->lsp-proxy
    Frontend-- HTTP ---indexer
    Frontend-- HTTP ---repo-updater
    Frontend-- net/rpc -->indexed-search
    indexed-search[indexed-search/zoekt]-- HTTP -->Frontend

    indexer-- HTTP -->gitserver
    repo-updater-- HTTP -->gitserver

    lsp-proxy-->redis-cache

    lsp-proxy-- LSP over TCP -->langservers[Language servers: Go, Java, etc.]

    react[React App]-- Graphql -->Frontend
    react[React App]-- LSP over HTTP -->Frontend

    browser_extensions[Browser Extensions]-- Graphql -->Frontend
    browser_extensions[Browser Extensions]-- LSP over HTTP -->Frontend
```

## Services

Here are the services that compose Sourcegraph.

### frontend ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/frontend))

The frontend serves our [web app](web_app.md) and hosts our [GraphQL API](../api/graphql.md).

Application data is stored in our Postgresql database.

Session data is stored in redis.

### github-proxy ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/github-proxy))

Proxies all requests to github.com to keep track of rate limits and prevent triggering abuse mechanisms.

### gitserver ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/gitserver))

Mirrors repositories from their code host. All other Sourcegraph services talk to gitserver when they need data from git. Requests for fetch operations, however, should go through repo-updater.

### indexer ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/indexer))

The indexer has a few responsibilities:

- It keeps the cross-repo code intelligence indexes for repositories up to date.
- It makes sure the appropriate language servers are enabled (when a Docker socket is available, such as when using `sourcegraph/server` with the default `docker run` command).
- It is how the frontend enqueues repositories for updating when (e.g. a user visits a repository).

### lsp-proxy ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/lsp-proxy))

[Language Server Protocol](https://microsoft.github.io/language-server-protocol/)

Handles all LSP requests and routes them to the appropriate language server.

### Language servers

Language servers implement the [Language Server Protocol](https://microsoft.github.io/language-server-protocol/) to provide code intelligence (hover tooltips, jump to definition, find references).

We have built some language servers ourself ([Go](https://github.com/sourcegraph/go-langserver), [Java](https://github.com/sourcegraph/java-langserver), [TypeScript/JavaScript](https://github.com/sourcegraph/javascript-typescript-langserver), [Python](https://github.com/sourcegraph/python-langserver), [PHP](https://github.com/felixfbecker/php-language-server)), and we can also integrate language servers built by the community by wrapping them with [lsp-adapter](https://github.com/sourcegraph/lsp-adapter).

### query-runner ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/query-runner))

Periodically runs saved searches and sends notification emails.

### repo-updater ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/repo-updater))

Repo-updater (which may get renamed since it does more than that) tracks the state of repos, and is responsible for automatically scheduling updates ("git fetch" runs) using gitserver. Other apps which desire updates or fetches should be telling repo-updater, rather than using gitserver directly, so repo-updater can take their changes into account.

### searcher ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/searcher))

Provides on-demand search for repositories. It scans through a git archive fetched from gitserver to find results.

### indexed-search/zoekt ([code](https://github.com/sourcegraph/zoekt))

Provides search results for repositories that have been indexed. This is a paid feature.

We forked https://github.com/google/zoekt.

### symbols ([code](https://github.com/sourcegraph/sourcegraph/tree/master/cmd/symbols))

Indexes symbols in repositories using Ctags.

### syntect ([code](https://github.com/sourcegraph/syntect_server))

Syntect is a Rust service that is responsible for syntax highlighting.

### Browser extensions ([code](https://github.com/sourcegraph/browser-extensions) | [docs](https://docs.sourcegraph.com/integration/browser_extension))

We publish browser extensions for Chrome, Firefox, and Safari, that provide code intelligence (hover tooltips, jump to definition, find references) when browsing code on code hosts. By default it works for open-source code, but it also works for private code if your company has a Sourcegraph deployment.

It uses GraphQL APIs exposed by the frontend to fetch data.

### Editor extensions ([docs](https://docs.sourcegraph.com/integration/editor))

Our editor extensions provide lightweight hooks into Sourcegraph, currently.
