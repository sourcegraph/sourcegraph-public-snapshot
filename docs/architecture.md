# Sourcegraph Architecture

This is a high level overview of our architecture at Sourcegraph so you can understand how our services fit together.

You should take a moment to browse our public documentation and marketing to see what our customers see:
https://about.sourcegraph.com/

## Services

Here are the services that compose Sourcegraph.

### frontend ([code](https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph/-/tree/cmd/frontend))

The frontend serves our [web app](https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph/-/tree/web) and hosts our [GraphQL API](https://about.sourcegraph.com/docs/features/api/).

Application data is stored in our Postgresql database.

Session data is stored in redis.

### github-proxy ([code](https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph/-/tree/cmd/github-proxy))

Proxies all requests to github.com to keep track of rate limits and prevent triggering abuse mechanisms.

### gitserver ([code](https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph/-/tree/cmd/gitserver))

Mirrors repositories from their code host. All other Sourcegraph services talk to gitserver when they need data from git.

### indexer ([code](https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph/-/tree/cmd/indexer))

The indexer has a few responsibilities:

* It keeps the cross-repo code intelligence indexes for repositories up to date.
* It makes sure the appropriate language servers are enabled (Sourcegraph Server only).
* It is how the frontend enqueues repositories for updating when (e.g. a user visits a repository).

### lsp-proxy ([code](https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph/-/tree/cmd/lsp-proxy))

[Language Server Protocol](https://microsoft.github.io/language-server-protocol/)

Handles all LSP requests and routes them to the appropriate language server.

### Language servers

Language servers implement the [Language Server Protocol](https://microsoft.github.io/language-server-protocol/) to provide code intelligence (hover tooltips, jump to definition, find references).

We have built some language servers ourself ([Go](https://github.com/sourcegraph/go-langserver), [Java](https://github.com/sourcegraph/java-langserver), [TypeScript/JavaScript](https://github.com/sourcegraph/javascript-typescript-langserver), [Python](https://github.com/sourcegraph/python-langserver), [PHP](https://github.com/felixfbecker/php-language-server)), and we can also integrate language servers built by the community by wrapping them with [lsp-adapter](https://github.com/sourcegraph/lsp-adapter).

### query-runner ([code](https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph/-/tree/cmd/query-runner))

Periodically runs saved searches and sends notification emails.

### repo-updater ([code](https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph/-/tree/cmd/repo-updater))

TODO

### searcher ([code](https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph/-/tree/cmd/searcher))

Provides on-demand search for repositories. It scans through a git archive fetched from gitserver to find results.

### indexed-search/zoekt ([code](https://github.com/sourcegraph/zoekt))

Provides search results for repositories that have been indexed. This is a paid feature.

We forked https://github.com/google/zoekt.

### symbols ([code](https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph/-/tree/cmd/symbols))

Indexes symbols in repositories using Ctags.

### syntect ([code](https://github.com/sourcegraph/syntect_server))

Syntect is a Rust service that is responsible for syntax highlighting.

## Other products

### Browser extensions ([code](https://sourcegraph.sgdev.org/github.com/sourcegraph/browser-extension) | [docs](https://about.sourcegraph.com/docs/features/browser-extension/))

We publish browser extensions for Chrome, Firefox, and Safari, that provide code intelligence (hover tooltips, jump to definition, find references) when browsing code on code hosts. By default it works for open source code, but it also works for private code if your company has a Sourcegraph deployment.

It uses GraphQL APIs exposed by the frontend to fetch data.

### Editor extensions ([docs](https://about.sourcegraph.com/docs/integrations/editor-plugins))

Our editor extensions provide lightweight hooks into Sourcegraph.

### Sourcegraph editor ([code](https://github.com/sourcegraph/src))

We forked Visual Studio Code to experiment with doing code review in the editor. We are not actively working on this anymore.
