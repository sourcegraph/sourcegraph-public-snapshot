# Indexed ranking

This document describes the current strategies used in Sourcegraph to rank results. Currently, ranking only
applies to indexed (Zoekt-based) search. When a search is unindexed, for example when searching at an old
revision, the results are not ranked.

> Note: this is an area of active research and is subject to change.

## Streaming-based search

Zoekt takes a streaming approach to executing searches. This makes ranking results more challenging, as the search
doesn't visit the full set of matching documents before returning results. As a compromise, Zoekt performs an initial
non-streamed search step, then ranks and returns those candidates before streaming the rest of the results.

Specifically, each Zoekt replica:
1. Collects candidate matches until a certain time limit (default of 500ms)
2. Ranks and streams back the ranked matches
3. Then switches to streaming execution, where matching results are immediately returned

Frontend waits until it has received at least one response from every replica, then merges and ranks the results
before returning them. After this initial ranked batch, frontend switches to immediately streaming out results.

Because the Zoekt limit is time-based, it's possible that executing the same search twice can result in different
ranked results. To mitigate this issue, you can increase the time limit through the site config `experimentalFeatures.ranking.flushWallTimeMS`.
Larger values give a more stable ranking, but searches can take longer to return an initial result.

## Result Ranking

Zoekt creates a [score for a match](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/zoekt%24+matchScore&patternType=literal) based on a few heuristics. In order of importance:

- It matches a symbol, such as an exact match on the name of a class.
- The match is at the start or end of a symbol. For example if you search for `Indexed`, then a class called `IndexedRepo` will score more highly than one named `NonIndexedRepo`.
- It partially matches a symbol. Symbols are a sign of something important, so any overlap is better than none.
- It matches a full word. For example, if you search `rank`, then `result rank` will score more highly than `ranked list`.
- It partially matches a word. For example, if you search `rank`, then `result rank` will score more highly than `ranked list`.

We also incorporate smaller signals based on the repository and file importance (described in the next section), as well as the number of query components that match (in the case of OR queries).

## Ordering files by importance

When creating indexes, we lay out the files such that we search more important files and repositories first. This means when streaming we're more likely to encounter important candidates first, leading to a better set of ranked results.

Zoekt indexes are partitioned by repository. The search proceeds through each repository in order of their priority.
The [repository priority](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+stars+reporank&patternType=regexp) is the number of stars a repository has received. Admins can manually adjust the priority of a repository through a [site configuration option](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+repoRankFromConfig&patternType=regexp).

Within each repository, files are ordered in terms of importance:
- Down rank generated code. This code is usually the least interesting in results.
- Down rank vendored code. Developers are normally looking for code written by their organisation.
- Down rank test code. Developers normally prefer results in non-test code over test code.
- Up rank files with lots of symbols. These files are usually edited a lot.
- Up rank small files. If you have similiar symbol levels, prefer the shorter file.
- Up rank short names. The closer to the project root the likely more important you are.
- Up rank branch count. if the same document appears on multiple branches its likely more important.

## References

- [RFC 359](https://docs.google.com/document/d/1EiD_dKkogqBNAbKN3BbanII4lQwROI7a0aGaZ7i-0AU/edit#heading=h.trqab8y0kufp): Search Result Ranking
- [Zoekt design reference](https://github.com/sourcegraph/zoekt/blob/master/doc/design.md#ranking)

