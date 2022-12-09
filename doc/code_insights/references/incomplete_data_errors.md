# Incomplete data errors

There are a few cases when a code insight may not return complete data. 

In all of these cases, if data is returned at all, it will be an undercount. 

See the below situations for tips on avoiding and troubleshooting these errors. 

## Timeout errors

For searches that take a long time to complete, it's possible for them to timeout before the search ends and we record the data value. 

To address this, try to minimize or avoid cases where your insight data series: 

1. Runs over extra large sets of repositories (scope your insight further to fewer repositories)
1. Uses many boolean combinations (consider splitting into multiple data series)
1. Runs a complex search over an especially large monorepo (consider optimizing your search query to be more specific or include more filters, like `lang:` or `path:`)

In addition: 

1. Timeout errors on points that have been backfilled before the creation date of the insight are more likely to occur on single, large repositories, because backfill points are calculated by running many searches, repository by repository, individually. 
1. Timeout errors on points that have been snapshot after the creation date of the insight are more likely to occur on insights running complex searches over large numbers of repositories, because snapshot points are calculated by running a single search against the current index 

## Searcher errors

[something about deployment time/when it's down? Tell them how to troubleshoot]

## Other errors

For other errors, please reach out to our support team. 