# Search Jobs <span class="badge badge-experimental" style="margin-left: 0.5rem; vertical-align:middle;">Experimental</span>

<p class="subtitle">Use Search Jobs to search code at scale for large-scale organizations.</p>

>NOTE: Search Jobs is available only on Enterprise accounts.

Search Jobs allows you to run search queries across your organization's codebase (all repositories, branches, and revisions) at scale. It enhances the existing Sourcegraph's search capabilities, enabling you to run searches without query timeouts or incomplete results.

With Search Jobs, you can start a search, let it run in the background, and then download the CSV file results from the Search Jobs UI when it's done. Site administrators can **enable** or **disable** the Search Jobs feature, making it accessible to all users on the Sourcegaph instance.

## Enable Search Jobs

To enable Search Jobs:

- Login to your Sourcegraph instance and go to the site admin
- Next, click the site configuration
- From here, you'll see `experimentalFeatures`
- Set `searchJobs` to `true` and then refresh the page

## Using Search Jobs

To use Search Jobs, you need to:

- Run a search query from your Sourcegraph instance
- Click the result menu below the search bar to see if your query is valid for the long search

![run-query-for-search-jobs](https://storage.googleapis.com/sourcegraph-assets/Docs/query-serach-jobs.png)

- If your query is valid, click **Run search job** to initiate the search job
- You will be redirected to the "Search Jobs UI" page at `/search-jobs`, where you can view all your created search jobs. If you're a site admin, you can also view search jobs from other users on the instance

![view-search-jobs](https://storage.googleapis.com/sourcegraph-assets/Docs/view-search-jobs.png)

## Limitations

Search Jobs supports queries of `type:file` and it automatically appends this to the search query. Other result types (like `diff`, `commit`, `path`, and `repo`) will be ignored. However, there are some limitations on the supported query syntax. These include:

- `OR`, `AND` operators
- file predicates, such as `file:has.content`, `file:has.owner`, `file:has.contributor`, `file:contains.content`
- `.*` regexp search
- Multiple `rev` filters
- Queries with `index: filter`

>NOTE: Sourcegraph already offers an [Exhaustive Search](./../how-to/exhaustive.md) with the `count:all` operator. However, there are certain limitations when generating results within your codebase.
