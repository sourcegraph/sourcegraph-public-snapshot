# Filtering a code insight

This how-to assumes that you already have created some search insights.
If you have yet to create any code insights, start with the [quickstart](../quickstart.md) guide.

### 1. Click the filter icon button for the insight you want to filter

While on a dashboard page with the insight you want to filter in front of you, find the filter icon in the top right and click it.
It will open the filter popover.

### 2. Include or exclude repositories using a regular expression

The filter popover gives you two inputs to filter the insight to a subset of repositories either by inclusion or by exclusion.
If you specify an inclusion pattern, only repositories that have a repository name matching the regular expression will be counted.
If you specify an exclusion pattern, only repositories that have a repository name **not** matching the regular expression will be counted.
You can also combine both filters, in which case the inclusion pattern will be applied first, then the exclusion pattern.

The repository name is the part of the URL when navigating to a repository that follows your instance domain.
The regular expression to filter for a repository works the same as the [`repo:` filter in the search box](../../code_search/reference/queries#keywords-all-searches).

Examples:

| Pattern | Explanation |
|---------|-------------|
| `^github\.com/sourcegraph/sourcegraph$` | Include or exclude the specific repository `github.com/sourcegraph/sourcegraph` |
| `^github\.com/sourcegraph/(sourcegraph\|about\|docsite)$` | Include or exclude the specific repositories `github.com/sourcegraph/sourcegraph`, `github.com/sourcegraph/about` and `github.com/sourcegraph/docsite` |
| `^github\.com/sourcegraph/go-` | Include or exclude all repositories with the prefix `github.com/sourcegraph/go-` |
| `service` | Include or exclude all repositories that contain the word `service` in their name |
| `\.js$` | Include or exclude all repositories that end in `.js` |

You can savely click outside the panel to close it and your filters will still be applied (until a page reload).
The little dot next to the filters icon indicates that an insight currently has filters applied for you.

### 5. Edit or reset filters

Bring up the filters panel at any time for an insight by clicking the filters button again to edit them or resetting them with the "Reset" buttons.

### 3. Update default filters (optional)

By default, the filter will only be visible to you and reset after a page reload.
To persist the filter and apply it for other viewers viewing the insight too, click "Update default filters".

### 4. Save as new view (optional)

If you'd like to not modify the insight you're looking at for everyone, but rather fork the insight into a second one with the filters applied for everyone to see an compare to, you can click the button "Save as new view".
The view appears as a separate chart on the dashboard, with default filters applied, while the original insight stays as-is.

