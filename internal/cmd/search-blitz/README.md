# search-blitz

The purpose of search-blitz is to provide a baseline for our search performance. Search-blitz calls the GraphQL API of
Sourcegraph.com for typical queries in regular intervals collecting performance metrics relevant for search.
The set of queries is fixed. The queries are segmented by type, and performance metrics are available for each segment
on dedicated instances of Prometheus and Grafana. The raw data, including links to Jaeger traces,
are available as logs (stream and file).

## How to track your query

1. Create a new file in `./data`
   - Give it a name that is descriptive for the type of queries you want to track. The file name without extension is used as label for prometheus metrics.
   - Add 1 query per line.
2. [Deploy](#how-to-deploy)

## How to deploy a new version

1. Merge your changes to _main_
2. Build and upload a new docker image:
   ```
   ./scripts/build.sh <next-version, e.g. 0.0.2>
   ```
3. Update the image tag in [deploy-sourcegraph-dot-com](https://github.com/sourcegraph/deploy-sourcegraph-dot-com/blob/release/configure/search-blitz/search-blitz.StatefulSet.yaml#L36)
   ```
   ./scripts/update-deploy-sourcegraph-dot-com.sh <next-version, e.g. 0.0.2>
   ```

## How to access the Grafana dashboard

```
kubectl port-forward search-blitz-0 3000:3000 -n monitoring
```

open http://localhost:3000

Credentials: 1password.

## How to download logs

Logs are stored in `search-blitz-data/logs` on the persistent volume which is attached to all containers.

```
kubectl cp monitoring/search-blitz-0:search-blitz-data/logs -c search-blitz .
```

Logs are rotated every 10 mb.

## Search-blitz user

search-blitz calls the GraphQL API with a token of a dedicated user.

Credentials: 1password.
