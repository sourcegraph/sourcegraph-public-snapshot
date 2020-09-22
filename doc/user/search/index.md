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

---

## Other tips

- When viewing a file or directory, press the `y` key to expand the URL to its canonical form (with the full 40-character Git commit SHA).
- To share a link to multi-line range in a file, click on the starting line number and shift-click on the ending line number (in the left-hand gutter).

### Max file size

By default, files larger than 1 MB are excluded from search results. Use the [search.largeFiles](../../admin/config/site_config.md#search-largeFiles) keyword to specify files to be indexed and searched regardless of size.

---

## Sourcegraph Cloud

[Sourcegraph Cloud](https://sourcegraph.com/search) is a public instance of Sourcegraph that lets you search inside any open-source project on GitHub. For demo purposes, you'll be prompted to narrow your query if it would search across more than 50 repositories. To lift this limitation or to search your organization's internal code, [run your own Sourcegraph instance](../../admin/install/index.md).
