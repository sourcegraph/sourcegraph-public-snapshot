# Basic code intelligence

Sourcegraph comes with out of the box code intelligence provided by search based heuristics.

## How does it work?

The basic code intelligence extension performs a symbol search query to determine the location of the definition. This enables _Jump to Definition_ and provides context from which hover text can be extracted. Word-boundary search queries are used to determine references across multiple repositories.

## Why are my results sometimes incorrect?

This is because we are searching for results with the name of the requested token, rather than parsing the tree. You will see incorrect results more often for tokens with common names (such as `Get`) than for tokens with more unique names simply because those tokens exist in the search index at a greater frequency. We do make attempts to filter out the _obviously_ wrong results (based on file extension, the imported packages for certain languages, and local syntatic analysis).

If you would like to have precise results where you are 100% confident that the definition or reference you are navigating to is for the symbol you hovered, we recommend utilizing [LSIF]((./lsif.md)) for precise code intelligence.

You may occasionally see results from basic code intelligence even when you have uploaded LSIF data. This can happen in the following scenarios:

- The commit you are viewing is not _close enough_ (e.g. 20 commits ahead) to a commit for which you uploaded LSIF data.
- The commit you are viewing is close enough to a commit for which you uploaded LSIF data, but the region of code you are viewing has changed since that commit.
- The _Find References_ panel will always include search-based results, but only after all of the precise results have been displayed.

## What languages are supported?

Basic code intelligence is supported for all of the most popular [programming language extensions](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22).

Are you using a language we don't support? [File a GitHub issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose) or [submit a PR](./todo_link_to_docs_for_adding_a_new_language).
