# Developing the code insights backend

- [MVP state of the backend](#mvp-state-of-the-backend)
- [Architecture](#architecture)
- [Life of an insight](#life-of-an-insight)
    - [(1) User defines insight in settings](#1-user-defines-insight-in-settings)
    - [(2) The _insight enqueuer_ detects the new insight](#2-the-insight-enqueuer-detects-the-new-insight)
    - [(3) The queryrunner worker gets work and runs the search query](#3-the-queryrunner-worker-gets-work-and-runs-the-search-query)
    - [(4) The historical data enqueuer gets to work](#4-the-historical-data-enqueuer-gets-to-work)
    - [(5) Query-time and rendering!](#5-query-time-and-rendering)
- [Debugging](#debugging)
    - [Accessing the TimescaleDB instance](#accessing-the-timescaledb-instance)
    - [Finding logs](#finding-logs)
    - [Inspecting the Timescale database](#inspecting-the-timescale-database)
        - [Querying data](#querying-data)
        - [Inserting data](#inserting-data)
- [Creating DB migrations](#creating-db-migrations)

## MVP state of the backend

The current code insights backend is an MVP version contributed by @slimsag - it:

* Supports search-based insights today, but was designed to support any type of insight in the future.
* Is backed by a [TImescaleDB](https://www.timescale.com) instance.
* Follows Sourcegraph architecture pretty well, is as clean as I could make it, has pretty good test coverage, and is generally an OK starting point.
* Definitely has some work to be done / is very much an MVP.
* There is a [project board](https://github.com/orgs/sourcegraph/projects/122) with detailed info on what work remains in my eyes.

Ultimately, if you're reading this you should view this code as a "take it or leave it" offering. My hope is you'll benefit from it (even if just from a "what problems will I face?" explanation) and can use it as a starting point - but you should do what makes you most comfortable and successful. If throwing it all out helps, do that! Happy to help in any way I can, message me @slimsag on GitHub / @stephen on Slack.

## Architecture

The following architecture diagram shows how the backend fits into the two Sourcegraph services "frontend" (the Sourcegraph monolithic service) and "repo-updater" (the Sourcegraph "background-worker" service), click to expand:

[![](diagrams/architecture.svg)](https://raw.githubusercontent.com/sourcegraph/sourcegraph/main/doc/dev/background-information/insights/diagrams/architecture.svg)

## Feature Flags
Code Insights is currently an experimental feature, and ships with an "escape hatch" feature flag that will completely disable the dependency on TimescaleDB (named `codeinsights-db`). This feature flag is implemented as an environment variable that if set true `DISABLE_CODE_INSIGHTS=true` will disable the dependency and will not start the Code Insights background workers or GraphQL resolvers. This variable must be set on both the `repo-updater` and `frontend` services to remove the dependency. If the flag is not set on both services, the `codeinsights-db` dependency will be required.

Implementation of this environment variable can be found in the [`frontend`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/internal/insights/insights.go#L43) and [`repo-updater`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/internal/insights/background/background.go#L30) services.

This flag should be used judiciously and should generally be considered a last resort for Sourcegraph installations that need to disable Code Insights or remove the database dependency.

## Life of an insight

### (1) User defines insight in settings

A user defines an insight in their user settings (today, only global user settings are supported [#18397](https://github.com/sourcegraph/sourcegraph/issues/18397)), e.g. at https://sourcegraph.com/site-admin/global-settings by adding:

```jsonb
    "insights": [
      {
        "title": "fmt usage",
        "description": "fmt.Errorf/fmt.Printf usage",
        "series": [
          {
            "label": "fmt.Errorf",
            "search": "errorf",
          },
          {
            "label": "printf",
            "search": "fmt.Printf",
          }
        ]
      }
    ],
```

This defines a single insight called `fmt usage`, with two series of data (the number of search results the queries `errorf` and `fmt.Printf` return, respectively.)

Once defined in user settings, the insight will immediately show up for users at e.g. https://sourcegraph.com/insights - as long as they have the feature flag turned on in their user/org/global settings:

```jsonb
  "experimentalFeatures": {
    "codeInsights": true
  },
```

### (2) The _insight enqueuer_ detects the new insight

The _insight enqueuer_ ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+newInsightEnqueuer&patternType=literal)) is a background goroutine running in the `repo-updater` service of Sourcegraph ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+StartBackgroundJobs&patternType=literal)), which runs all background goroutines for Sourcegraph - so long as `DISABLE_CODE_INSIGHTS=true` is not set on the repo-updater container/process.

Every 12 hours on and after process startup ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+file:insight_enqueuer.go+NewPeriodic&patternType=literal)) it does the following:

1. Discovers insights defined in global/org/user settings ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+file:insight_enqueuer.go+discovery.Discover&patternType=literal)) by enumerating all settings on the instance and looking for the `insights` key, compiling a list of them (today, just global settings [#18397](https://github.com/sourcegraph/sourcegraph/issues/18397)).
2. Determines which _series_ are unique. For example, if Jane defines a search insight with `"search": "fmt.Printf"` and Bob does too, there is no reason for us to collect data on those separately since they represent the same exact series of data. Thus, we hash the insight definition ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+file:insight_enqueuer.go+EncodeSeriesID&patternType=literal)) in order to deduplicate them and produce a _series ID_ string that will uniquely identify that series of data. We also use this ID to identify the series of data in the `series_points` TimescaleDB database table later.
3. For every unique series, enqueues a job for the _queryrunner_ worker to later run the search query and collect information on it (like the # of search results.) ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+file:insight_enqueuer.go+enqueueQueryRunnerJob&patternType=literal))

### (3) The queryrunner worker gets work and runs the search query

The queryrunner ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+file:queryrunner&patternType=literal)) is a background goroutine running in the `repo-updater` service of Sourcegraph ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+StartBackgroundJobs&patternType=literal)), it is responsible for:

1. Taking work that has been enqueued by the _insight enqueuer_ (specifically, just search-based insights) and dequeueing it. ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+file:queryrunner+dequeueJob&patternType=literal))
2. Handling each job ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+file:queryrunner+content:%22%29+Handle%28%22&patternType=literal)) by running a search query using Sourcegraph's internal/unauthenticated GraphQL API ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+file:queryrunner+content:%22search%28%22&patternType=literal)) (i.e. getting all results, even if the user doesn't have access to some repos)
3. Actually recording the number of results and other information we care about into the _insights store_ (i.e. into the `series_points` TimescaleDB table) ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+file:queryrunner+RecordSeriesPoint&patternType=literal)).

### (4) The historical data enqueuer gets to work

If we record one data point every 12h above, it would take months or longer for users to get any value out of backend insights. This introduces the need for us to backfill data by running search queries that answer "how many results existed in the past?" so we can populate historical data.

Similar to the _insight enqueuer_, the _historical insight enqueuer_ is a background goroutine ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:insights+lang:go+newInsightHistoricalEnqueuer&patternType=literal)) which locates and enqueues work to populate historical data points.

It is a moderately complex algorithm, to get an understanding for how it operates see these two explanations:

1. [what it does and general implementation thoughts](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@049b99763b10ddc7ef6f72c22b18f1fa5f4f7259/-/blob/enterprise/internal/insights/background/historical_enqueuer.go#L31-53)
2. [overview of the algorithm](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/internal/insights/background/historical_enqueuer.go?L185#L116-147)

Naively implemented, the historical backfiller would take a long time on any reasonably sized Sourcegraph installation. As an optimization,
the backfiller will only query for data frames that have recorded changes in each repository. This is accomplished by looking
at an index of commits and determining if that frame is eligible for removal. [code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/internal/insights/compression/compression.go?L46:1)

There is a rate limit associated with analyzing historical data frames. This limit can be configured using the site setting
`insights.historical.worker.rateLimit`. As a rule of thumb, this limit should be set as high as possible without performance
impact to `gitserver`. A likely safe starting point on most Sourcegraph installations is `insights.historical.worker.rateLimit=20`.

### (5) Query-time and rendering!

The webapp frontend invokes a GraphQL API which is served by the Sourcegraph `frontend` monolith backend service in order to query information about backend insights. ([cpde](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:enterprise/+lang:go+InsightConnectionResolver&patternType=literal))

* A GraphQL _series_ resolver returns all of the distinct data series in a single insight (UI panel) ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:enterprise/+file:resolver+lang:go+Series%28&patternType=literal))
* A GraphQL resolver ultimately provides data points for a single series of data ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:enterprise/+file:resolver+lang:go+Points%28&patternType=literal))
* The _series points resolver_ merely queries the _insights store_ for the data points it needs, and the store itself merely runs SQL queries against the TimescaleDB database to get the datapoints ([code](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:enterprise/+file:store+lang:go+SeriesPoints%28&patternType=literal))

