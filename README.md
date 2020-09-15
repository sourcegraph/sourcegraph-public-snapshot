# search-blitz

The purpose of search-blitz is to provide a baseline for our search performance. The appraoch is to call 
Sourcegraph.com for typical queries in regular intervals collecting performance metrics relevant for search.
The set of queries is fixed. The queries are segmented by type and performance metrics are available for each segment.
The data is available as logs (stream and file) as well as in aggregated form on prometheus.


## Deployment
1. Commit your changes to _main_
2. We don't use CI for search-blitz, which means we build and upload a new docker image manually 
    ```
    ./scripts/build.sh <next-version, e.g. 0.0.2>
    ```
3.  Update the image tag in [deploy-sourcegraph-dot-com](https://github.com/sourcegraph/deploy-sourcegraph-dot-com/blob/release/configure/search-blitz/search-blitz.StatefulSet.yaml#L36)

