# Pings

Sourcegraph periodically sends a ping to Sourcegraph.com to help our product and customer teams. It sends only the high-level data below. It never sends code, repository names, usernames, or any other specific data. To learn more, go to the **Site admin > Pings** page on your instance. (The URL is `https://sourcegraph.example.com/site-admin/pings`.)

## Critical telemetry

Critical telemetry includes only the high-level data below required for billing, support, updates, and security notices.

- Randomly generated site identifier
- The email address of the initial site installer (or if deleted, the first active site admin), to know who to contact regarding sales, product updates, security updates, and policy updates
- Sourcegraph version string (e.g. "vX.X.X")
- Dependency versions (e.g. "6.0.9" for Redis, or "13.0" for Postgres)
- Deployment type (single Docker image, Docker Compose, Kubernetes cluster, Helm, or pure Docker cluster)
- License key associated with your Sourcegraph subscription
- Aggregate count of current monthly users
- Total count of existing user accounts
- Aggregated repository statistics
  - Total size of git repositories stored in bytes
  - Total number of lines of code stored in text search index
- Code Insights: total count of insights 

## Other telemetry

By default, Sourcegraph also aggregates usage and performance metrics for some product features. No personal or specific information is ever included.

This telemetry can be disabled using the `disableNonCriticalTelemetry` option in
[site configuration](config/site_config.md#disableNonCriticalTelemetry).

- Whether the instance is deployed on localhost (true/false)
- Which category of authentication provider is in use (built-in, OpenID Connect, an HTTP proxy, SAML, GitHub, GitLab)
- Which code hosts are in use (GitHub, Bitbucket Server / Bitbucket Data Center, GitLab, Phabricator, Gitolite, AWS CodeCommit, Other)
  - Which versions of the code hosts are used
- Whether new user signup is allowed (true/false)
- Whether a repository has ever been added (true/false)
- Whether a code search has ever been executed (true/false)
- Whether code intelligence has ever been used (true/false)
- Aggregate counts of current daily, weekly, and monthly users
- Aggregate counts of current daily, weekly, and monthly users, by whether they are using code host integrations
- Aggregate daily, weekly, and monthly latencies (in ms) of search queries
- Aggregate daily, weekly, and monthly integer counts of the following query syntax:
  - The number of boolean operators (`and`, `or`, `not` keywords)
  - The number of built-in predicate keywords (`contains`, `contains.file`, `contains.repo`, `contains.commit.after`, `dependencies`)
  - The number of `select` keywords by kind (`repo`, `file`, `content`, `symbol`, `commit.diff.added`, `commit.diff.removed`)
  - The number of queries using the `context:` filter without the default `global` value
  - The number of queries with only patterns (e.g., without filters like `repo:` or `file:`)
  - The number of queries with three or more patterns
- Aggregate daily, weekly, and monthly user counts of search queries with the above properties
- Code intelligence usage data
  - Total number of repositories with and without an uploaded LSIF index
  - Total number of code intelligence queries (e.g., hover tooltips) per week grouped by language
  - Number of users performing code intelligence queries (e.g., hover tooltips) per week grouped by language
<!-- depends-on-source: ~/internal/usagestats/batches.go -->
- Batch Changes usage data
  - Total count of page views on the batch change apply page
  - Total count of page views on the batch change details page after creating a batch change
  - Total count of page views on the batch change details page after updating a batch change
  - Total count of created changeset specs
  - Total count of created batch change specs
  - Total count of created batch changes
  - Total count of closed batch changes
  - Total count of changesets created by batch changes
  - Aggregate counts of lines changed, added, deleted in all changesets
  - Total count of changesets created by batch changes that have been merged
  - Aggregate counts of lines changed, added, deleted in all merged changesets
  - Total count of changesets manually added to a batch change
  - Total count of changesets manually added to a batch change that have been merged
  - Aggregate counts of unique monthly users, by:
      - Whether they have contributed to batch changes
      - Whether they only viewed batch changes
      - Whether they have performed a bulk operation
  - Weekly batch change (open, closed) and changesets counts (imported, published, unpublished, open, draft, merged, closed) for batch change cohorts created in the last 12 months
  - Weekly bulk operations count (grouped by operation)
  - Total count of executors connected
  - Cumulative executor runtime monthly
  - Total count of `publish` bulk operation
  - Total count of bulk operations (grouped by operation type)
  - Changeset distribution for batch change (grouped by batch change source: `local` or `executor`)
  - Total count of users that ran a job on an executor monthly
  - Total count of published changesets and batch changes created via:
      - executor
      - local (using `src-cli`)
- Aggregated counts of users created, deleted, retained, resurrected and churned within the month
- Saved searches usage data
  - Count of saved searches
  - Count of users using saved searches
  - Count of notifications triggered
  - Count of notifications clicked
  - Count of saved search views
- Homepage panel engagement
  - Percentage of panel clicks (out of total views)
  - Total count of unique users engaging with the panels
- Weekly retention rates for user cohorts created in the last 12 weeks
- Search onboarding engagement
  - Total number of views of the onboarding tour
  - Total number of views of each step in the onboarding tour
  - Total number of tours closed
- Sourcegraph extension activation statistics
  - Total number of users that use a given non-default Sourcegraph extension
  - Average number of activations for users that use a given non-default Sourcegraph extension
  - Total number of users that use non-default Sourcegraph extensions
  - Average number of non-default extensions enabled for users that use non-default Sourcegraph extensions
- Code insights usage data
  - Total count of page views on the insights page
  - Count of unique viewers on the insights page
  - Total counts of hovers, clicks, and drags of insights by type (e.g. search, code stats)
  - Total counts of edits, additions, and removals of insights by type
  - Total count of clicks on the "Add more insights" and "Configure insights" buttons on the insights page
  - Weekly count of users that have created an insight, and count of users that have created their first insight this week                  
  - Weekly count of total and unique views to the `Create new insight`, `Create search insight`, and `Create language insight` pages
  - Weekly count of total and unique clicks of the `Create search insight`, `Create language usage insight`, and `Explore the extensions` buttons on the `Create new insight` page
  - Weekly count of total and unique clicks of the `Create` and `Cancel` buttons on the `Create search insight` and `Create language insight` pages
  - Total count of insights grouped by time interval (step size) in days  
  - Total count of insights set organization visible grouped by insight type
  - Total count of insights grouped by presentation type, series type, and presentation-series type.
  - Weekly count of unique users that have viewed code insights in-product landing page
  - Weekly count of per user changes that have been made over the query field insight example.
  - Weekly count of per user changes that have been made over the repositories field insight example.
  - Weekly count of clicks on the "Create your first insight" CTA button on the in-product landing page.
  - Weekly count of clicks on the code insights in-product template section's tabs.
  - Weekly count of clicks on the use/explore template card's button.
  - Weekly count of clicks on the "view more" template section button.
  - Weekly count of clicks on the in-product landing page documentation links.
- Code monitoring usage data
  - Total number of views of the code monitoring page
  - Total number of views of the create code monitor page
  - Total number of views of the create code monitor page with a pre-populated trigger query
  - Total number of views of the create code monitor page without a pre-populated trigger query
  - Total number of views of the manage code monitor page
  - Total number of clicks on the code monitor email search link
- Notebooks usage data
  - Total number of views of the notebook page
  - Total number of views of the notebooks list page
  - Total number of views of the embedded notebook page
  - Total number of created notebooks
  - Total number of added notebook stars
  - Total number of added notebook markdown blocks
  - Total number of added notebook query blocks
  - Total number of added notebook file blocks
  - Total number of added notebook symbol blocks
  - Total number of added notebook compute blocks
- Code Host integration usage data (Browser extension / Native Integration)
  - Aggregate counts of current daily, weekly, and monthly unique users and total events
  - Aggregate counts of current daily, weekly, and monthly unique users and total events who visited Sourcegraph instance from browser extension
- IDE extensions usage data
  - Aggregate counts of current daily, weekly, and monthly searches performed:
    - Count of unique users who performed searches
    - Count of total searches performed
  - Aggregate counts of current daily user state:
    - Count of users who installed the extension
    - Count of users who uninstalled the extension 
  - Aggregate count of current daily redirects from extension to Sourcegraph instance



- CTA usage data
  - Browser extension
    - Total number of users who viewed the "install browser extension" CTA on the file page
    - Total number of users who clicked the "install browser extension" CTA on the file page
    - Total number of users who viewed the "install browser extension" CTA on the search page
    - Total number of users who clicked the "install browser extension" CTA on the search page
    - Total number of views of the "install browser extension" CTA on the file page
    - Total number of clicks on the "install browser extension" CTA on the file page
    - Total number of views of the "install browser extension" CTA on the search page
    - Total number of clicks on the "install browser extension" CTA on the search page

- CTA usage data
  - Browser extension
    - Number of users who viewed / clicked the "install browser extension" CTA on the file / search pages today
    - Number of views / clicks on the "install browser extension" CTA on the file / search pages today
  - IDE extension
    - Number of users who viewed / clicked the "install IDE extension" CTA on the file / search pages today
    - Number of views / clicks on the "install IDE extension" CTA on the file / search pages today

## CIDR Range for Sourcegraph

Sourcegraph currently uses Cloudflare to provide web application security. You should allow access to all [Cloudflare IP ranges](https://www.cloudflare.com/ips/)

## Connections to Sourcegraph.com

Sourcegraph only connects to Sourcegraph.com for two purposes:

1. The pings described above are sent, in order to:
   - Check for new product updates.
   - Send [anonymous, non-specific, aggregate metrics](#pings) back to Sourcegraph.com (see the full list above).
1. [Sourcegraph extensions](../extensions/index.md) are fetched from Sourcegraph.com`s extension registry (unless you are using a [private extension registry](extensions.md#publish-extensions-to-a-private-extension-registry)).

There are no other automatic external connections to Sourcegraph.com (or any other site on the internet).

