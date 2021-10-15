# Developing the code insights backend

- [Beta state of the backend](#beta-state-of-the-backend)
- [Architecture](#architecture)
- [Life of an insight](#life-of-an-insight)
    - [(1) User defines insight in settings](#1-user-defines-insight-in-settings)
    - [(2) The _insight enqueuer_ detects the new insight](#2-the-insight-enqueuer-indexed-recorder-detects-the-new-insight)
    - [(3) The queryrunner worker gets work and runs the search query](#3-the-historical-data-enqueuer-historical-recorder-gets-to-work)
    - [(4) The historical data enqueuer gets to work](#4-the-queryrunner-worker-gets-work-and-runs-the-search-query)
    - [(5) Query-time and rendering!](#5-query-time-and-rendering)
- [Debugging](#debugging)
    - [Accessing the TimescaleDB instance](#accessing-the-timescaledb-instance)
    - [Finding logs](#finding-logs)
    - [Inspecting the Timescale database](#inspecting-the-timescale-database)
        - [Querying data](#querying-data)
        - [Inserting data](#inserting-data)
- [Creating DB migrations](#creating-db-migrations)

## Beta state of the backend

The current code insights backend is a beta version contributed by [Coury](https://about.sourcegraph.com/handbook/company/team#coury-clark-he-him) (based on previous work by [Stephen](https://about.sourcegraph.com/handbook/company/team#stephen-gutekanst)) - it:

* Supports running search-based insights over all indexable repositories on the Sourcegraph installation.
* Is backed by a [TimescaleDB](https://www.timescale.com) instance. See the [database section](#database) below for more information.
* Optimizes unnecessary search queries by using an index of commits to query only for time periods that have had at least one commit.
* Supports regexp based drilldown on repository name.
* Provides permissions restrictions by filtering of repositories that are not visible to the user at query time.
* Does not yet support synchronous insight creation through an API. Read more below in the [Insight Metadata section](#insight-metadata-section).

The current version of the backend is an MVP to achieve beta status to unblock the feature request of "running an insight over all my repos".

## Architecture

The following architecture diagram shows how the backend fits into the two Sourcegraph services "frontend" (the Sourcegraph monolithic service) and "worker" (the Sourcegraph "background-worker" service), click to expand:

[![Architecture diagram](diagrams/architecture.svg)](https://raw.githubusercontent.com/sourcegraph/sourcegraph/main/doc/dev/background-information/insights/diagrams/architecture.svg)

## Deployment Status
Code Insights backend is currently disabled on `sourcegraph.com` until solutions can be built to address the large indexed repo count.

## Feature Flags
Code Insights is currently an experimental feature, and ships with an "escape hatch" feature flag that will completely disable the dependency on TimescaleDB (named `codeinsights-db`). This feature flag is implemented as an environment variable that if set true `DISABLE_CODE_INSIGHTS=true` will disable the dependency and will not start the Code Insights background workers or GraphQL resolvers. This variable must be set on both the `worker` and `frontend` services to remove the dependency. If the flag is not set on both services, the `codeinsights-db` dependency will be required.

Implementation of this environment variable can be found in the [`frontend`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/internal/insights/insights.go#L43) and [`worker`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/internal/insights/background/background.go#L30) services.

This flag should be used judiciously and should generally be considered a last resort for Sourcegraph installations that need to disable Code Insights or remove the database dependency.

With version 3.31 this flag has moved from the `repo-updater` service to the `worker` service.

### Sourcegraph Setting
Code Insights is currently behind an experimental feature on Sourcegraph. You can enable it in settings.

```json
  "experimentalFeatures": {
    "codeInsights": true
  },
```

## Database
Currently, Code Insights uses a [TimescaleDB](https://www.timescale.com) database running on the OSS license. The original intention was to use
some of the timeseries query features, as well as the hypertable. Many of these are behind a proprietary license that would require non-trivial
work to bundle with Sourcegraph.

Additionally, we have many customers running on managed databases for Postgres (RDS, Cloud SQL, etc) that do not support the TimescaleDB plugin.
Recently our distribution team has started to encourage customers to use managed DB solutions as the product grows. Given entire categories of customers
would be excluded from using Code Insights, we have decided we must move away from TimescaleDB.

A final decision has not yet been made, but a very likely candidate is falling back to vanilla Postgres. This will simplify our operations, support, and likely
will not present a performance problem given the primary constraint on Code Insights is search throughput.

It is reasonable to expect this migration to occur some time during the beta period for Code Insights.

## Insight Metadata
Historically, insights ran entirely within the Sourcegraph extensions API on the browser. These insights are limited to small sets of manually defined repositories
since they execute in real time on page load with no persistence of the timeseries data. Sourcegraph extensions have access to settings (user / org / global) ,
so the original storage location for extension based insight metadata (query string, labels, title, etc) was settings.

This storage location persisted for the backend MVP, but is in the process of being deprecated by moving the metadata to the database. Given roadmap constraints
an API does not currently exist to synchronously interact with the database for metadata. A background process attempts to sync insight metadata that is flagged as
"backend compatible" on a regular interval.

As expected, this async process causes many strange UX / UI bugs that are difficult or impossible to solve. An API to fully deprecate the settings storage is a priority
for Q3.

As an additional note, [extension based insights are read from settings](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@43062781be6c648f40b5baec32ce8241c03cbd18/-/blob/internal/usagestats/code_insights.go?L179) for the purposes of sending aggregated pings.

## Life of an insight

### (1) User defines insight in settings

A user creates a code insight using the creation UI, and selects the option to run the insight over all repositories. The Code Insights will create a JSON object
in the appropriate settings (user / org) and place it in the `insights.allrepos` dictionary. Note: only insights placed in the `insights.allrepos` dictionary are considered eligible for 
sync to prevent conflicts with extensions insights.

An example backend-compatible insight definition in settings:
```json
"insights.allrepos": {
    "searchInsights.insight.soManyInsights": {
      "title": "So many insights",
      "series": [
        {
          "name": "\"insights\" insights",
          "stroke": "var(--oc-blue-7)",
          "query": "insights"
        }
      ]
    },
}
```

### Unique ID
An Insight View is defined to have a globally unique referencable ID. For the time being to match feature parity with extensions insights the ID is generated as the
chart title prefixed with `searchInsights.insight.`.

In the above example, the ID is `searchInsights.insight.soManyInsights`.

[Read more about Insight Views](./insight_view.md)

### Sync to the database

The settings sync [job](https://sourcegraph.com/github.coem/sourcegraph/sourcegraph@4306278/-/blob/enterprise/internal/insights/discovery/discovery.go?L138:6)
will execute and attempt to migrate the defined insight. Currently, the sync job does not handle updates and will only sync if the insight view unique ID is not found.

Until the insight metadata is synced, the GraphQL response will not return any information if given the unique ID. Temporarily, the frontend treats all `404` errors
as a transient "Insight is processing" error to solve for this weird UX.

Once the sync job is complete, the following database rows will have been created:
1. An Insight View (`insight_view`) with UniqueID `searchInsights.insight.soManyInsights`
2. An Insight Series (`insight_series`) with metadata required to generate the data series
3. A link from the view to the data series (`insight_view_series`)

#### A note about data series
Currently, data series are defined without scope for specific repositories or any other subset of repositories (all data series iterate over all repos). Data series are uniquely identified
by [hashing the query string](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@4306278/-/blob/enterprise/internal/insights/discovery/series_id.go?L32:6), with the `s:` prefix.
This field is known as the `series_id`. It must be globally unique, and any collisions will be assumed to be the same exact data series.

In the medium term this semantic will change to include repository scopes (assigning specific repos to a datseries), and may possibly change entirely. This is one important area
of design and work for Q3.

The `series_id` for the example Insight data above series would be `s:7F1FE30EF252BF75FAB0C9680C7BCFFF648154165AFE718155091051255A0A99`

The `series_id` is how the underlying data series is referenced throughout the system; however, it is not currently exposed in the GraphQL API. The current model
prefers to obfuscate the underlying data series behind an [Insight View](./insight_view.md). This model is not highly validated, and may need to change in the future
to expose more direct functionality around data series.

### (2) The _insight enqueuer_ (indexed recorder) detects the new insight

The _insight enqueuer_ ([code](https://sourcegraph.com/search?qe=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+newInsightEnqueuer&patternType=literal)) is a background goroutine running in the `worker` service of Sourcegraph ([code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@55be9054a2609e06a1d916cc2f782827421dd2a3/-/blob/enterprise/internal/insights/background/insight_enqueuer.go?L27:6)), which runs all background goroutines for Sourcegraph - so long as `DISABLE_CODE_INSIGHTS=true` is not set on the `worker` container/process.
Its job is to periodically schedule a recording of 'current' values for Insights by queuing a snapshot recording using an indexed query. This only requires a single query per insight regardless of the number of repositories,
and will return results for all the matched repositories. Each repository will still be recorded individually. These queries are placed on the same queue as historical queries (`insights_query_runner_jobs`) and can
be identified by the lack of a revision and repo filter on the query string. 
For example, `insights` might be an indexed recording, where `insights repo:^codehost\.com/myorg/somerepo@ref$` would be a historical recording for a specific repo / revision.
You can find these search queries for queued jobs on the (primary postgres) table `insights_query_runner_jobs.search_query`

Insight recordings are scheduled using the database field (codeinsights-db) `insight_series.next_recording_after`, and will only be taken if the field time is less than the execution time of the job.
Recordings are currently always scheduled to occur on the first day of the following month, after `00:00:00`. For example, if a recording was taken at `2021-08-27T15:29:00.000Z` the next
recording will be scheduled for `2021-09-01T00:00:00.000Z`. The first indexed recording after insight creation will occur on the same interval.

Note: There is a field (codeinsights-db) `insight_series.recording_interval_days` that was intended to provide some configurable value to this recording interval. We have limited
product validation with respect to time intervals and the granularity of recordings, so beta has launched with fixed `first-of-month` scheduling. 
This will be an area of development throughout Q3 and into Q4.

### (3) The historical data enqueuer (historical recorder) gets to work

If we only record one data point per repo every month, it would take months or longer for users to get any value out of backend insights. This introduces the need for us to backfill data by running search queries that answer "how many results existed in the past?" so we can populate historical data.

Similar to the _insight enqueuer_, the _historical insight enqueuer_ is a background goroutine ([code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@55be9054a2609e06a1d916cc2f782827421dd2a3/-/blob/enterprise/internal/insights/background/historical_enqueuer.go?L75:6)) which locates and enqueues work to populate historical data points.

The most naive implementation of backfilling is as follows:
``` 
For each relevant repository:
  For each relevant time point:
    Execute a search at the most recent revision
```

Naively implemented, the historical backfiller would take a long time on any reasonably sized Sourcegraph installation. As an optimization,
the backfiller will only query for data frames that have recorded changes in each repository. This is accomplished by looking
at an index of commits and determining if that frame is eligible for removal. 
Read more [below](#Backfill-compression)

There is a rate limit associated with analyzing historical data frames. This limit can be configured using the site setting
`insights.historical.worker.rateLimit`. As a rule of thumb, this limit should be set as high as possible without performance
impact to `gitserver`. A likely safe starting point on most Sourcegraph installations is `insights.historical.worker.rateLimit=20`.

#### Backfill compression
Read more about the backfilling compression in the proposal [RFC 392](https://docs.google.com/document/d/1VDk5Buks48THxKPwB-b7F42q3tlKuJkmUmaCxv2oEzI/edit#heading=h.3babtpth82k2)

We maintain an index of commits (table `commit_index` in codeinsights-db) that are used to filter out repositories that do not need a search query. This index is 
periodically [refreshed](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@55be9054a2609e06a1d916cc2f782827421dd2a3/-/blob/enterprise/internal/insights/compression/worker.go?L41) 
with changes since its previous refresh. Metadata for each repositories refresh is tracked in a table `commit_index_metadata`. 

To avoid race conditions with the index, data frames are only [filtered](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@55be9054a2609e06a1d916cc2f782827421dd2a3/-/blob/enterprise/internal/insights/compression/compression.go?L84)
out if the `commit_index_metadata.last_updated_at` is greater than the data point we are attempting to compress.

Currently, we only generate 12 months of history for this commit index to keep it reasonably sized. We do not currently do any pruning, but that is likely an area
we will need to expand in Q3 - Q4.

#### Limiting to a scope of repositories
Naturally, some insights will not need or want to execute over all repositories and would prefer to execute over a subset to generate faster. As a trade off to reach beta
we made the decision that all insights will execute over all repositories. The primary justification was that the most significant blocker for beta was the ability to run
over all insights, and therefore unlocking that capability also unlocks the capability for users that want to run over a subset, they will just need to wait longer.

This is non-trivial problem to solve, and raises many questions:
1. How do we represent these sets? Do we list each repository out for each insight? This could result in a very large cardinality and grow the database substantially.
2. What happens if users change the set of repositories after we have already backfilled?
3. What does the architecture of this look like internally? How do we balance the priority of backfilling other larger insights with much smaller ones?

This is also a blocker to migrate all functionality away from extensions and to the backend, because the extesions *do* support small numbers of repositories at this time.

This will be an area of work for Q3 - Q4.

#### Detecting if an insight is _complete_
Given the large possible cardinality of required queries to backfill an insight, it is clear this process can take some time. Through dogfooding we have found
on a Sourcegraph installation with ~36,000 repositories, we can expect to backfill an average insight in 20-30 minutes. The actual benchmarks of how long 
this will take vary greatly depending on the commit patterns and size of the Installation.

One important piece of information that needs to be surfaced to users is the answer to the question `is my insight still processing?`
This is a non-trivial question to answer:
1. Work is processed asynchronously, so querying the state of the queue is necessary
2. Iterating many thousands of repositories can result in some transient errors causing individual repositories to fail, and ultimately not be included in the queue [issue](https://github.com/sourcegraph/sourcegraph/issues/23844)
3. The current shared state between settings and the database leaves a lot of intermediate undefined states, such as prior to the sync

As a temporary measure to try and answer this question with some degree of accuracy, the historical backfiller applies the following semantic:
Flag an insight as `completed backfill` if the insight was able to complete one full iteration of all repositories without any `hard` errors (such as low level DB errors, etc).
This flag is represented as the database field `insight_series.backfill_queued_at` and is [set](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24%4055be905+StampBackfill&patternType=literal) at the end of the complete repository iteration.

This semantic does not fully capture all possible states. For example, if a repository encounters a `soft` error (unable to fetch git metadata, for example)
it will be skipped and ultimately not populate in the data series. Improving this is an area of design and work for Q3 - Q4.

### (4) The queryrunner worker gets work and runs the search query

The queryrunner ([code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@55be9054a2609e06a1d916cc2f782827421dd2a3/-/blob/enterprise/internal/insights/background/queryrunner/worker.go)) is a background goroutine running in the `worker` service of Sourcegraph ([code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@55be9054a2609e06a1d916cc2f782827421dd2a3/-/blob/enterprise/internal/insights/background/queryrunner/worker.go?L42:6)), it is responsible for:

1. Dequeueing search queries that have been queued by the either the indexed or historical recorder. Queries are stored with a `priority` field that 
   [dequeues](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@55be905/-/blob/enterprise/internal/insights/background/queryrunner/worker.go?L134) queries in ascending priority order (0 is higher priority than 100).
2. Executing a search against Sourcegraph with the provided query. These queries are executed against the `internal` GraphQL endpoint, meaning they are *unauthorized* and can see all results. This allows us to build global results and filter based on user permissions at query time.
3. Flagging any error states (such as limitHit, meaning there was some reason the search did not return all possible results) as a `dirty query`.
   These queries are stored in a table `insight_dirty_queries` that allow us to surface some information to the end user about the data series.
   Not all error states are currently collected here, and this will be an area of work for Q3.
4. Aggregating the search results, per repository (and in the near-future, per unique match to support capture groups) and storing them in the `series_points` table.

The queue is managed by a common executor called `Worker` (note: the naming collision with the `worker` service is confusing, but they are not the same).
[Read more about `Worker` and how it works in this search notebook](https://sourcegraph.com/search/notebook#md:%23%23%20Background%20Workers%0AA%20quick%20introduction%20to%20the%20background%20processing%20system%20in%20the%20Sourcegraph%20codebase.,md:%23%23%23%20Summary%0ASourcegraph%20uses%20a%20persistent%20queueing%20mechanism%20for%20long%20running%20background%20tasks%20called%20%60Worker%60.,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20dbworker.NewWorker,md:These%20tasks%20are%20stored%20in%20a%20table%20in%20the%20Postgres%20database%20where%20a%20single%20row%20represents%20a%20single%20invocation%20of%20a%20%60Handler%60.%20Each%20%60Worker%60%20uses%20a%20unique%20table.%20A%20background%20process%20will%20periodically%20%60dequeue%60%20records%20from%20the%20associated%20queue%20table%20and%20pass%20them%20to%20the%20provided%20%60Handler%60%20callback.,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20file%3Aworkerutil%20type%20Handler%20interface,md:See%20implementations%20of%20the%20%60Handler%60%20throughout%20the%20codebase,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20_%20workerutil.Handler,md:The%20%60Worker%60%20can%20be%20configured%20with%20options%20such%20as%20query%20interval%2C%20heartbeat%20interval%2C%20name%2C%20and%20more.,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20workerutil.WorkerOptions,md:You%20can%20create%20a%20%60Resetter%60%20to%20periodically%20reset%20any%20records%20that%20might%20have%20stalled.%20This%20is%20useful%20to%20make%20sure%20records%20process%20at%20least%20once%20without%20concern%20for%20transient%20errors%20%28such%20as%20pods%20terminating%2C%20etc%28,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20dbworker.NewResetter,md:If%20you%20want%20to%20add%20a%20new%20persistent%20queue%2C%20you%20will%20need%20to%20create%20a%20table%20that%20has%20all%20of%20the%20default%20queue%20columns%2C%20and%20any%20additional%20columns%20you%20want.,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20file%3Amigration%20create%20table%20.*_jobs%20patterntype%3Aregexp%20,md:You%20can%20interact%20with%20the%20queue%20table%20through%20a%20special%20%60Store%60.%20You%20can%20initialize%20the%20%60Store%60%20to%20automatically%20capture%20and%20report%20metrics.%20The%20metrics%20will%20have%20a%20prefix%20%60workerutil_dbworker_store%60.,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20dbworkerstore.NewWithMetrics,md:%60Worker%60%20%60Handler%60%20can%20be%20configured%20to%20emit%20metrics.%20Note%3A%20the%20provided%20name%20must%20have%20the%20%60_processor%60%20suffix%20to%20use%20a%20generated%20dashboard.,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20workerutil.NewMetrics,md:%60Resetter%60%20can%20be%20configured%20to%20emit%20metrics.,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20dbworker.NewMetrics,md:Dashboards%20can%20be%20generated%20for%20%60Worker%60%20%60Handler%60%20operations.,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20WorkerutilGroupOptions,md:Note%3A%20%60Handler%60%20metrics%20must%20be%20emitted%20with%20a%20postfix%20%60_processor%60%20for%20these%20dashbaords,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20_processor,md:Dashboards%20can%20be%20generated%20for%20%60Resetter%60%20operations,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20ResetterGroupOptions,md:Dashboards%20can%20be%20generated%20for%20the%20underlying%20%60Store%60.%20Note%3A%20the%20metrics%20are%20emitted%20with%20a%20prefix%20%60workerutil_dbworker_store%60,query:repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%20workerutil_dbworker_store_).

These queries can be executed concurrently by using the site setting `insights.query.worker.concurrency` and providing
the desired concurrency factor. With `insights.query.worker.concurrency=1` queries will be executed in serial.

There is a rate limit associated with the query worker. This limit is shared across all concurrent handlers and can be configured
using the site setting `insights.query.worker.rateLimit`. This value to set will depend on the size and scale of the Sourcegraph
installations `Searcher` service.

### (5) Query-time and rendering!

The webapp frontend invokes a GraphQL API which is served by the Sourcegraph `frontend` monolith backend service in order to query information about backend insights. ([cpde](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:enterprise/+lang:go+InsightConnectionResolver&patternType=literal))

1. A GraphQL _series_ resolver returns all of the distinct data series in a single insight (UI panel) ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:enterprise/+file:resolver+lang:go+Series%28&patternType=literal))
2. A GraphQL resolver ultimately provides data points for a single series of data ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:enterprise/+file:resolver+lang:go+Points%28&patternType=literal))
3. The _series points resolver_ merely queries the _insights store_ for the data points it needs, and the store itself merely runs SQL queries against the TimescaleDB database to get the datapoints ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:enterprise/+file:store+lang:go+SeriesPoints%28&patternType=literal))

Note: There are other better developer docs which explain the general reasoning for why we have a "store" abstraction. Insights usage of it is pretty minimal, we mostly follow it to separate SQL operations from GraphQL resolver code and to remain consistent with the rest of Sourcegraph's architecture.

Once the web client gets data points back, it renders them! For more information, please contact an @codeinsights frontend engineer.

#### User Permissions
We made the decision to generate data series for all repositories and restrict the information returned to the user at query time. There were a few driving factors
behind this decision:
1. We have split feedback between customers that want to share insights globally without regard for permissions, and other customers that want strict permissions mapped to repository visibility.
  In order to possibly support both (or either), we gain the most flexibility by performing query time limitations.
2. We can reuse pre-calculated data series across multiple users if they provide the same query to generate an insight. This not only reduces the storage overhead, but makes
  the user experience substantially better if the data series is already calculated.
   
Given the large possible cardinality of the visible repository set, it is not practical to select all repos a user has access to at query time. Additionally, this data does not live
in the same database as the timeseries data, requiring some network traversal.

User permissions are currently implemented by negating the set of repos a user does *not* have access to. This is based on the assumption that most users
of Sourcegraph have access to most repositories. This is a fairly highly validated assumption, and matches the premise of Sourcegraph to begin with (that you can search across all repos).
This may not be suitable for Sourcegraph installations with highly controlled repository permissions, and may need revisiting.

### Storage Format
The code insights time series are currently stored entirely within Postgres. 

As a design, insight data is stored as a full vector of match results per unique time point. This means that for some time `T`, all of the unique timeseries that fall under
one insight series can be aggregated to form the total result. Given that the processing system will execute every query at-least once, the possiblity of duplicates
exist within a unique timeseries. A simple deduplication is performed at query time.

Read more about the [history](https://github.com/sourcegraph/sourcegraph/issues/23690) of this format.

## Debugging

This being a pretty complex, high cardinality, and slow-moving system - debugging can be tricky.

In this section, I'll cover useful tips I have for debugging the system when developing it or otherwise using it.

### Accessing the TimescaleDB instance

#### Dev and docker compose deployments

```
docker exec -it codeinsights-db psql -U postgres
```

#### Kubernetes deployments

```
kubectl exec -it deployment/codeinsights-db -- psql -U postgres
```

* If trying to access Sourcegraph.com's DB: `kubectl -n prod exec -it deployment/codeinsights-db -- psql -U postgres`
* If trying to access k8s.sgdev.org's DB: `kubectl -n dogfood-k8s exec -it deployment/codeinsights-db -- psql -U postgres`

### Finding logs

Since insights runs inside of the `frontend` and `worker` containers/processes, it can be difficult to locate the relevant logs. Best way to do it is to grep for `insights`.

The `frontend` will contain logs about e.g. the GraphQL resolvers and TimescaleDB migrations being ran, while `worker` will have the vast majority of logs coming from the insights background workers.

#### Docker compose deployments

```
docker logs sourcegraph-frontend-0 | grep insights
```

and

```
docker logs worker | grep insights
```

### Inspecting the Timescale database

Read the [initial schema migration](https://github.com/sourcegraph/sourcegraph/blob/main/migrations/codeinsights/1000000001_initial_schema.up.sql) which contains all of the tables we create in TimescaleDB and describes them in detail. This will explain the general layout of the database schema, etc.

The most important table in TimescaleDB is `series_points`, that's where the actual data is stored. It's a [hypertable](https://docs.timescale.com/latest/using-timescaledb/hypertables).

#### Querying data

```sql
SELECT * FROM series_points ORDER BY time DESC LIMIT 100;
```

##### Query data, filtering by repo and returning metadata

```sql
SELECT *
FROM series_points
JOIN metadata ON metadata.id = metadata_id
WHERE repo_name_id IN (
    SELECT id FROM repo_names WHERE name ~ '.*-renamed'
)
ORDER BY time
DESC LIMIT 100;
```

(note: we don't actually use metadata currently, so it's always empty.)

###### Query data, filter by metadata containing `{"hello": "world"}`

```sql
SELECT *
FROM series_points
JOIN metadata ON metadata.id = metadata_id
WHERE metadata @> '{"hello": "world"}'
ORDER BY time
DESC LIMIT 100;
```

(note: we don't actually use metadata currently, so it's always empty.)

###### Query data, filter by metadata containing Go languages

```sql
SELECT *
FROM series_points
JOIN metadata ON metadata.id = metadata_id
WHERE metadata @> '{"languages": ["Go"]}'
ORDER BY time
DESC LIMIT 100;
```

(note: we don't actually use metadata currently, so it's always empty. The above gives you some ideas for how we intended to use it.)

See https://www.postgresql.org/docs/9.6/functions-json.html for more metadata `jsonb` operator possibilities. Only `?`, `?&`, `?|`, and `@>` operators are indexed (gin index)

##### Query data the way we do for the frontend, but for every series

```sql
SELECT sub.series_id, sub.interval_time, SUM(sub.value) AS value, sub.metadata
FROM (
       SELECT sp.repo_name_id, sp.series_id, sp.time AS interval_time, MAX(value) AS value, NULL AS metadata
       FROM series_points sp
              JOIN repo_names rn ON sp.repo_name_id = rn.id
       GROUP BY sp.series_id, interval_time, sp.repo_name_id
       ORDER BY sp.series_id, interval_time, sp.repo_name_id DESC
     ) sub
GROUP BY sub.series_id, sub.interval_time, sub.metadata
ORDER BY sub.series_id, sub.interval_time DESC
```

#### Inserting data

##### Upserting repository names

The `repo_names` table contains a mapping of repository names to small numeric identifiers. You can upsert one into the database using e.g.:

```sql
WITH e AS(
    INSERT INTO repo_names(name)
    VALUES ('github.com/gorilla/mux-original')
    ON CONFLICT DO NOTHING
    RETURNING id
)
SELECT * FROM e
UNION
    SELECT id FROM repo_names WHERE name='github.com/gorilla/mux-original';
```

##### Upserting event metadata

Similar to `repo_names`, there is a separate `metadata` table which stores unique metadata jsonb payloads and maps them to small numeric identifiers. You can upsert metadata using e.g.:

```sql
WITH e AS(
    INSERT INTO metadata(metadata)
    VALUES ('{"hello": "world", "languages": ["Go", "Python", "Java"]}')
    ON CONFLICT DO NOTHING
    RETURNING id
)
SELECT * FROM e
UNION
    SELECT id FROM metadata WHERE metadata='{"hello": "world", "languages": ["Go", "Python", "Java"]}';
```

##### Inserting a data point

You can insert a data point using e.g.:

```sql
INSERT INTO series_points(
    series_id,
    time,
    value,
    metadata_id,
    repo_id,
    repo_name_id,
    original_repo_name_id
) VALUES(
    "my unique test series ID",
    now(),
    0.5,
    (SELECT id FROM metadata WHERE metadata = '{"hello": "world", "languages": ["Go", "Python", "Java"]}'),
    2,
    (SELECT id FROM repo_names WHERE name = 'github.com/gorilla/mux-renamed'),
    (SELECT id FROM repo_names WHERE name = 'github.com/gorilla/mux-original')
);
```

You can omit all of the `*repo*` fields (nullable) if you want to store a data point describing a global (associated with no repository) series of data.

##### Inserting fake generated data points

TimescaleDB has a `generate_series` function you can use like this to insert one data point every 15 days for the last year:

```
INSERT INTO series_points(
    series_id,
    time,
    value,
    metadata_id,
    repo_id,
    repo_name_id,
    original_repo_name_id)
SELECT time,
    "my unique test series ID",
    random()*80 - 40,
    (SELECT id FROM metadata WHERE metadata = '{"hello": "world", "languages": ["Go", "Python", "Java"]}'),
    2,
    (SELECT id FROM repo_names WHERE name = 'github.com/gorilla/mux-renamed'),
    (SELECT id FROM repo_names WHERE name = 'github.com/gorilla/mux-original')
    FROM generate_series(TIMESTAMP '2020-01-01 00:00:00', TIMESTAMP '2021-01-01 00:00:00', INTERVAL '15 day') AS time;
```

## Creating DB migrations

Since TimescaleDB is just Postgres (with an extension), we use the same SQL migration framework we use for our other Postgres databases. `migrations/codeinsights` in the root of this repository contains the migrations for the Code Insights Timescale database, they are executed when the frontend starts up (as is the same with e.g. codeintel DB migrations.)

Currently, the migration process blocks `frontend` and `worker` startup - which is one issue [we will need to solve](https://github.com/sourcegraph/sourcegraph/issues/18388).
