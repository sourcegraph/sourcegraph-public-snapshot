# Basic code intelligence

Sourcegraph comes with built-in code intelligence provided by search-based heuristics.

## How does it work?

[Basic code intelligence](https://github.com/sourcegraph/sourcegraph-basic-code-intel) implements the 3 main code intelligence features:

- Jump to definition: it performs a [symbol search](../search/index.md#symbol-search)
- Hover documentation: it first finds the definition then extracts documentation from comments near the definition
- Find references: it performs a case-sensitive word-boundary cross-repository [plain text search](../search/index.md#powerful-flexible-queries) for the given symbol

Basic code intelligence also filters results by file extension and by imports at the top of the file for some languages.

## Why are my results sometimes incorrect?

Basic code intelligence uses search-based heuristics, rather than parsing the code into an AST. You will see incorrect results more often for tokens with common names (such as `Get`) than for tokens with more unique names simply because those tokens appear more often in the search index.

If you would like to have precise results where you are 100% confident that the definition or reference you are navigating to is for the symbol you hovered, we recommend utilizing LSIF for precise code intelligence. Even with LSIF enabled, you may occasionally see results from basic code intelligence. These scenarios are described in more detail in the [LSIF docs](./lsif.md).

## What languages are supported?

Basic code intelligence supports all of [the most popular programming languages](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22).

Are you using a language we don't support? [File a GitHub issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose) or [submit a PR](https://github.com/sourcegraph/sourcegraph-basic-code-intel#adding-a-new-sourcegraphsourcegraph-lang-extension).
