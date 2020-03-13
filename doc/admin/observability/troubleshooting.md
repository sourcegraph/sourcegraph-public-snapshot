# Troubleshooting guide

If you're troubleshooting a specific production instance in Sourcegraph, we recommend starting with
the following resources:

1. [View verbose logs](index.md#logs) (most common)
1. [Inspect traces](tracing.md#inspecting-traces-jaeger-or-lightstep)
1. [Inspect the Go net/trace information](tracing.md#viewing-go-net-trace-information) for individual services (rarely needed)

## Reporting Sourcegraph search timeouts

If your users are experiencing search timeouts or search performance issues, please perform the following steps:

1. Access Grafana directly by [following these steps](metrics.md#accessing-grafana-directly).
2. Select the **+** icon on the left-hand side, then choose **Import**.
3. Paste [this JSON](https://gist.githubusercontent.com/slimsag/3fcc134f5ce09728188b94b463131527/raw/f8b545f4ce14b0c30a93f05cd1ee469594957a2c/sourcegraph-debug-search-timeouts.json) into the input and click **Load**.

Once the dashboard appears, please send us screenshots of **the entire** dashboard.
