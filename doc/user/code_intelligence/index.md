# Code intelligence overview

Code intelligence provides advanced code navigation and cross-references for your code on Sourcegraph, your code host, and your code review tools:

- [Hover tooltips with documentation and type signatures](#hover-tooltips-with-documentation-and-type-signatures)
- [Go to definition](#go-to-definition)
- [Find references](#find-references)
- [Symbol search](#symbol-search)

Code intelligence works out of the box with all of the most popular [programming language extensions](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22) using our search-based [basic code intelligence](./basic_code_intelligence.md).

To get more precise code intelligence, you can upload [LSIF data](./lsif.md) for your repositories.

Code intelligence is provided by [Sourcegraph extensions](../../extensions/index.md).

By spinning up Sourcegraph, you can get code intelligence:

- On the Sourcegraph web interface
- On code files on your code host, via our [integrations](../../integration/index.md)
- On diffs in your code review tool, via our [integrations](../../integration/index.md)
- Via the Sourcegraph API (for programmatic access)

## Getting started

- [Set up Sourcegraph](../../admin/install/index.md), then enable the [Sourcegraph extension](../index.md) for each language you want to use. The language extensions should be on by default for a new instance.
- To add code intelligence to your code host and/or code review tool, see the [browser extension documentation](../../integration/browser_extension.md).
- Interested in trying it out on public code? See [this sample file](https://sourcegraph.com/github.com/dgrijalva/jwt-go/-/blob/token.go#L37:6$references) on Sourcegraph.com.

## Code intelligence features

### On a GitHub pull request

Using our [integrations](../../integration/index.md), all code intelligence features are available everywhere you read code!

<img src="img/CodeReview.gif" width="450" style="margin-left:0;margin-right:0;"/>

### Hover tooltips with documentation and type signatures

Hover tooltips allow you to quickly glance at the type signature and accompanying documentation of a symbol definition without having to context switch to another source file (which may or may not be available while browsing code).

<img src="img/hover-tooltip.png" width="500"/>

### Go to definition

When you select 'Go to definition' from the hover tooltip, you will be navigated directly to the definition of the symbol.

<img src="img/go-to-def.gif" width="500"/>

### Find references

When you select 'Find references' from the hover tooltip, a panel will be shown at the bottom of the page that lists all of the references found for both precise (LSIF or language server) and basic results (from search heuristics). This panel will separate references by repository, and you can optionally group them by file.

> NOTE: When a particular token returns a large number of references, we truncate the results to < 500 to optimize for browser loading speed. We are planning to improve this in the future with the ability to view it as a search so that users can utilize the powerful filtering of Sourcegraph's search to find the references they are looking for.

<img src="img/find-refs.gif" width="450"/>

### Symbol search

We use [Ctags](https://github.com/universal-ctags/ctags) to index the symbols of a repository on-demand. These symbols are used to implement symbol search, which will match declarations instead of plain-text.

<img src="img/Symbols.png" width="500"/>

#### Symbol sidebar

We use [Ctags](https://github.com/universal-ctags/ctags) to index the symbols of a repository on-demand. These symbols are also used for the symbol sidebar, which categorizes declarations by type (variable, function, interface, etc). Clicking on a symbol in the sidebar jumps you to the line where it is defined.

<img src="img/SymbolSidebar.png" width="500"/>
