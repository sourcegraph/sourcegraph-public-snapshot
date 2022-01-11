# Managing Code Insights with the API

Learn how to manage [Code Insights](../../code_insights/index.md) on private Sourcegraph instances with the API. If you haven't used the API before, learn more about [the GraphQL API and how to use it](index.md).

## Prerequisites

* TODO-JOEL: what do we want to put here? Maybe a release number? A note about the environment flag that disables settings? That they have insights turned on to begin with?

### General

This API is considered experimental and there may be significant changes in the future. TODO-JOEL: is this the right thing to say here? We've talked about the API being experimental, and it seems worth calling it out in some way.

### Permissions and visibility overview

TODO-JOEL: Thinking about permissions and what might be useful to document here.. insights themselves don't really have permissions. By default the user who created the insight will have permission, but there's no way to share that with other users without first creating a dashboard. Would it be useful to call out how to create a dashboard and add/remove insights from it, just for sharing purposes? Or is this not a compelling use case via the API?

## Creating a Code Insight

To create a Code Insight that will generate series data on the backend, use the mutation below.

Important: leave `dataSeries.repositoryScope.repositories` empty to specify that the query should be run across all repositories. Otherwise, no series data will be generated and saved on the backend.

TODO: What did we end up doing about the `lineColor`? I thought we had decided to go with hex values, but our frontend still uses the variables. Should the example be a hex value? Does it matter?

```gql
mutation CreateLineChartSearchInsight($input: LineChartSearchInsightInput!) {
  createLineChartSearchInsight(input: $input) {
    view {
      id
    }
  }
}
```

Example variables:

```json
{
  "input": {
    "options": {
      "title": "Javascript to Typescript migration"
    },
    "dataSeries": [{
      "query": "lang:javascript",
      "options": {
        "label": "javascript",
        "lineColor": "var(--oc-grape-7)"
      },
      "repositoryScope": {
        "repositories": []
      },
      "timeScope": {
        "stepInterval": {
          "unit": MONTH,
          "value": 1
        }
      }
    },
    {
      "query": "lang:typescript",
      "options": {
        "label": "typescript",
        "lineColor": "var(--oc-red-7)"
      },
      "repositoryScope": {
        "repositories": []
      },
      "timeScope": {
        "stepInterval": {
          "unit": MONTH,
          "value": 1
        }
      }
    }]
  }
}
```

## Read a single Code Insight

Use the query below to read a Code Insight by `id`. `filters` are optional, and only useful if you want to filter the results down to specific repositories.

Notes on the return object:

- The `dataSeries`, `dataSeriesDefinitions` and `seriesPresentation` arrays each store different information about the same series. The `seriesId` field on each can be used to match them up.
- `dataSeries.status` is useful to guage the progress of the series point generation. More information can be found in the API docs (TODO link)

```gql
query InsightViews($id: ID, $filters: InsightViewFiltersInput) {
  nodes {
    id,
    dataSeries {
      label,
      points {
        dateTime,
        value
      },
      status {
        pendingJobs,
        completedJobs,
        failedJobs,
        backfillQueuedAt
      }
    }
    dataSeriesDefinitions {
      ... on SearchInsightDataSeriesDefinition {
        seriesId,
        query,
        repositoryScope {
          repositories
        }
        timeScope {
          ... on InsightIntervalTimeScope {
            unit,
            value
          }
        }
      }
    }
    presentation {
      ... on LineChartInsightViewPresentation {
        title,
        seriesPresentation {
          seriesId,
          label,
          color
        }
      }
    }
  }
}
```

Example variables:

```json
{ "id": "insight-id-1",
  "filters": {
    "includeRepoRegex": "sourcegraph/sourcegraph",
    "excludeRepoRegex": "sourcegraph/handbook"
  } 
}
```

## List of Code Insights

The query below will list all of the Code Insights that you can see based on permissions. The query and return object is the same as for reading a single Code Insight. All input parameters are optional, and can be used for cursor-based pagination and repository filtering.

```gql
query InsightViews($first: int, $after: String, $filters: InsightViewFiltersInput) {
  nodes {
    id
  },
  pageInfo {
    endCursor,
    hasNextPage
  }
}
```

TODO: I don't think `hasNextPage` works right now. Should we leave it out?

Example variables:

```json
{
  "first": 10,
  "after": "endCursor-value",
}
```


## Update a Code Insight

Below is a GraphQL mutation that updates an existing Code Insight. The format is almost identical to the creation mutation, except that it now takes an `id`. The input object must be complete as it will completely overwrite the existing Code Insight.

Important: series are added, updated, and deleted from the insight as follows:

- A `dataSeries` without a `seriesId` or with an unknown `seriesId` will be treated as a new series and be added to the insight.
- A `dataSeries` with a `seriesId` that already exists for this insight will be updated in place.
- Any `dataSeries` that previously existed on the insight but are NOT included in the update **will be deleted.**

```gql
mutation UpdateLineChartSearchInsight($id: ID!, $input: UpdateLineChartSearchInsightInput!) {
  updateLineChartSearchInsight(id: $id, input: $input) {
    view {
      id
    }
  }
}
```

Example variables:

This is an example of updating the Code Insight from the creation mutation example. This updates `options.title`, the `dataSeries.timeScope.stepInterval` of the existing javascript series, and deletes the typescript series by omiting it.

```json
{
  "id": "insight-id-1",
  "input": {
    "options": {
      "title": "Javascript weekly"
    },
    "dataSeries": [{
      "seriesId": "series-id-1",
      "query": "lang:javascript",
      "options": {
        "label": "javascript",
        "lineColor": "var(--oc-grape-7)"
      },
      "repositoryScope": {
        "repositories": []
      },
      "timeScope": {
        "stepInterval": {
          "unit": WEEK,
          "value": 1
        }
      }
    }]
  }
}
```

## Delete a Code Insight

Below is a GraphQL query that deletes a Code Insight by ID.

```gql
mutation DeleteInsightView($id: ID!) {
  deleteInsightView(id: $id) {
    alwaysNil
  }
}
```

Example variables:

```json
{ "id": "insight-id-1" }
```
