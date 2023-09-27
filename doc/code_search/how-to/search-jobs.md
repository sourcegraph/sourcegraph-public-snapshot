# Search Jobs <span class="badge badge-experimental" style="margin-left: 0.5rem; vertical-align:middle;">Experimental</span>

<p class="subtitle">Use Search Jobs to search code at scale for large-scale organizations.</p>

>NOTE: Search Jobs is available only on Enterprise accounts.

Search Jobs allows you to run search queries across your organization's codebase (all repositories, branches, and revisions) at scale. It enhances the existing Sourcegraph's search capabilities, enabling you to run searches without query timeouts or incomplete results.

With Search Jobs, you can start a search, let it run in the background, and then download the CSV file results from the Search Jobs UI when it's done. Site administrators can **enable** or **disable** the Search Jobs feature, making it accessible to all users on the Sourcegaph instance.

## Using Search Jobs

To use Search Jobs, you need to:

- Run a search query from your Sourcegraph instance
- Click the result menu below the search bar to see if your query is valid for the long search

![run-query-for-search-jobs](https://storage.googleapis.com/sourcegraph-assets/Docs/query-serach-jobs.png)

- If your query is valid, click **Run search job** to initiate the search job
- You will be redirected to the "Search Jobs UI" page at `/search-jobs`, where you can view all your created search jobs. If you're a site admin, you can also view search jobs from other users on the instance

![view-search-jobs](https://storage.googleapis.com/sourcegraph-assets/Docs/view-search-jobs.png)

## Limitations

The Search Job feature is not supported on queries that have:

- `OR` operator
- `has.content` or `has.file` predicates
- `*.` regexp search
- Multiple `rev` filters

>NOTE: Sourcegraph already offers an [Exhaustive Search](./../how-to/exhaustive.md) with the `count:all` operator. However, there are certain limitations when generating results within your codebase.
