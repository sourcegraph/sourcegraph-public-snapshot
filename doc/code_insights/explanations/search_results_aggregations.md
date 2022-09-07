# Search results aggregations

In version 4.0 and later, Code Insights provides aggregations shown on the search screen.

## Available aggregations: 

You can aggregate any search by: 

1. The repositories with search results
1. The files with search results (for non-commit and non-diff searches)
1. The authors who created the search results (for commit and diff searches)
1. All found matches for the first capture group pattern (for regexp searches with a capture group)

Aggregations are returned in order of greatest to least results count. 

Aggregations are exhaustive across all repositories the user running the search has access to, unless the chart notes otherwise (see [Limitations](#limitations) below). 

We may continue adding new aggregation categories, like date and code host, based on feedback. If there are categories you'd like to see, please [let us know](mailto:feedback@sourcegraph.com).

## Feature visibility

You can turn the aggregations on with the experimental feature setting: `enableSearchResultsAggregations`

You can turn off just the proactive aggregations with the setting: `disableProactiveSearchAggregations`

## Drilldowns 

You can drilldown into a search aggregation by clicking a result in the chart. Your original search query will be updated with a `repo`, `file`, `author` filter or a regexp pattern depending on the aggregation mode.

## Limitations

### Mode limitations

If you attempt to run a query for which a given mode is not supported, the tooltip will inform you why that mode is not available. 

### Timeout limits

At the moment all aggregations search queries are run with a 2-second timeout, even if your search specified a timeout. If the aggregation times out, you will be able to trigger a longer search with a 1-minute timeout by clicking `Run aggregation`. 

### Count limits

At the moment all aggregations search queries are run with `count:all`, even if your search specified a count. 

### Best effort aggregation

Results are aggregated in a best-effort approach using a limited-size buffer to hold group labels in order not to strain the webapp when these aggregations are run. 
This means that in some cases we might miss some high-count results. 
Aggregations that hit such non-exhaustive paths are reported back to the user.

You can control the size of the buffer using the site setting `insights.aggregations.bufferSize`. It is set to 500 by default. Note that if increasing this you might notice decreased performance on your instance.

### Number of bars displayed

The side panel will display a maximum of 10 bars. If expanded, a maximum of 30 bars will be displayed. If there are more results this will be displayed on the panel.

### Single capture group 

Aggregations by capture group will match on the first capture group in the search query only. For example, for a query:

```regexp
hello-(\w+)-(\w+)
```

and a match like `hello-beautiful-world` only `beautiful` will be shown as a result.

### Saving aggregations to a code insights dashboard

Saving aggregations to a dashboard of code insights is not yet available. 

### Slow diff and commit queries

Running aggregations by author is only allowed for `type:diff` and `type:commit` queries, which are likely not to complete within a 2-second timeout.
You can trigger an explicit search with an extended 1-minute timeout, or you can limit your query using a single-repo filter (like `repo:^github\.com/sourcegraph/sourcegraph$`) combined with a `before` or `after` filter.

### Structural searches

Aggregations for structural searches are unlikely to complete within a 2-second timeout. You can try to trigger an explicit aggregation for such cases.

### Standard searches with embedded regexp

Standard searches with embedded regexp such as below do not support aggregation by capture group. This is because they are functionally similar to a query with an `or` operator.
```regexp
database /(\w+)/
```
