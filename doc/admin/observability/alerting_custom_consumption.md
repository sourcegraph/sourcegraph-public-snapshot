# Custom consumption of Sourcegraph alerts

If Sourcegraph's builtin [alerting](alerting.md) (which can notify you via email, Slack, PagerDuty, webhook, and more) is not sufficient for you, or if you just prefer to consume the alerts programatically for some reason, then this page is for you.

Below are examples of how to query alerts that are being monitored or are currently firing in Sourcegraph.

## The difference between "critical" and "warning" alerts

Please note that there is a key difference between Sourcegraph's "critical" and "warning" alerts:

- _Critical_ alerts are guaranteed to be a real issue with Sourcegraph.
  - If you see one, it means something is definitely wrong.
  - We suggest e.g. emailing the site admin when these occur.
- _Warning_ alerts are worth looking into, but may not be a real issue with Sourcegraph.
  - We suggest checking in on these periodically, or using a notification channel that will not bother anyone if it is spammed.
  - If you see warning alerts firing, please let us know so that we can improve them.
  - Over time, as warning alerts become stable and reliable across many Sourcegraph deployments, they will also be promoted to critical alerts in an update by Sourcegraph.

For more information about the differences, see the: [high level alerting metrics](metrics_guide.md#high-level-alerting-metrics)

## "How many critical alerts were firing in the last minute?"

Prometheus query:

```prometheus
sum(max by (service_name,name,description)(max_over_time(alert_count{level="critical",name!=""}[1m])))
```

Example `curl` query:

```sh
curl 'http://$PROMETHEUS_URL/api/v1/query?query=sum%28max%20by%20%28service_name%2Cname%2Cdescription%29%28max_over_time%28alert_count%7Blevel%3D%22critical%22%2Cname%21%3D%22%22%7D%5B1m%5D%29%29%29
```

Example response:

```json
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {},
        "value": [
          1585250319.243,
          "0"
        ]
      }
    ]
  }
}
```

This only ever returns a single result, representing the maximum number of critical alerts firing across all Sourcegraph services in the last minute (relative to the time the query executed / the returned unix timestamp `1585250319.243`). The above shows that `"0"` alerts were firing, and if the number was non-zero, it would represent the max number of alerts firing across all services in the last minute.

## "How many critical alerts were firing in the last minute, per service?"

Prometheus query:

```prometheus
sum by (service_name)(max by (service_name,name,description)(max_over_time(alert_count{level="critical",name!=""}[1m])))
```

Example `curl` query:

```sh
curl 'http://$PROMETHEUS_URL/api/v1/query?query=sum%20by%20%28service_name%29%28max%20by%20%28service_name%2Cname%2Cdescription%29%28max_over_time%28alert_count%7Blevel%3D%22critical%22%2Cname%21%3D%22%22%7D%5B1m%5D%29%29%29'
```

Example response:

```json
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "service_name": "frontend"
        },
        "value": [
          1585250083.874,
          "0"
        ]
      },
      {
        "metric": {
          "service_name": "zoekt-indexserver"
        },
        "value": [
          1585250083.874,
          "0"
        ]
      },
      ...
    ]
  }
}
```

This returns a `result` for each service of Sourcegraph where at least one critical alert is defined and being monitored. For example, the first result (`frontend`) indicates that in the last minute (relative to the time the query executed / the returned unix timestamp `1585250083.874`) that `"0"` alerts for the `frontend` service were firing. If the number was non-zero, it would represent the max number of alerts firing on that service in the last minute.


## "How many critical alerts were firing in the last minute, per defined alert?"

Prometheus query:

```prometheus
max by (service_name,name,description)(max_over_time(alert_count{level="critical",name!=""}[1m]))
```

Example `curl` query:

```sh
curl 'http://$PROMETHEUS_URL/api/v1/query?query=max%20by%20%28service_name%2Cname%2Cdescription%29%28max_over_time%28alert_count%7Blevel%3D%22critical%22%2Cname%21%3D%22%22%7D%5B1m%5D%29%29'
```

Example response:

```json
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "description": "gitserver: 50+ concurrent command executions (abnormally high load)",
          "name": "high_concurrent_execs",
          "service_name": "gitserver"
        },
        "value": [
          1585249844.475,
          "0"
        ]
      },
      {
        "metric": {
          "description": "gitserver: 100+ concurrent command executions (abnormally high load)",
          "name": "high_concurrent_execs",
          "service_name": "gitserver"
        },
        "value": [
          1585249844.475,
          "0"
        ]
      },
      ...
    ]
  }
}
```

This returns a `result` for each defined critical alert that Sourcegraph is monitoring. For example, the first result (`high_concurrent_execs`) indicates that in the last minute (relative to the time the query executed / the returned unix timestamp `1585249844.475`) that `"0"` alerts were firing. Any value >= 1 here, would indicate that alert has fired in the last minute.

## Warning alerts

If you do wish to query warning alerts, too, then simply replace `critical` with `warning` in any of the above query examples.
