# Current limitations of Code Insights (Beta limitations)

Because code insights is currently a private [beta feature](../../admin/beta_and_experimental_features.md#beta-features), there are some limitations that we have not finished building solutions for yet. 

If you have strong feedback, please do [let us know](mailto:feedback@sourcegraph.com). 

_Limitations listed here are relevant for Sourcegraph 3.31. Limitations that are no longer current are [documented at the bottom](#older-versions-limitations) for the benefit of customers who have not yet upgraded._

## Performance speed considerations for a data series running over all repositories

To accurately return historical data for insights running over all of your repositories, the backend service must run a large number of Sourcegraph searches. This means that unlike code insights running over just a few repositories, results are not returned instantly, but more often on the scale of 20-120 minutes, depending on:

* _N_: how many repositories you have connected to your instance; in our tests, we used 26,400 repositories
* _q_: the performance and resources of your Sourcegraph code insights instance in queries-per-second; in our tests, 7 queries per second was average
* _c_: how well we can "compress" repositories so we don't need to re-run queries every month (e.g., if a repository hasn't changed in two months); in our tests, C = ~2

A _very_ general formula for estimating how long an individual data series (query) will take to run on your instance in seconds  _N_ * 1/_c_ * 1/_q_. 

On our test instance, we find a code insight data series takes approximately:

26,400 repositories * 1/2 compression factor * 1/7 queries per second = 31 minutes

The number of insights you have does not affect the overall speed at which they run: it will take the same total time to run all of them whether or not you let each one finish before creating the next one. Insights currently [populate in parallel](https://github.com/sourcegraph/sourcegraph/pull/23101), prioritizing most-recent-in-time datapoints first. 

> NOTE: we have many performance improvements planned. We'll likely release considerable performance gains in the upcoming releases of 2021. 

## Feature parity limitations 

### Features currently available only on insights over all your repositories

* **[Filtering insights](code_insights_filters.md)**: we do not yet allow filtering for insights that run over explicitly defined lists of repositories. (If you want to filter those insights' repository lists, you can quickly add/remove repositories on the edit screen and results will return equally quickly as filtering an insight over all repositories.) 

### Features currently available only on insights over explicitly defined repository lists

Because these insights need to run dramatically fewer queries than insights over thousands of repositories, you will have access to a number of features not _yet_ supported for insights over all repositories. These are: 

* **Dynamic x-axis ranges**: set a custom amount of historical data you care about
* **Editing data series queries after creation**: for insights over all repositories, you must make a new insight if you wish to run a different query
* **Live previews**: showing the preview of your insight in real time
* **"Diff click"**: click a datapoint on your insight and be taken to a diff search showing any changes contributing to the difference between a datapoint and the prior one

> NOTE: many of the above-listed features will become available for insights over all repositories as well. The above list is ordererd top-down, where items on the top of the list will arrive roughly sooner than items on the bottom. 

## In certain cases, chart datapoints don't match the result count of a Sourcegraph search

There are currently a few subtle differences in how code insights and Sourcegraph web app searches handle defaults when searching over all repositories. Refer to [Common reasons code insights may not match search results](../references/common_reasons_code_insights_may_not_match_search_results.md). 

## Known bugs

Known bugs we plan to fix are tracked in our [GitHub repository here](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Abug+label%3Ateam%2Fcode-insights). 

## Older versions' limitations

### Version 3.30 (July 2021) or older

#### Search-based Code Insights can only run over ~50-70 repositories 

Because this version of the prototype runs on frontend API calls to Sourcegraph searches, it may run slowly (or possibly timeout) if you're using it over many repositories or with many data series for each insight. 

#### The max match count is 5,000 matches per repository 

The current limit on searching over historical versions of repositories, which is an unindexed search, is 5,000 results per repository. If there are more than 5,000 matches, the search stops and returns a count of 5,000, and the code insight graph will calculate the overall chart using 5,000 as the match count for that repository. (This means if you query over two repositories and one of them hits this limit, the value shown on the graph will be 5,000 + [the match count in the other repository]). 

_This limit was lifted in the August 2021 release of Sourcegraph 3.31_


