local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local common = import './common.libsonnet';

// How long to look back for rate() queries
local timeRange = '1m';

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

//
// Standard Panels

// Apply defaults defined above to panel constructors
local makeHttpRequestsPanel(titleValue, metricValue) = common.makeHttpRequestsPanel(titleValue, metricValue, timeRange=timeRange, buckets=httpBuckets, colors=bucketColors);
local makeHttpErrorRatePanel(titleValue, metricValue) = common.makeHttpErrorRatePanel(titleValue, metricValue, timeRange=timeRange, patterns=httpPatterns, colors=errorColors);
local makeHttpDurationPercentilesPanel(titleValue, metricValue) = common.makeHttpDurationPercentilesPanel(titleValue, metricValue, timeRange=timeRange, percentiles=percentiles, colors=percentileColors);
local makeRequestsPanel(titleValue, metricValue) = common.makeRequestsPanel(titleValue, metricValue, timeRange=timeRange, buckets=buckets, colors=bucketColors);
local makeErrorRatePanel(titleValue, metricValue) = common.makeErrorRatePanel(titleValue, metricValue, timeRange=timeRange);
local makeDurationPercentilesPanel(titleValue, metricValue) = common.makeDurationPercentilesPanel(titleValue, metricValue, timeRange=timeRange, percentiles=percentiles, colors=percentileColors);

local gitserverRequestsPanel = makeHttpRequestsPanel(titleValue='Gitserver requests', metricValue='src_gitserver_request');
local gitserverErrorRatePanel = makeHttpErrorRatePanel(titleValue='Gitserver', metricValue='src_gitserver_request');
local gitserverDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='Gitserver request', metricValue='src_gitserver_request');

local repoupdaterRequestsPanel = makeHttpRequestsPanel(titleValue='Repoupdater requests', metricValue='src_repoupdater_request');
local repoupdaterErrorRatePanel = makeHttpErrorRatePanel(titleValue='Repoupdater', metricValue='src_repoupdater_request');
local repoupdaterDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='Repoupdater request', metricValue='src_repoupdater_request');

//
// Dashboard Construction

common.makeDashboard(title='Cluster-Internal Network Activity')
.addRow(title='Requests to Gitserver', panels=[gitserverRequestsPanel, gitserverErrorRatePanel, gitserverDurationPercentilesPanel])
.addRow(title='Requests to Repoupdater', panels=[repoupdaterRequestsPanel, repoupdaterErrorRatePanel, repoupdaterDurationPercentilesPanel])
