# Current limitations of Code Insights

Because code insights is currently a [prototype feature](../../admin/beta_and_prototype_features.md#prototype-features), there are some limitations that we have not finished building solutions for yet. 

If you have strong feedback, please do [let us know](mailto:feedback@sourcegraph.com).

## The runtime of search-based Code Insights degrades quickly after ~50 repositories

Because the Code Insights prototype currently runs on frontend API calls to Sourcegraph searches, it may run slowly (or possibly timeout) if you're using it over many repositories or with many data series for each insight. That said, a code insight caches locally after you've run it the first time.

We're currently developing a scalable backend service that fixes this limitation, with a planned release by September 2021, that will allow you to run code insights over thousands of repositories at once. As of now, performance generally gets noticeably slower around 50 repositories, and becomes functionally unusable above 200 repositories.

> Note: if your data series query is a `diff` search, there is an additional hard limit of 50 repositories. This limit will also be lifted as the product matures. 

## Known bugs

Known bugs we plan to fix are tracked in our [GitHub repository here](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Abug+label%3Ateam%2Fcode-insights). 