Note: There are other better developer docs which explain the general reasoning for why we have a "store" abstraction. Insights usage of it is pretty minimal, we mostly follow it to separate SQL operations from GraphQL resolver code and to remain consistent with the rest of Sourcegraph's architecture.

Once the web client gets data points back, it renders them! Contact @felixfbecker for details on where/how that happens.

These queries can be executed concurrently by using the site setting `insights.query.worker.concurrency` and providing
the desired concurrency factor. With `insights.query.worker.concurrency=1` queries will be executed in serial.

There is a rate limit associated with the query worker. This limit is shared across all concurrent handlers and can be configured
using the site setting `insights.query.worker.rateLimit`. This value to set will depend on the size and scale of the Sourcegraph
installations `Searcher` service.

## Debugging

This being a pretty complex and slow-moving system, debugging can be tricky. This is definitely one area we need to improve especially from a user experience point of view ([#18964](https://github.com/sourcegraph/sourcegraph/issues/18964)) and general customer debugging point of view ([#18399](https://github.com/sourcegraph/sourcegraph/issues/18399)).

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

Since insights runs inside of the `frontend` and `repo-updater` containers/processes, it can be difficult to locate the relevant logs. Best way to do it is to grep for `insights`.

The `frontend` will contain logs about e.g. the GraphQL resolvers and TimescaleDB migrations being ran, while `repo-updater` will have the vast majority of logs coming from the insights background workers.

#### Docker compose deployments

```
docker logs sourcegraph-frontend-0 | grep insights
```

and

```
docker logs repo-updater | grep insights
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

##### Query data the way we do for the frontend, but for every series in the last 6mo

```sql
SELECT sub.series_id, sub.interval_time, SUM(value) AS value, NULL AS metadata
FROM (WITH target_times AS (SELECT *
                            FROM GENERATE_SERIES(CURRENT_TIMESTAMP::DATE - INTERVAL '6 months', CURRENT_TIMESTAMP::DATE,
                                                 '2 weeks') AS interval_time)
      SELECT sub.series_id, sub.repo_id, sub.value, interval_time
      FROM (SELECT DISTINCT repo_id, series_id FROM series_points) AS r
             CROSS JOIN target_times tt
             JOIN LATERAL (
        SELECT sp.*
        FROM series_points AS sp
        WHERE sp.repo_id = r.repo_id
          AND sp.time <= tt.interval_time
          AND sp.series_id = r.series_id
        ORDER BY time DESC
        LIMIT 1
        ) sub ON sub.repo_id = r.repo_id AND r.series_id = sub.series_id
      ORDER BY interval_time, repo_id) AS sub
GROUP BY sub.series_id, sub.interval_time
ORDER BY interval_time DESC
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

Currently, the migration process blocks `frontend` and `repo-updater` startup - which is one issue [we will need to solve](https://github.com/sourcegraph/sourcegraph/issues/18388).
