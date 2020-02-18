# Reporting Sourcegraph search timeouts

If your users are experiencing search timeouts, please send us screenshots **for each of the four pages linked below**, replacing `https://sourcegraph.example.com` with your actual Sourcegraph URL.

Each of these will show us the total number of searches in the past 7d resulting in:

#### 1) Successful searches (no errors or timeouts)

https://sourcegraph.example.com/-/debug/grafana/explore?orgId=1&left=%5B%22now-7d%22,%22now%22,%22Prometheus%22,%7B%22expr%22:%22sum%20by%20(status)(src_graphql_search_response%7Bstatus%3D%5C%22success%5C%22%7D)%22,%22context%22:%22explore%22%7D,%7B%22mode%22:%22Metrics%22%7D,%7B%22ui%22:%5Btrue,true,true,%22none%22%5D%7D%5D

#### 2) Searches ending in errors and timeouts (no results returned)

https://sourcegraph.example.com/-/debug/grafana/explore?orgId=1&left=%5B%22now-7d%22,%22now%22,%22Prometheus%22,%7B%22expr%22:%22sum%20by%20(status)(src_graphql_search_response%7Bstatus!~%5C%22error%7Ctimeout%5C%22%7D)%22,%22context%22:%22explore%22%7D,%7B%22mode%22:%22Metrics%22%7D,%7B%22ui%22:%5Btrue,true,true,%22none%22%5D%7D%5D

#### 3) Searches ending in partial timeouts (some results returned)

https://sourcegraph.example.com/-/debug/grafana/explore?orgId=1&left=%5B%22now-7d%22,%22now%22,%22Prometheus%22,%7B%22expr%22:%22sum%20by%20(status)(src_graphql_search_response%7Bstatus%3D%5C%22partial_timeout%5C%22%7D)%22,%22context%22:%22explore%22%7D,%7B%22mode%22:%22Metrics%22%7D,%7B%22ui%22:%5Btrue,true,true,%22none%22%5D%7D%5D

#### 4) Searches ending in a user suggestion alert (no results returned)

https://sourcegraph.example.com/-/debug/grafana/explore?orgId=1&left=%5B%22now-7d%22,%22now%22,%22Prometheus%22,%7B%22expr%22:%22sum%20by%20(status,%20alert_type)(src_graphql_search_response%7Bstatus%3D%5C%22alert%5C%22%7D)%22,%22context%22:%22explore%22%7D,%7B%22mode%22:%22Metrics%22%7D,%7B%22ui%22:%5Btrue,true,true,%22none%22%5D%7D%5D
