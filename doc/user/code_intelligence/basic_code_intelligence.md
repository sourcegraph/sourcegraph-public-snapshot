# Basic code intelligence

Sourcegraph comes with out of the box code intelligence provided by search based heuristics... Desccribe basic code intelligence in <3 sentences

## How does it work?

TODO

## Why are my results sometimes incorrect?

This is because we are searching for results with the name of the requested token, rather than parsing the tree.

If you would like to have precise results where you are 100% confident that the definition or reference you are navigating to is for the symbol you hovered, we recommend utilizing [LSIF]((./lsif.md)) for precise code intelligence.

You may occasionally see results from basic code intelligence even when you have uploaded LSIF data. This can happen in the following scenarios:

- This file has changed since the last time LSIF data was uploaded.
- TODO...

## What languages are supported?

Basic code intelligence is supported for all of the most popular [programming language extensions](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22).

Are you using a language we don't support? [File a GitHub issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose) or [submit a PR](./todo_link_to_docs_for_adding_a_new_language).
