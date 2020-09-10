# search-blitz

The purpose of search-blitz is to provide a baseline for our search performance. The appraoch is to call 
Sourcegraph.com for typical queries in regular intervals collecting performance metrics relevant for search.
The set of queries is fixed. The queries are segmented by type and performance metrics are available for each segment.
The data is available as logs (stream and file) as well as in aggregated form on prometheus.


