# Quickstart for Code Insights [Prototype]

> ## _This is a prototype_ (Considerations)
>
> This is a prototype, and it's not yet optimized. 
>
> For example, a search insight runs all the search queries on the frontend and tries to cache them – it may take awhile on first view but less time on subsequent views. The number of queries to run is (number of repos in your repo list) x (number of steps). The number of steps is currently hardcoded to 7. If you're running into loading time issues, first try limiting the number of repos.
>
> Please do slack/email Joel Kwartler (@joel slack; @joelkw github; joel@sourcegraph.com) with feedback, bugs, and feature requests. 



## 1. Enable the experimental feature flag

Add the following to your Sourcegraph user `/users/[username]/settings` or organization `/organizations/[your_org]/settings` settings: 

```jsx
"experimentalFeatures": { "codeInsights": true }
```

## 2. Enable the extension
 
Visit the [search insights extension page](https://sourcegraph.com/extensions/sourcegraph/search-insights) and enable it.

## 3. Configure your first insight

Configure the insight by adding the following structure to your user, org, or global settings: 

```jsx
// Choose any name - only the prefix "searchInsights.insight." is mandatory.
"searchInsights.insight.uniqueNameForYourInsight": {

  // Shown as the title of the insight.
  "title": "Migration to React function components",

  // a list of repositories to run the search over
  "repositories": ["github.com/sourcegraph/sourcegraph"],

  // The lines of the chart.
  "series": [
    {
      // Name shown in the legend and tooltip.
      "name": "Function components",

      // The search query that will be run for the interval defined in "step".
      // Do not include the "repo:" filter as it will be added automatically for the current repository. Example query
      "query": "patternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent",

      // An optional color for the line.
      // Can be any hex color code, rgb(), hsl(),
      // or reference color variables from the OpenColor palette (recommended): https://yeun.github.io/open-color/
      // The semantic colors var(--danger), var(--warning) and var(--info) are also available.
      "stroke": "var(--oc-teal-7)"
    },
    {
      // this is a second series on the same chart 
      "name": "Class components",
      "query": "patternType:regexp extends\\s(React\\.)?(Pure)?Component",
      "stroke": "var(--oc-indigo-7)"
    }
  ],
  // The step between two data points. Supports days, months, hours, etc.
  "step": {
    "weeks": 2
  }
}
```

## 4. View your insight

The insight will appear on your insights page (`YourSourcegraphUrl.com/insights`) as well as below the search bar on the home page. 

Regardless of whether you define a repo, the insight will show up on every directory page (and run for that directory). With repos listed, it will also show up on the /insights and search pages. Without listing repos, it will show nothing on those two pages. 

If you put these settings in your org settings, everyone will see them on the /insights page. If you put them in your personal settings, only you will see them there. 

## 5. Configure additional insights

Multiple search insights can be defined by adding more objects under different keys, as long as they are prefixed with `searchInsights.insight.` .

## 6. Other insights

Sourcegraph also has [language usage insights](https://sourcegraph.com/extensions/sourcegraph/code-stats-insights), which adds a pie chart of lines of code by language to your directory and insights page. 

The setup is similar, and you enable [the extension](https://sourcegraph.com/extensions/sourcegraph/code-stats-insights) and then add to your user or global settings: 

```json
// Title
"codeStatsInsights.title": "Language usage",

// The file query to limit the global insights page to (here it's limited to a repo)
"codeStatsInsights.query": "repo:^github\\.com/sourcegraph/sourcegraph$",

// The threshold for grouping all other languages into an "other" category
"codeStatsInsights.otherThreshold": 0.02,
```
> Please do slack/email Joel Kwartler (@joel slack; @joelkw github; joel@sourcegraph.com) with any feedback, bugs, and feature requests! 