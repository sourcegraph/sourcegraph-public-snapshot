# Current limitations of Code Insights

Because code insights is currently a [prototype feature](../../admin/beta_and_prototype_features.md#prototype-features), there are some limitations that we have not finished building solutions for yet. 

If you have strong feedback, please do [let us know](mailto:feedback@sourcegraph.com).

## The number of repositories in a search-based insight must be less than 50

Code Insights currently depends on frontend API calls to Sourcegraph searches, so it can't run historical searches over more than 50 repositories due to that search limitation. 

We're currently developing a scalable backend service that fixes this and are planning to release it by September 2021. 

## Search-based Code Insights may run slowly for large numbers of repositories or data series

Because Code Insights runs on the frontend, it may run slowly (or possibly timeout) if you're using it over very many repositories or with very many data series for each insight. That said, a code insight caches locally after you've run it the first time. 

We're currently working on both backend and frontend improvements to increase the speed at which code insights returns data, and you can expect significant improvement by September 2021. 

## Known bugs

Known bugs we plan to fix are tracked in our [GitHub repository here](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Abug+label%3Ateam%2Fcode-insights). 
