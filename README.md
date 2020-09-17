# search-blitz

The purpose of search-blitz is to provide a baseline for our search performance. The appraoch is to call 
Sourcegraph.com for typical queries in regular intervals collecting performance metrics relevant for search.
The set of queries is fixed. The queries are segmented by type and performance metrics are available for each segment.
The data is available as logs (stream and file) as well as in aggregated form on prometheus.

## How to track your query
1. create a new `.txt` file in `./data`
    - give it a name that is descriptive for the type of queries you want to track. The name is used as label for prometheus metrics.
    - add 1 query per line
2. [deploy](#how-to-deploy)

## How to deploy
1. Merge your changes to _main_
2. We don't use CI for search-blitz, which means we build and upload a new docker image manually.
    ```
    ./scripts/build.sh <next-version, e.g. 0.0.2>
    ```
3.  Update the image tag in [deploy-sourcegraph-dot-com](https://github.com/sourcegraph/deploy-sourcegraph-dot-com/blob/release/configure/search-blitz/search-blitz.StatefulSet.yaml#L36)

## How to access the Grafana dashboard
```
kubectl port-forward search-blitz-0 3000:3000 -n monitoring
```

open http://localhost:3000

## How to download logs

Logs are rotated every 10 mb and stored in `search-blitz-data/logs` on the persistent volume which is attached to all containers. 

```
kubectl cp monitoring/search-blitz-0:search-blitz-data/logs -c search-blitz .
```
