# Code search overview

> See [**search query syntax**](queries.md) reference.

[A recently published research paper from Google](https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/43835.pdf) and a [Google developer survey](https://docs.google.com/document/d/1LQxLk4E3lrb3fIsVKlANu_pUjnILteoWMMNiJQmqNVU/edit#heading=h.xxziwxixfqq3) showed that 98% of developers consider their Sourcegraph-like internal code search tool to be critical, and developers use it on average for 5.3 sessions each day, primarily to (in order of frequency):

- find example code,
- explore/read code,
- debug issues, and
- determine the impact of changes.

Sourcegraph code search helps developers perform these tasks more quickly and effectively.

Sourcegraph provides fast, advanced code search across multiple repositories. With Sourcegraph's code search, you can:

- Use regular expressions and exact queries to perform full-text searches
- Search any branch and commit, with no indexing required
- Search [commit diffs](#commit-diff-search) and [commit messages](#commit-message-search) to see how code has changed
- Narrow your search by repository and file pattern
- Define saved [search scopes](#search-scopes) for easier searching
- Curate [saved searches](#saved-searches) for yourself or your org
- Set up notifications for code changes that match a query
- View [language statistics](#statistics) for search results

This document is for code search users. To get code search, [install Sourcegraph](../../admin/install/index.md).

---

## Features

### Powerful, flexible queries

Sourcegraph code search performs full-text searches and supports both regular expression and exact queries. By default, Sourcegraph searches across all your repositories. Our search [query syntax](queries.md) allows for advanced queries, such as searching over any branch or commit, narrowing searches by programming language or file pattern, and more.

See the [query syntax documentation](queries.md) for a comprehensive list of tokens.

### Commit diff search

Search over commit diffs using `type:diff` to see how your codebase has changed over time. This is often used to find changes to particular functions, classes, or areas of the codebase when debugging.

You can also search within commit diffs on multiple branches by specifying the branches in a `repo:` field after the `@` sign. After the `@`, separate Git refs with `:`, specify Git ref globs by prefixing them with `*`, and exclude commits reachable from a ref by prefixing it with `^`.

Diff searches can be further narrowed down with filters such as author and time. See the [query syntax documentation](queries.md#diff-and-commit-searches-only) for a comprehensive list of supported tokens.

### Commit message search

Searching over commit messages is supported in Sourcegraph by adding `type:commit` to your search query.

Separately, you can also use the `message:"any string"` token to filter `type:diff` searches for a given commit message.

Commit message searches can be further narrowed down with filters such as author and time. See our [query syntax documentation](queries.md#diff-and-commit-searches-only) for a comprehensive list of supported tokens.

### Symbol search

Searching for symbols makes it easier to find specific functions, variables and more. Use the `type:symbol` filter to search for symbol results. Symbol results also appear in typeahead suggestions, so you can jump directly to symbols by name.

### Saved searches

Saved searches let you save and describe search queries so you can easily monitor the results on an ongoing basis. You can create a saved search for anything, including diffs and commits across all branches of your repositories. Saved searches can be an early warning system for common problems in your code--and a way to monitor best practices, the progress of refactors, etc.

See the [saved searches documentation](saved_searches.md) for instructions for setting up and configuring saved searches.

### Search scopes

Every project and team has a different set of repositories they commonly work with and search over. Custom search scopes enable users and organizations to quickly filter their searches to predefined subsets of files and repositories. Instead of typing out the subset of repositories or files you want to search over, you can save and select scopes using the search scopes buttons whenever you need.

### Suggestions

As you type a query, the menu below will contain suggestions based on the query. Use the keyboard or mouse to select a suggestion to navigate directly to it. For example, if your query is `repo:foo file:\.js$ hello`, the suggestions will consist of the list of files that match your query.

You can also type in the partial name of a repository or filename to quickly jump to it. For example, typing in just `foo` would show you a list of repositories (first) and files with names containing _foo_.

### Statistics

> NOTE: To enable this experimental feature, set `{"experimentalFeatures": {"searchStats": true} }` in user settings.

On a search results page, press the **Stats** button to view a language breakdown of all results matching the query. Each matching file is analyzed to detect its language, and line count statistics are computed as follows:

- Query matches entire repositories (e.g., using only `repo:`): all lines (in all files) in matching repositories are counted.
- Query matches entire files (e.g., using only `file:` or `lang:`): all lines in all matching files are counted.
- Query matches text in files (e.g., using a term such as `foo`): all lines that match the query are counted.

Examples:

- If your search query was `file:test` and you had a single 100-line Java test file (and no other files whose name contains `test`), the statistics would show 100 Java lines.
- If your search query was `foo` and that term appeared on 3 lines in Java files and on 1 line in a Python file, the statistics would show 3 Java lines and 1 Python line.

Tip: On the statistics page, you can enter an empty query to see statistics across all repositories.

---

## Details

### Data freshness

Searches scoped to specific repositories are always up-to-date. Sourcegraph automatically refetches repository contents upon any user action specific to the repository and makes new commits and branches available for searching and browsing immediately.

Unscoped search results over large repository sets may trail latest default branch revisions by some interval of time. This interval is a function of the number of repositories and the computational resources devoted to search indexing.

### Max file size

By default, files larger than 1 MB are excluded from search results. Use the [search.largeFiles](../../admin/config/site_config.md#search-largeFiles) keyword to specify files to be indexed and searched regardless of size.

---

## Other tips

- When viewing a file or directory, press the `y` key to expand the URL to its canonical form (with the full 40-character Git commit SHA).
- To share a link to multi-line range in a file, click on the starting line number and shift-click on the ending line number (in the left-hand gutter).

---

## Sourcegraph.com

[Sourcegraph.com](https://sourcegraph.com/search) is a public instance of [Sourcegraph](../../admin/install/index.md) that lets you search inside any open-source project on GitHub. For demo purposes, you'll be prompted to narrow your query if it would search across more than 50 repositories. To lift this limitation or to search your organization's internal code, [run your own Sourcegraph instance](../../admin/install/index.md).
