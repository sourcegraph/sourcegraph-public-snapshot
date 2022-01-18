# Indexed ranking

This document describes the current strategies used in Sourcegraph to rank results.

> Note: this is an area of active research and is subject to change.

## Streaming note

We are streaming based internally. However, this does not mean we can't do ranking. We have buffers in several layers which collect results and based on heuristics send out the buffered results in a ranked order. Additionally when creating indexes we lay out files and repositories such that we search more important files and repositories first. This means when streaming we receive likely more important results first.

When searching in general there are limits you hit before you inspect the full corpus of code. These limits are usually time or result based. As such searching more important code first is a normal strategy employed such that the candidate documents are more likely to be relevant.

## Repository Ranking

We sort results bucketed by repositories. We then sort those buckets based on repository priority. So any ranking based on file contents or query only happens within those buckets.

The [repository priority](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+stars+reporank&patternType=regexp) is the number of stars a repository has received. Additionally an admin can adjust the priority of a repository via [configuration](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+repoRankFromConfig&patternType=regexp).

This has the downside of a poor result in a highly ranked repository will appear before a good result in a poorly ranked repository. The upside is normally it does the right thing and leads to more deterministic ordering in results.

## Result Ranking

Zoekt [ranks](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/zoekt%24+func+rank&patternType=literal) documents before storing them in the index. It uses this rank to ensure more important documents are searched first. The current heuristic signals in order of importance:

- down rank generated code :: This code is usually the least interesting in results.
- down rank vendored code :: Developers are normally looking for code written by their organisation.
- down rank test code :: Developers normally prefer results in non-test code over test code.
- up rank files with lots of symbols :: These files are usually edited a lot.
- up rank small files :: if you have similiar symbol levels, prefer the shorter file.
- up rank short names :: The closer to the project root the likely more important you are.
- up rank branch count :: if the same document appears on multiple branches its likely more important.

This ranking is used to decide the order we search the documents, so is most important when hitting limits. However, the order of documents is used as a signal when ranking so is used to distingiush similar looking results in different files. IE a match for the symbol `MyClass` for the query `MyClass` will be ranked higher in normal code vs test code.

Zoekt creates a [score for a match](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/zoekt%24+matchScore&patternType=literal) based on a few heuristics. In order of importance:

- Is your match a symbol. eg exactly matching the name of a class.
- Is your match at the start or end of a symbol. eg you search `Foo` is better than `Bar` for a class called `FooBarBaz`.
- Is your match partially in a symbol. Symbols are a sign of something important, so any overlap is better than none.
- Word match on any text. eg `ranking` is better than `rank` for the text `search result ranking`.
- Partial word match. eg `rank` is better than `ank` for the text `search result ranking`.

These are the main inputs for deciding the rank of a match. There is some minor signals based on the file rank (described at the start of the section), the number of matches in a file, etc. These are used more as tiebreaking, where the above ranking is close.

## References

- [RFC 359](https://docs.google.com/document/d/1EiD_dKkogqBNAbKN3BbanII4lQwROI7a0aGaZ7i-0AU/edit#heading=h.trqab8y0kufp): Search Result Ranking
- Zoekt Design :: [Ranking](https://github.com/google/zoekt/blob/master/doc/design.md#ranking)
