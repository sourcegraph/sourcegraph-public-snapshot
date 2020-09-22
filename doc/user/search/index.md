# Code search

> → See the query [**syntax reference**](queries.md) and [**language reference**](language.md). See [search examples](examples.md) for inspiration.

This document is for code search users. To get code search, [install Sourcegraph](../../admin/install/index.md).

[A recently published research paper from Google](https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/43835.pdf) and a [Google developer survey](https://docs.google.com/document/d/1LQxLk4E3lrb3fIsVKlANu_pUjnILteoWMMNiJQmqNVU/edit#heading=h.xxziwxixfqq3) showed that 98% of developers consider their Sourcegraph-like internal code search tool to be critical, and developers use it on average for 5.3 sessions each day, primarily to (in order of frequency):

- find example code
- explore/read code
- debug issues
- determine the impact of changes

Sourcegraph code search helps developers perform these tasks more quickly and effectively by providing fast, advanced code search across multiple repositories. With Sourcegraph's code search, you can:

Sourcegraph provides fast, advanced code search across multiple repositories. With Sourcegraph's code search, you can:

- Use regular expressions and exact queries to perform full-text searches.
- Perform [language-aware structural search](#language-aware-structural-code-search) on code structure.
- Search any branch and commit, with no indexing required.
- Search [commit diffs](#commit-diff-search) and [commit messages](#commit-message-search) to see how code has changed.
- Narrow your search by repository and file pattern.
- Define saved [search scopes](#search-scopes) for easier searching.
- Curate [saved searches](#saved-searches) for yourself or your org.
- Set up notifications for code changes that match a query.
- View [language statistics](#statistics) for search results.

