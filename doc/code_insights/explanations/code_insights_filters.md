# Code Insights filters

Code insights filters allow you to filter a search insight to a subset of repositories. Each search insight has a filter icon that opens the filtering options. 

Filters take effect immediately, without needing to re-process the data series of the insight.

> NOTE: filters are not yet available on language statistics insights. 

## Filter options

### `repo:` filters

You can include or exclude repositories from your insight using regular expressions, the same way you can with the `repo:` filter in [Sourcegraph searches](../../code_search/reference/queries.md#keywords-all-searches).

For inclusion, only repositories that have a repository name matching the regular expression will be counted.

For exclusion, only repositories that have a repository name **not** matching the regular expression will be counted.

If you combine both filters, the inclusion pattern will be applied first, then the exclusion pattern.

### Other filtering options

We're currently exploring additional filters that would be valuable. If you have feedback about a particular filter you'd like for code insights, we would [love to hear your feedback](mailto:feedback@sourcegraph.com).

## Filter persistance and sharing

### Filters are temporary by default

By default, filters are temporary (present until you refresh the page or switch to a different dashboard) and local (no one else can see them). You can modify filters without affecting what others see, unless or until you save the filters.

### Saving filters as defaults

You can set filters to be persistent, or default, even after a page reload on a code insights in two ways:

1. Create a filter and click "save/update default filters": this will save the filters so they persist for all viewers of this insight, on any dashboard page the insight appears. 
1. Create a filter and "save as new view": this will create a new insight chart with your filter applied as the default filter on that insight. It will _not_ also save these filters to the existing insight unless you also select "save/update default filters". 

### Filter indicator 

Insights that have any filters applied will have a small dot on their filter icon. The filter can always be reset by opening the filter panel again and clicking "Reset" above each filter – but note that, like any filter edit, you must also then "update default filters" to save that reset state if you want it to persist. 

### Saving a filter as a new view

When you create a filter and "save as new view," it will create a new insight chart with this filter saved as the default filteres. Except the title and filters, all other configuration options will be cloned from the original insight. 

Filters and all other configuration options for the newly created view and the original insight are indpendent. Editing the forked insight (e.g. changing the data series query) will not change the original insight.
