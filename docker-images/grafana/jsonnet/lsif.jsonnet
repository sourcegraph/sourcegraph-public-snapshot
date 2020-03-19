local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local common = import './common.libsonnet';

// How long to look back for rate() queries
local timeRange = '1m';

// The duration percentiles to display
local percentiles = ['0.5', '0.9', '0.99'];

// Colors to pair to percentiles above (green, yellow, red)
local percentileColors = ['#7eb26d', '#cca300', '#bf1b00'];

// The histogram buckets for db queries and insertions
local buckets = ['0.2', '0.5', '1', '2', '5', '10', '30', '+Inf'];

// The histogram buckets for HTTP requests
local httpBuckets = ['0.2', '0.5', '1', '2', '5', '10', '30', '+Inf'];

// Colors to pair to buckets defined above (green to red)
local bucketColors = ['#96d98d', '#56a64b', '#37872d', '#e0b400', '#f2cc0c', '#ffee52', '#fa6400', '#c4162a'];

// The status code patterns for error responses
local httpPatterns = ['5..', '4..'];

// Colors to pair to the patterns above (red, yellow)
local errorColors = ['#bf1b00', '#cca300'];

// The classes of jobs performed by the LSIF dump processor
local jobClasses = ['convert', 'update-tips'];

//
// Utils

// Title-case a single word.
local titleCase(val) = '%s%s' % [std.asciiUpper(std.substr(val, 0, 1)), std.substr(val, 1, std.length(val) - 1)];

//
// Standard Panels

// Apply defaults defined above to panel constructors
local makeHttpTotalRequestsPanel(titleValue, metricValue, metricFilter='') = common.makeHttpTotalRequestsPanel(titleValue, metricValue, metricFilter=metricFilter, timeRange=timeRange);
local makeHttpRequestsPanel(titleValue, metricValue, metricFilter='') = common.makeHttpRequestsPanel(titleValue, metricValue, metricFilter=metricFilter, timeRange=timeRange, buckets=httpBuckets, colors=bucketColors);
local makeHttpErrorRatePanel(titleValue, metricValue, metricFilter='') = common.makeHttpErrorRatePanel(titleValue, metricValue, metricFilter=metricFilter, timeRange=timeRange, patterns=httpPatterns, colors=errorColors);
local makeHttpDurationPercentilesPanel(titleValue, metricValue, metricFilter='') = common.makeHttpDurationPercentilesPanel(titleValue, metricValue, metricFilter=metricFilter, timeRange=timeRange, percentiles=percentiles, colors=percentileColors);
local makeRequestsPanel(titleValue, metricValue, metricFilter='') = common.makeRequestsPanel(titleValue, metricValue, metricFilter=metricFilter, timeRange=timeRange, buckets=buckets, colors=bucketColors);
local makeErrorRatePanel(titleValue, metricValue, metricFilter='') = common.makeErrorRatePanel(titleValue, metricValue, metricFilter=metricFilter, timeRange=timeRange);
local makeDurationPercentilesPanel(titleValue, metricValue, metricFilter='') = common.makeDurationPercentilesPanel(titleValue, metricValue, metricFilter=metricFilter, timeRange=timeRange, percentiles=percentiles, colors=percentileColors);

// Make panels
local httpUploadTotalRequestsPanel = makeHttpTotalRequestsPanel(titleValue='upload requests', metricValue='lsif_http_upload_request');
local httpUploadRequestsPanel = makeHttpRequestsPanel(titleValue='upload requests', metricValue='lsif_http_upload_request');
local httpUploadErrorRatePanel = makeHttpErrorRatePanel(titleValue='Upload', metricValue='lsif_http_upload_request');
local httpUploadDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='Upload request', metricValue='lsif_http_upload_request');
local httpQueryTotalRequestsPanel = makeHttpTotalRequestsPanel(titleValue='query requests', metricValue='lsif_http_query_request');
local httpQueryRequestsPanel = makeHttpRequestsPanel(titleValue='query requests', metricValue='lsif_http_query_request');
local httpQueryErrorRatePanel = makeHttpErrorRatePanel(titleValue='Query', metricValue='lsif_http_query_request');
local httpQueryDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='Query request', metricValue='lsif_http_query_request');
local databaseQueryRequestsPanel = makeRequestsPanel(titleValue='database queries', metricValue='lsif_database_query');
local databaseQueryErrorRatePanel = makeErrorRatePanel(titleValue='Database query', metricValue='lsif_database_query');
local databaseQueryDurationPercentilesPanel = makeDurationPercentilesPanel(titleValue='Database query', metricValue='lsif_database_query');
local xrepoQueryRequestsPanel = makeRequestsPanel(titleValue='cross-repository queries', metricValue='lsif_xrepo_query');
local xrepoQueryErrorRatePanel = makeErrorRatePanel(titleValue='Cross-repository query', metricValue='lsif_xrepo_query');
local xrepoQueryDurationPercentilesPanel = makeDurationPercentilesPanel(titleValue='Cross-repository query', metricValue='lsif_xrepo_query');
local databaseInsertionRequestsPanel = makeRequestsPanel(titleValue='database insertions', metricValue='lsif_database_insertion');
local databaseInsertionErrorRatePanel = makeErrorRatePanel(titleValue='Database insertion', metricValue='lsif_database_insertion');
local databaseInsertionDurationPercentilesPanel = makeDurationPercentilesPanel(titleValue='Database query', metricValue='lsif_database_insertion');
local xrepoInsertionRequestsPanel = makeRequestsPanel(titleValue='cross-repository insertions', metricValue='lsif_xrepo_insertion');
local xrepoInsertionErrorRatePanel = makeErrorRatePanel(titleValue='Cross-repository insertion', metricValue='lsif_xrepo_insertion');
local xrepoInsertionDurationPercentilesPanel = makeDurationPercentilesPanel(titleValue='Cross-repository insertion', metricValue='lsif_xrepo_insertion');

