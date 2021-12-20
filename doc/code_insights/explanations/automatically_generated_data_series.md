# Detect and track patterns with automatically generated data series 

> Note: "Detect and track patterns" insights are available in Sourcegraph 3.35 and above. In Sourcegraph 3.35 this feature is in its earliest version. Stability and additional features coming in following releases.

Code Insights with automatically generated data series allow you to "detect and track patterns" and versions within your codebase. You can use these insights to track versions of languages, packages, terraform, docker images, or anything else that can be captured with a [regular expression capture group](#regular-expression-capture-group-resources).

## Automatic generation of individual series

Match patterns can be returned with a regular expression containing a capture group, or a pattern within parentheses. 

For example, `file:\.pom$ <java\.version>(.*)</java\.version>` will match all of: 

```
<java.version>1.5</java.version>
<java.version>1.7</java.version>
<java.version>1.8</java.version>
```

Code Insights will find all matches, and then automatically generate a data series and color for each unique value of the capture group. In this case, the chart would show data series for 1.5, 1.7, and 1.8, with the values being the number of matches for each unique value. 

## New matching data gets automatically added 

Capture groups will automatically create new data series for new matches as they appear in your codebase. You do not need to update or manually re-create the insights to track newly added versions or patterns.

For the above example, this means that if `<java.version>1.9</java.version>` was committed to the codebase in the future, it would appear on the insight without any additoinal action, and you would see a series for `1.9`. 

## Current limitations 

Code Insights is in Beta and this feature has some yet-released limitations. In rough order, with limitations listed first likely to be removed soonest, they are: 

### Must specify repository list 

Code Insights using capture groups only run over an explicit list of repositories (max 50-70) rather than running over all connected repositories. 

### Limited to 20 matches

Capture groups will only display 20 returned match values to prevent extremely large result sets from being rendered. If there are more than 20 matches, **a non-deterministic 20 series will be rendered on every reload**. Until we add controls to make this deterministic and configurable, for now we strongly recommend you refine your regular expression so that there are no more than 20 different values returned. 

### No capture groups in filter strings 

This type of Code Insight only supports capture groups in the main query string. You cannot use a capture group in a filter keywork, like a `repo:` or `file:` filter. 

### No combinations of capture groups 

You cannot combinatorially combine capture groups. Queries containing multiple capture groups will return data series with a label for a single capture group only. For example, `([0-9])\.([0-9])\.([0-9])` will match results like 1.2.3, 4.5.6, and 4.5.7, but will return data series for 1, 2, 3, 4, 5, 6, and 7 individually. (If you made the string one capture group `([0-9]\.[0-9]\.[0-9])` instead, it would return series and counts for 1.2.3, 4.5.6, and 4.5.7). 

### No commit or diff search 

You can't use capture groups in queries for `type:commit` or `type:diff`. 

### Regular expression capture group resources

- [Example regular expressions for common use cases](../references/common_use_cases.md#automatic-version-and-pattern-tracking)
- [General additional capture groups documentation](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Regular_Expressions/Groups_and_Ranges)