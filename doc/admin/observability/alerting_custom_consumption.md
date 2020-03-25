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

## Total number of critical alerts firing currently

Prometheus query:

```prometheus
sum(alert_count{level="critical"})
```

Example `curl` query:

```sh
curl http://$PROMETHEUS_URL/api/v1/query?query=sum%28alert_count%7Blevel%3D%22critical%22%7D%29
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
          1585177196.683,
          "0"
        ]
      }
    ]
  }
}
```

This shows that "0" critical alerts are currently firing.

## Total number of warning alerts firing currently

Prometheus query:

```prometheus
sum(alert_count{level="warning"})
```

Example `curl` query:

```sh
curl http://$PROMETHEUS_URL/api/v1/query?query=sum%28alert_count%7Blevel%3D%22warning%22%7D%29
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
          1585177540.114,
          "12"
        ]
      }
    ]
  }
}
```

This shows that "12" warning alerts are currently firing.

