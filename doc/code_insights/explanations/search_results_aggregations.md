## Search results aggregations

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

## Limitations

### Timeout limits

TODO

### Number of bars displayed

TODO

### Single capture group 

TODO 

### Saving aggregations to a code insights dashboard

Saving aggregations to a dashboard of code insights is not yet available. 

### Other (limitations perhaps for only certain search types)

TODO
