# Code Insights filters

Code insights filters allow you to filter a chart of search insight to a subset of repositories at the time of viewing it.
Filters are applied immediately without having to recompute the insight and can be reset at any time.
The filter panel is available through the filter icon in the top right of every search insight.

## Filter options

Repositories are included or excluded from the insight using regular expressions, similar to the `repo:` filter in [Sourcegraph searches](../../code_search/reference/queries#keywords-all-searches).

## Persistance and sharing

Filters you specify are temporary by default and only seen by you.
Unless you click "Update default filters" or "Save as new view", the filters will not apply to the chart other viewers of the dashboard see and will disappear after a page reload.

### Default filters

By clicking "Update default filters", the filters you entered will be attached to the insight, persist even after a page reload, and be shown to every viewer of the insight/dashboard.
Viewers can locally modify filters to their liking without affecting what other viewers see, until they click "Update default filters".

Insights that have default filters applied will be indicated by a small dot at the filter icon. The filter can be reset by opening the filter panel again and clicking "Reset" above each filter or "Reset all filters".

### Insight views

By clicking "Save as new view", the insight will be forked into a new insight with the specified filters applied.
The view appears as a separate chart on the dashboard.
Filters for the newly created view and the original insight can be edited and persisted through "Update default insights" independently.
Editing the forked insight (e.g. changing the title) will not change the original insight.