//
// Process Metrics

local cpuPanel = common.makePanel(
  title='CPU utilization',
  targets=[
    prometheus.target('sum(rate(lsif_process_cpu_user_seconds_total[1m])) by (instance)', legendFormat='user, {{instance}}'),
    prometheus.target('sum(rate(lsif_process_cpu_system_seconds_total[1m])) by (instance)', legendFormat='system, {{instance}}'),
    prometheus.target('sum(rate(lsif_process_cpu_seconds_total[1m])) by (instance)', legendFormat='combined, {{instance}}'),
  ],
);

local memoryPanel = common.makePanel(
  title='Average memory usage',
  extra={
    yaxes: common.makeYAxes({
      format: 'decmbytes',
    }),
  },
  targets=[
    prometheus.target('avg(lsif_nodejs_external_memory_bytes / 1024 / 1024) by (instance)', legendFormat='memory usage of {{instance}}'),
  ],
);

//
// Queue and Jobs Panels

local queueSizePanel = common.makePanel(
  title='Queue size',
  targets=[
    prometheus.target('lsif_queue_size', legendFormat='queue size'),
  ],
);

local durationPanelsByJob = std.flattenArrays(std.map(function(class) [
  makeRequestsPanel(titleValue='%s jobs' % class, metricValue='lsif_job', metricFilter='class="%s"' % class),
  makeDurationPercentilesPanel(titleValue='%s jobs' % titleCase(class), metricValue='lsif_job', metricFilter='class="%s"' % class),
], jobClasses));

//
// Cache Panels

local makeCacheTargets(metricValue) = [
  prometheus.target('rate(%s_events_total{type="hit"}[%s])' % [metricValue, timeRange], legendFormat='hits'),
  prometheus.target('rate(%s_events_total{type="miss"}[%s])' % [metricValue, timeRange], legendFormat='misses'),
];

local makeCacheEvictionTargets(metricValue) = [
  prometheus.target('rate(%s_events_total{type="eviction"}[%s])' % [metricValue, timeRange], legendFormat='evictions'),
  prometheus.target('rate(%s_events_total{type="locked-eviction"}[%s])' % [metricValue, timeRange], legendFormat='locked entries'),
];

local cacheUtilizationPanel = common.makePanel(
  title='Cache utilization',
  extra={
    max: 1,
    yaxes: common.makeYAxes({
      max: 1,
      format: 'percentunit',
    }),
    tooltip: {
      common: false,
    },
  },
  targets=[
    prometheus.target('lsif_connection_cache_size / lsif_connection_cache_capacity', legendFormat='connection cache utilization'),
    prometheus.target('lsif_document_cache_size / lsif_document_cache_capacity', legendFormat='document cache utilization'),
    prometheus.target('lsif_result_chunk_cache_size / lsif_result_chunk_cache_capacity', legendFormat='result chunk cache utilization'),
  ],
);

local connectionCacheEventsPanel = common.makePanel(
  title='Connection cache events',
  targets=makeCacheTargets('lsif_connection_cache') + makeCacheEvictionTargets('lsif_connection_cache')
);

local documentCacheEventsPanel = common.makePanel(
  title='Document cache events',
  targets=makeCacheTargets('lsif_document_cache') + makeCacheEvictionTargets('lsif_document_cache')
);

local resultChunkCacheEventsPanel = common.makePanel(
  title='Result chunk cache events',
  targets=makeCacheTargets('lsif_result_chunk_cache') + makeCacheEvictionTargets('lsif_result_chunk_cache')
);

local bloomFilterEventsPanel = common.makePanel(
  title='Bloom filter events',
  targets=makeCacheTargets('lsif_bloom_filter')
);

//
// Dashboard Construction

common.makeDashboard(title='LSIF')
.addRow(title='Process metrics', panels=[cpuPanel, memoryPanel])
.addRow(title='Upload requests', panels=[httpUploadTotalRequestsPanel, httpUploadRequestsPanel, httpUploadErrorRatePanel, httpUploadDurationPercentilesPanel])
.addRow(title='Query requests', panels=[httpQueryTotalRequestsPanel, httpQueryRequestsPanel, httpQueryErrorRatePanel, httpQueryDurationPercentilesPanel])
.addRow(title='Queue and job stats', panels=[queueSizePanel] + durationPanelsByJob)
.addRow(title='Database queries', panels=[databaseQueryRequestsPanel, databaseQueryErrorRatePanel, databaseQueryDurationPercentilesPanel])
.addRow(title='Cross-repository queries', panels=[xrepoQueryRequestsPanel, xrepoQueryErrorRatePanel, xrepoQueryDurationPercentilesPanel])
.addRow(title='Database insertions', panels=[databaseInsertionRequestsPanel, databaseInsertionErrorRatePanel, databaseInsertionDurationPercentilesPanel])
.addRow(title='Cross-repository insertions', panels=[xrepoInsertionRequestsPanel, xrepoInsertionErrorRatePanel, xrepoInsertionDurationPercentilesPanel])
.addRow(title='Caches and Filters', panels=[cacheUtilizationPanel, connectionCacheEventsPanel, documentCacheEventsPanel, resultChunkCacheEventsPanel, bloomFilterEventsPanel])
