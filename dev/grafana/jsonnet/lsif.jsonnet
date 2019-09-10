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

// The histogram bucket sfor HTTP requests
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

// Make panels
local httpRequestsPanel = makeHttpRequestsPanel(titleValue='server requests', metricValue='http_request');
local httpErrorRatePanel = makeHttpErrorRatePanel(titleValue='Server', metricValue='http_request');
local httpDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='Server request', metricValue='http_request');
local databaseQueryRequestsPanel = makeRequestsPanel(titleValue='database queries', metricValue='lsif_database_query');
local databaseQueryErrorRatePanel = makeErrorRatePanel(titleValue='Database query', metricValue='lsif_database_query');
local databaseQueryDurationPercentilesPanel = makeDurationPercentilesPanel(titleValue='Database query', metricValue='lsif_database_query');
local xrepoQueryRequestsPanel = makeRequestsPanel(titleValue='xrepo queries', metricValue='lsif_xrepo_query');
local xrepoQueryErrorRatePanel = makeErrorRatePanel(titleValue='Xrepo query', metricValue='lsif_xrepo_query');
local xrepoQueryDurationPercentilesPanel = makeDurationPercentilesPanel(titleValue='Xrepo query', metricValue='lsif_xrepo_query');
local databaseInsertionRequestsPanel = makeRequestsPanel(titleValue='database insertions', metricValue='lsif_database_insertion');
local databaseInsertionErrorRatePanel = makeErrorRatePanel(titleValue='Database insertion', metricValue='lsif_database_insertion');
local databaseInsertionDurationPercentilesPanel = makeDurationPercentilesPanel(titleValue='Database query', metricValue='lsif_database_insertion');
local xrepoInsertionRequestsPanel = makeRequestsPanel(titleValue='xrepo insertions', metricValue='lsif_xrepo_insertion');
local xrepoInsertionErrorRatePanel = makeErrorRatePanel(titleValue='Xrepo insertion', metricValue='lsif_xrepo_insertion');
local xrepoInsertionDurationPercentilesPanel = makeDurationPercentilesPanel(titleValue='Xrepo insertion', metricValue='lsif_xrepo_insertion');

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
.addRow(title='HTTP requests', panels=[httpRequestsPanel, httpErrorRatePanel, httpDurationPercentilesPanel])
.addRow(title='Database queries', panels=[databaseQueryRequestsPanel, databaseQueryErrorRatePanel, databaseQueryDurationPercentilesPanel])
.addRow(title='Xrepo queries', panels=[xrepoQueryRequestsPanel, xrepoQueryErrorRatePanel, xrepoQueryDurationPercentilesPanel])
.addRow(title='Database insertions', panels=[databaseInsertionRequestsPanel, databaseInsertionErrorRatePanel, databaseInsertionDurationPercentilesPanel])
.addRow(title='Xrepo insertions', panels=[xrepoInsertionRequestsPanel, xrepoInsertionErrorRatePanel, xrepoInsertionDurationPercentilesPanel])
.addRow(title='Caches and Filters', panels=[cacheUtilizationPanel, connectionCacheEventsPanel, documentCacheEventsPanel, resultChunkCacheEventsPanel, bloomFilterEventsPanel])
