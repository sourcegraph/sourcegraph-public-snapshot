local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local common = import './common.libsonnet';

// The duration percentiles to display
local percentiles = ['0.5', '0.9', '0.99'];

// Colors to pair to percentiles above (red, yellow, green)
local percentileColors = ['#7eb26d', '#cca300', '#bf1b00'];

// The histogram buckets for GraphQL requests
local graphQLBuckets = ['0.03', '0.1', '0.3', '1.5', '10', '+Inf'];

// Colors to pair to buckets defined above (green to red)
local bucketColors = ['#96d98d', '#56a64b', '#37872d', '#e0b400', '#f2cc0c', '#ffee52', '#fa6400', '#c4162a'];

// Colors to pair to the patterns above (red, yellow)
local errorColors = ['#7eb26d', '#cca300'];

// How long to look back for rate() queries
local timeRange = '$interval';

local dashboardTemplatingVars =  {
    "templating": {
        "list": [
          {
            "allValue": null,
            "current": {
              "text": "All",
              "value": ["$__all"]
            },
            "datasource": "Prometheus",
            "definition": "label_values(src_graphql_field_seconds_count, instance)",
            "hide": 0,
            "includeAll": true,
            "label": null,
            "multi": true,
            "name": "server_instance",
            "options": [],
            "query": "label_values(src_graphql_field_seconds_count, instance)",
            "refresh": 2,
            "regex": "",
            "skipUrlSync": false,
            "sort": 1,
            "tagValuesQuery": "",
            "tags": [],
            "tagsQuery": "",
            "type": "query",
            "useTags": false
          },
          {
            "auto": false,
            "auto_count": 30,
            "auto_min": "10s",
            "current": {
              "text": "5m",
              "value": "5m"
            },
            "datasource": null,
            "hide": 0,
            "includeAll": false,
            "label": "",
            "multi": false,
            "name": "interval",
            "options": [
              {
                "selected": false,
                "text": "1m",
                "value": "1m"
              },
              {
                "selected": true,
                "text": "5m",
                "value": "5m"
              },
              {
                "selected": false,
                "text": "10m",
                "value": "10m"
              },
              {
                "selected": false,
                "text": "30m",
                "value": "30m"
              },
              {
                "selected": false,
                "text": "1h",
                "value": "1h"
              }
            ],
            "query": "1m,5m,10m,30m,1h",
            "refresh": 2,
            "type": "interval"
          }
        ]
     }
};

local graphQLRequestsPanel = common.makePanel(
   title='GraphQL Request Rate',
   targets=[
     prometheus.target(
       'rate(src_graphql_field_seconds_count{instance=~"$server_instance"}[$interval])',
       legendFormat='requests',
     ),
   ]
 );

local graphQLErrorRatePanel = common.makePanel(
  title='GraphQL Error Rate',
  targets=[
    prometheus.target(
      'rate(src_graphql_field_seconds_count{instance=~"$server_instance", error="true"}[$interval])',
      legendFormat='errors',
    ),
  ]
);

local graphQLDurationPercentilesPanel = common.makePanel(
    title='GraphQL requests duration percentiles',
    extra={
      yaxes: common.makeYAxes({ format: 's' }),
      seriesOverrides: common.makePercentileSeriesOverrides(percentiles, percentileColors),
    },
    targets=std.map(
      function(percentile) prometheus.target(
        'histogram_quantile(%s, rate(src_graphql_field_seconds_bucket{instance=~"$server_instance"}[$interval]))' % [
          percentile,
        ],
        legendFormat='%sp' % percentile,
      ),
      percentiles
    )
);

//
// Dashboard Construction

common.makeDashboard(title='GraphQL Requests', extra=dashboardTemplatingVars)
.addRow(title='Requests', panels=[graphQLRequestsPanel])
.addRow(title='Errors', panels=[graphQLErrorRatePanel])
.addRow(title='Duration', panels=[graphQLDurationPercentilesPanel])


