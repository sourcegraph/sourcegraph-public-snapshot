# Features

## Powerful, flexible queries

Sourcegraph code search performs full-text searches and supports both regular expression and exact queries. By default, Sourcegraph searches across all your repositories. Our search query syntax allows for advanced queries, such as searching over any branch or commit, narrowing searches by programming language or file pattern, and more.

See the [query syntax](../reference/queries.md) and [query reference](../reference/language.md) documentation for a comprehensive overview of supported syntax.

## Language-aware structural code search

Sourcegraph supports advanced code search for specifically matching patterns inside code structures, like function parameters and loop bodies.

See the [structural search](../reference/structural.md) documentation for a detailed explanation of this search mode.

## Commit diff search

Search over commit diffs using `type:diff` to see how your codebase has changed over time. This is often used to find changes to particular functions, classes, or areas of the codebase when debugging.

You can also search within commit diffs on multiple branches by specifying the branches in a `repo:` field after the `@` sign. After the `@`, separate Git refs with `:`, specify Git ref globs by prefixing them with `*`, and exclude a commit reachable from a ref by prefixing it with `^`. Diff searches can be further narrowed down with parameters that filter by author and time.

See the [query syntax](../reference/queries.md#diff-and-commit-searches-only) documentation for a comprehensive list of supported parameters.

## Commit message search

Searching over commit messages is supported in Sourcegraph by adding `type:commit` to your search query. Separately, you can also use the `message:"any string"` parameter to filter `type:diff` searches for a given commit message. Commit message searches can narrowed down further with filters such as author and time.

See our [query syntax](../reference/queries.md#diff-and-commit-searches-only) documentation for a comprehensive list of supported parameters.

## Symbol search

Searching for symbols makes it easier to find specific functions, variables, and more. Use the `type:symbol` filter to search for symbol results. Symbol results also appear in typeahead suggestions, so you can jump directly to symbols by name. When on an [indexed](../../admin/search.md#indexed-search) commit, it uses Zoekt. Otherwise it uses the [symbols service](../../code_navigation/explanations/features.md#symbol-search)

## Saved searches

Saved searches let you save and describe search queries so you can easily monitor the results on an ongoing basis. You can create a saved search for anything, including diffs and commits across all branches of your repositories. Saved searches can be an early warning system for common problems in your code and a way to monitor best practices, the progress of refactors, etc.

See the [saved searches](../how-to/saved_searches.md) documentation for instructions for setting up and configuring saved searches.

## Suggestions

As you type a query, the menu below will contain suggestions based on the query. Use the keyboard or mouse to select a suggestion. For example, if your query is `repo:foo file:\.js$ hello`, the suggestions will consist of the list of files that match your query.

You can also type in the partial name of a repository or filename to quickly jump to it. For example, typing in just `foo` would show you a list of repositories (first) and files with names containing _foo_.

## Search contexts

Search contexts help you search the code you care about on Sourcegraph. A search context represents a set of repositories at specific revisions on a Sourcegraph instance that will be targeted by search queries by default.

Every search on Sourcegraph uses a search context. Search contexts can be defined with the contexts selector shown in the search input, or entered directly in a search query.

**Sourcegraph.com** supports a [set of predefined search contexts](https://sourcegraph.com/contexts?order=spec-asc&visible=17&owner=all) that include:

- The global context, `context:global`, which includes all repositories on Sourcegraph.com.
- Search contexts for various software communities like [CNCF](https://sourcegraph.com/search?q=context:CNCF), [crates.io](https://sourcegraph.com/search?q=context:crates.io), [JVM](https://sourcegraph.com/search?q=context:JVM), and more.  

If no search context is specified, `context:global` is used.

**Private Sourcegraph instances** support custom search contexts:

- Contexts owned by a user, such as `context:@username/context-name`, which can be private to the user or public to all users on the Sourcegraph instance.
- Contexts at the global level, such as `context:example-context`, which can be private to site admins or public to all users on the Sourcegraph instance.
- The global context, `context:global`, which includes all repositories on the Sourcegraph instance.

This feature is currently under active development for self-hosted Sourcegraph instances and is disabled by default.

Your site admin can enable search contexts on your private instance in **global settings** using the following:

```json
"experimentalFeatures": {  
  "showSearchContext": true,
  "showSearchContextManagement": true
}
```

## Multi-branch indexing <span class="badge badge-primary">experimental</span>

> NOTE: This feature is still in active development and must be enabled by a Sourcegraph site admin in site configuration.

The most common branch to search is your default branch. To speed up this common operation, Sourcegraph maintains an index of the source code on your default branch. Some organizations have other branches that are regularly searched. To speed up search for those branches, Sourcegraph can be configured to index up to 64 branches per repository.

Your site admin can configure indexed branches in site configuration under the `experimentalFeatures.search.index.branches` setting. For example:

``` json
"experimentalFeatures": {
  "search.index.branches": {
   "github.com/sourcegraph/sourcegraph": ["3.15", "develop"],
   "github.com/sourcegraph/src-cli": ["next"]
  }
}
```

Indexing multiple branches will add additional resource requirements to Sourcegraph (particularly memory). The indexer will deduplicate documents between branches. So the size of your index will grow in relation to the number of unique documents. Refer to our [resource estimator](../../../admin/deploy/resource_estimator.md) to estimate whether additional resources are required.

> NOTE: The default branch (`HEAD`) is always indexed.

> NOTE: All revisions specified in version contexts are also indexed.

**Note**: While version contexts are located in the site configuration, search contexts are located in the global settings.

Reload the page after saving changes to see search contexts enabled.

If you currently use [version contexts](#version-contexts), you can automatically [convert your existing version contexts to search contexts](../../admin/how-to/converting-version-contexts-to-search-contexts.md). We recommend migrating to search contexts for a more intuitive, powerful search experience and the latest improvements and updates.

See the [search contexts](../how-to/search_contexts.md) documentation to learn how to use and create search contexts.
