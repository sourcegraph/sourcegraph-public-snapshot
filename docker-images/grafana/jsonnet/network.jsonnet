local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local common = import './common.libsonnet';

// The duration percentiles to display
local percentiles = ['0.5', '0.9', '0.99'];

// Colors to pair to percentiles above (red, yellow, green)
local percentileColors = ['#7eb26d', '#cca300', '#bf1b00'];

// The histogram buckets for db queries and insertions
local buckets = ['0.2', '0.5', '1', '2', '5', '10', '30', '+Inf'];

// The histogram buckets for HTTP requests
local httpBuckets = ['0.03', '0.1', '0.3', '1.5', '10', '+Inf'];

// Colors to pair to buckets defined above (green to red)
local bucketColors = ['#96d98d', '#56a64b', '#37872d', '#e0b400', '#f2cc0c', '#ffee52', '#fa6400', '#c4162a'];

// The status code patterns for error responses
local httpPatterns = ['5..', '4..'];

// Colors to pair to the patterns above (red, yellow)
local errorColors = ['#7eb26d', '#cca300'];

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
            "definition": "label_values(src_gitserver_request_duration_seconds_bucket, host)",
            "hide": 0,
            "includeAll": true,
            "label": null,
            "multi": true,
            "name": "gitserver",
            "options": [],
            "query": "label_values(src_gitserver_request_duration_seconds_bucket, host)",
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
            "allValue": null,
            "current": {
              "text": "All",
              "value": ["$__all"]
            },
            "datasource": "Prometheus",
            "definition": "label_values(src_gitserver_request_duration_seconds_bucket, instance)",
            "hide": 0,
            "includeAll": true,
            "label": null,
            "multi": true,
            "name": "client_instance",
            "options": [],
            "query": "label_values(src_gitserver_request_duration_seconds_bucket, instance)",
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
            "allValue": null,
            "current": {
              "text": "All",
              "value": ["$__all"]
            },
            "datasource": "Prometheus",
            "definition": "label_values(src_gitserver_request_duration_seconds_bucket, job)",
            "hide": 0,
            "includeAll": true,
            "label": null,
            "multi": true,
            "name": "client_job",
            "options": [],
            "query": "label_values(src_gitserver_request_duration_seconds_bucket, job)",
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

// How long to look back for rate() queries
local timeRange = '$interval';

local gitServerMetricFilter = 'host=~"$gitserver", instance=~"$client_instance", job=~"$client_job"';
local clientMetricFilter = 'instance=~"$client_instance", job=~"$client_job"';


//
// Standard Panels

// Apply defaults defined above to panel constructors
local makeHttpRequestsPanel(titleValue, metricValue, metricFilter='') = common.makeHttpRequestsPanel(titleValue, metricValue, timeRange=timeRange, buckets=httpBuckets, colors=bucketColors, metricFilter=metricFilter);
local makeHttpErrorRatePanel(titleValue, metricValue, metricFilter='') = common.makeHttpErrorRatePanel(titleValue, metricValue, timeRange=timeRange, patterns=httpPatterns, colors=errorColors, metricFilter=metricFilter);
local makeHttpDurationPercentilesPanel(titleValue, metricValue, metricFilter='') = common.makeHttpDurationPercentilesPanel(titleValue, metricValue, timeRange=timeRange, percentiles=percentiles, colors=percentileColors, metricFilter=metricFilter);
local makeRequestsPanel(titleValue, metricValue) = common.makeRequestsPanel(titleValue, metricValue, timeRange=timeRange, buckets=buckets, colors=bucketColors);
local makeErrorRatePanel(titleValue, metricValue) = common.makeErrorRatePanel(titleValue, metricValue, timeRange=timeRange);
local makeDurationPercentilesPanel(titleValue, metricValue) = common.makeDurationPercentilesPanel(titleValue, metricValue, timeRange=timeRange, percentiles=percentiles, colors=percentileColors);

local gitserverRequestsPanel = makeHttpRequestsPanel(titleValue='Gitserver requests', metricValue='src_gitserver_request', metricFilter=gitServerMetricFilter);
local gitserverErrorRatePanel = makeHttpErrorRatePanel(titleValue='Gitserver', metricValue='src_gitserver_request', metricFilter=gitServerMetricFilter);
local gitserverDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='Gitserver request', metricValue='src_gitserver_request', metricFilter=gitServerMetricFilter);

local gitserverDeadlineExceededRatePanel = common.makePanel(
    title='Deadline Exceeded Rate',
    targets=[
      prometheus.target(
        "rate(src_gitserver_client_deadline_exceeded[$interval])",
        legendFormat='Gitserver Deadline Exceeded Rate',
      ),
    ]
);


local repoupdaterRequestsPanel = makeHttpRequestsPanel(titleValue='Repoupdater requests', metricValue='src_repoupdater_request', metricFilter=clientMetricFilter);
local repoupdaterErrorRatePanel = makeHttpErrorRatePanel(titleValue='Repoupdater', metricValue='src_repoupdater_request', metricFilter=clientMetricFilter);
local repoupdaterDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='Repoupdater request', metricValue='src_repoupdater_request', metricFilter=clientMetricFilter);

local frontendInternalRequestsPanel = makeHttpRequestsPanel(titleValue='frontend_internal requests', metricValue='src_frontend_internal_request', metricFilter=clientMetricFilter);
local frontendInternalErrorRatePanel = makeHttpErrorRatePanel(titleValue='frontend_internal', metricValue='src_frontend_internal_request', metricFilter=clientMetricFilter);
local frontendInternalDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='frontend_internal request', metricValue='src_frontend_internal_request', metricFilter=clientMetricFilter);

local textsearchRequestsPanel = makeHttpRequestsPanel(titleValue='textsearch requests', metricValue='src_textsearch_request', metricFilter=clientMetricFilter);
local textsearchErrorRatePanel = makeHttpErrorRatePanel(titleValue='textsearch', metricValue='src_textsearch_request', metricFilter=clientMetricFilter);
local textsearchDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='textsearch request', metricValue='src_textsearch_request', metricFilter=clientMetricFilter);

local zoektRequestsPanel = makeHttpRequestsPanel(titleValue='zoekt requests', metricValue='src_zoekt_request', metricFilter=clientMetricFilter);
local zoektErrorRatePanel = makeHttpErrorRatePanel(titleValue='zoekt', metricValue='src_zoekt_request', metricFilter=clientMetricFilter);
local zoektDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='zoekt request', metricValue='src_zoekt_request', metricFilter=clientMetricFilter);

//
// Dashboard Construction

common.makeDashboard(title='Cluster-Internal Network Activity, Client POV', extra=dashboardTemplatingVars)
.addRow(title='Requests to Repoupdater', panels=[frontendInternalRequestsPanel, frontendInternalErrorRatePanel, frontendInternalDurationPercentilesPanel])
.addRow(title='Requests to Repoupdater', panels=[repoupdaterRequestsPanel, repoupdaterErrorRatePanel, repoupdaterDurationPercentilesPanel])
.addRow(title='Requests to Text Search', panels=[textsearchRequestsPanel, textsearchErrorRatePanel, textsearchDurationPercentilesPanel])
.addRow(title='Requests to Zoekt', panels=[zoektRequestsPanel, zoektErrorRatePanel, zoektDurationPercentilesPanel])
.addRow(title='Requests to Gitserver', panels=[gitserverRequestsPanel, gitserverErrorRatePanel, gitserverDurationPercentilesPanel])
.addRow(title='Requests to Gitserver', panels=[gitserverDeadlineExceededRatePanel])
