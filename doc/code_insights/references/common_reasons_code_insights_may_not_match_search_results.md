# Common reasons code insights may not match search results

There are a few reasons why chart data series' most recent datapoint may show you a different number of match counts than the same search query run in Sourcegraph manually. 

## If the chart data point shows *higher* counts than a manual search

### [For versions pre-3.40] Not including `fork:no` and `archived:no` in your insight query

Because code insights historical search defaults to `fork:yes` and `archived:yes`, but a Sourcegraph search via the web interface or CLI does not, it may be that your insight data series is including results from repositories that are excluded from a Sourcegraph search. Try running the same search again manually with `fork:yes` and `archived:yes` filters. 

> NOTE: 3.40+ version defaults to `fork:no` and `archived:no`, the same way the search UI does.

### Manual search will not include unindexed repositories

All repositories in a historical search are unindexed, but a manual Sourcegraph search only includes indexed repositories. It's possible your manual searches are missing results from unindexed repositories. 

To investigate this, one can compare the list of repositories in the manual search (use a `select:repository` filter) with the list of repositories in the insight `series_points` database table. To see why a repository may not be indexing, refer to [this guide](../../admin/troubleshooting.md#sourcegraph-is-not-returning-results-from-a-repository-unless-repo-is-included). 

## If the chart data point shows *lower* counts than a manual search 

### New matches created since the insight datapoint ran

Currently, a data series' most recent datapoint defaults to the end of the prior month. It's possible that in the time between when your insight ran and when you ran a manual search, new matches have been added to your codebase. To confirm this, you can run `type:diff` or `type:commit` searches using the `after:` filter, but note that those filters only support up to 10,000 repositories, so you may first need to limit your search repository set. 

> NOTE: Future releases of Code Insights may include [an always-up-to-date present-time point](https://github.com/sourcegraph/sourcegraph/issues/24186).

### Repository timeouts caused a datapoint to miss results

If your code insight is very large, it is possible that a few (\<1% in 100+ manual tests over 26,000 repositories) repositories failed to return match counts due to timing out while searching. To check this, you can run the following GraphQL query in the Sourcegraph GraphQL API: 
```graphql
{
  insights(ids: ["INSIGHT_ID"]) {
    totalCount
    nodes {
      id
      title
      description
      series{
        label
        status {
          pendingJobs
          completedJobs
          failedJobs
        }
        dirtyMetadata {
          time
          count
          reason
        }
      }
      id
    }
  }
}
```

where `INSIGHT_ID` can be found in the URL of the "edit" page for the insight (selectable from the three-dot dropdown on the insight) after `...edit/`. For example, `https://sourcegraph.yourdomain.com/insights/edit/searchInsights.insight.TestInsightTitleHere?dashboardId=all` contains the `INSIGHT_ID` of `searchInsights.insight.TestInsightTitleHere`. 

If there are `failedJobs`, there may be timeouts or similar issues affecting your insight. 
