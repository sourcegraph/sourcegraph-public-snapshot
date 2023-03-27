# GitHub API Oddities

This document serves as a catalogue of known oddities when it comes to GitHub's API.

## The search API

We use the GitHub search API when a GitHub connection is created using the `repositoryQuery` field. However, the GitHub search API comes with some limitations.

### Repository count limit

You cannot view more than the first 1,000 results of a GitHub search. If you perform a search, like `stars:>=5` that returns 3,000,000+ results, you will only be able to paginate through the first 1,000 results.

To overcome this, we recursively narrow our search using the `created:` parameter. We halve the time frame each time until we find a time window in which no more than 1,000 results are returned.

### Search result ordering inconsistencies

Search results aren't guaranteed to be ordered consistently. For one, the repository data can change while we're paginating through the results. If a repository is moved down to the first page, after we have already fetched the first page, then we will miss that repository while paginating over the results.

There is no elegant way around this. The `created` filter helps narrow this problem, but it does not eliminate it. Search results also aren't guaranteed to be in the same order, even if they stay the same. For example, two repositories with the same number of stars, when sorting with `sort:stars-asc`, can swap places in between searches.

GitHub offers the following sort options:
- Updated
- Stars
- Number of "Help Wanted" issues
- "Best Match"

None of these options are consistent, and all of them can change while paginating through results.

What this effectively means is that, when syncing a large number of repositories (the more pages, the worse it gets), successive syncs will find repositories it previously missed, and miss repositories it previously found, even though those repositories should all still be part of the search.

As of yet, there does not seem to be a way around this that is not extremely inefficient.
