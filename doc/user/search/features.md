
## Features

### Powerful, flexible queries

Sourcegraph code search performs full-text searches and supports both regular expression and exact queries. By default, Sourcegraph searches across all your repositories. Our search query syntax allows for advanced queries, such as searching over any branch or commit, narrowing searches by programming language or file pattern, and more.

See the [query syntax](queries.md) and [query reference](language.md) documentation for a comprehensive overview of supported syntax.

### Language-aware structural code search

Sourcegraph supports advanced code search for specifically matching patterns inside code structures, like function parameters and loop bodies. 

See the [structural search](structural.md) documentation for a detailed explanation of this search mode. 

#### Exclude files and directories

You can exclude files and directories from search by adding the file _.sourcegraph/ignore_ to
the root directory of your repository. Sourcegraph interprets each line in the _ignore_ file as globbing
pattern. Files or directories matching those patterns will not show up in the search results.
 
The _ignore_ file is tied to a commit. This means, if you committed an _ignore_ file to a 
feature branch but not to your default branch, then only search results for the feature branch
will be filtered while the default branch will show all results.

Example:
```
# .sourcegraph/ignore
# lines starting with # are comments and are ignored
# empty lines are ignored, too

# ignore the directory node_modules/
node_modules/

# ignore the directory src/data/
src/data/

# ** matches all characters, while * matches all characters except /
# ignore all JSON files
**.json

# ignore all JSON files at the root of the repository
*.json

# ignore all JSON files within the directory data/
data/**.json

# ignore all data folders
data/
**/data/

# ignore all files that start with numbers
[0-9]*.*
**/[0-9]*.*
```

Our syntax follows closely what is documented in 
[the linux documentation project](https://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm).
However, we distinguish between `*` and `**`: While `**` matches all characters, `*` matches all characters 
except the path separator `/`.

Note that invalid globbing patterns will cause an error and searches over commits containing a broken _ignore_ file 
will not return any result.

### Commit diff search

Search over commit diffs using `type:diff` to see how your codebase has changed over time. This is often used to find changes to particular functions, classes, or areas of the codebase when debugging.

You can also search within commit diffs on multiple branches by specifying the branches in a `repo:` field after the `@` sign. After the `@`, separate Git refs with `:`, specify Git ref globs by prefixing them with `*`, and exclude commits reachable from a ref by prefixing it with `^`. Diff searches can be further narrowed down with parameters that filter by author and time. 

See the [query syntax](queries.md#diff-and-commit-searches-only) documentation for a comprehensive list of supported parameters.

### Commit message search

Searching over commit messages is supported in Sourcegraph by adding `type:commit` to your search query. Separately, you can also use the `message:"any string"` parameter to filter `type:diff` searches for a given commit message. Commit message searches can narrowed down further with filters such as author and time. 

See our [query syntax ](queries.md#diff-and-commit-searches-only) documentation for a comprehensive list of supported parameters.

### Symbol search

Searching for symbols makes it easier to find specific functions, variables, and more. Use the `type:symbol` filter to search for symbol results. Symbol results also appear in typeahead suggestions, so you can jump directly to symbols by name.

### Saved searches

Saved searches let you save and describe search queries so you can easily monitor the results on an ongoing basis. You can create a saved search for anything, including diffs and commits across all branches of your repositories. Saved searches can be an early warning system for common problems in your code and a way to monitor best practices, the progress of refactors, etc.

See the [saved searches](saved_searches.md) documentation for instructions for setting up and configuring saved searches.

### Search scopes

Every project and team has a different set of repositories they commonly work with and search over. Custom search scopes enable users and organizations to quickly filter their searches to predefined subsets of files and repositories. Instead of typing out the subset of repositories or files you want to search, you can save and select scopes using the search scope buttons whenever you need.

#### Data freshness

Searches scoped to specific repositories are always up-to-date. Sourcegraph automatically fetches repository contents with any user action specific to the repository and makes new commits and branches available for searching and browsing immediately.

Unscoped search results over large repository sets may trail latest default branch revisions by some interval of time. This interval is a function of the number of repositories and the computational resources devoted to search indexing.

### Suggestions

As you type a query, the menu below will contain suggestions based on the query. Use the keyboard or mouse to select a suggestion. For example, if your query is `repo:foo file:\.js$ hello`, the suggestions will consist of the list of files that match your query.

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

### Version contexts <span class="badge badge-primary">experimental</span>

> NOTE: This feature is still in active development and must be enabled by a Sourcegraph site admin in site configuration.

Many organizations have old versions of code running in production and need to search across all the code for a specific release.

Version contexts allow creating sets of many repositories at specific revisions. When set, a version context limits your searches and code navigation actions (with basic code intelligence) to the repositories and revisions in the context.

Your site admin can add version contexts in site configuration under the `experimentalFeatures.versionContexts` setting. For example:

```json
"experimentalFeatures": {
  "versionContexts": [
   {
      "name": "srcgraph 3.15",
      "revisions": [
        {
          "repo": "github.com/sourcegraph/sourcegraph",
          "rev": "3.15"
        },
        {
          "repo": "github.com/sourcegraph/src-cli",
          "rev": "3.11.2"
        }
      ]
    }
  ]
}
```

To specify the default branch, you can set `"rev"` to `"HEAD"` or `""`.

After setting some version contexts, users can select version contexts in the dropdown to the left of the search bar.


> NOTE: All revisions specified in version contexts [will be indexed](#multi-branch-indexing-experimental).

### Multi-branch indexing <span class="badge badge-primary">experimental</span>

> NOTE: This feature is still in active development and must be enabled by a Sourcegraph site admin in site configuration.

The most common branch to search is your default branch. To speed up this common operation Sourcegraph maintains an index of the source code on your default branch. Some organizations have other branches that are regularly searched. To speed up search for those branches Sourcegraph can be configured to index up to 64 branches per repository.

Your site admin can configure indexed branches in site configuration under the `experimentalFeatures.search.index.branches` setting. For example:

``` json
"experimentalFeatures": {
  "search.index.branches": {
   "github.com/sourcegraph/sourcegraph": ["3.15", "develop"],
   "github.com/sourcegraph/src-cli": "next"
  }
}
```

Indexing multiple branches will add additional resource requirements to Sourcegraph (particularly memory). The indexer will deduplicate documents between branches. So the size of your index will grow in relation to the number of unique documents. Refer to our [resource estimator](../../admin/install/resource_estimator.md) to estimate whether additional resources are required.

> NOTE: The default branch (`HEAD`) is always indexed.

> NOTE: All revisions specified in version contexts are also indexed.

## Other tips

- When viewing a file or directory, press the `y` key to expand the URL to its canonical form (with the full 40-character Git commit SHA).
- To share a link to multi-line range in a file, click on the starting line number and shift-click on the ending line number (in the left-hand gutter).

### Max file size

By default, files larger than 1 MB are excluded from search results. Use the [search.largeFiles](../../admin/config/site_config.md#search-largeFiles) keyword to specify files to be indexed and searched regardless of size.

---

## Sourcegraph Cloud

[Sourcegraph Cloud](https://sourcegraph.com/search) is a public instance of Sourcegraph that lets you search inside any open-source project on GitHub. For demo purposes, you'll be prompted to narrow your query if it would search across more than 50 repositories. To lift this limitation or to search your organization's internal code, [run your own Sourcegraph instance](../../admin/install/index.md).
