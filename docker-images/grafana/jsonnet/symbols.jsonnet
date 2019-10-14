local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local common = import './common.libsonnet';

local dashboardTemplatingVars = {
    "templating": {
        "list": [
          {
            "allValue": null,
            "current": {
              "text": "All",
              "value": ["$__all"]
            },
            "datasource": "Prometheus",
            "definition": "label_values(symbols_parse_parse_failed, instance)",
            "hide": 0,
            "includeAll": true,
            "label": null,
            "multi": true,
            "name": "instance",
            "options": [],
            "query": "label_values(symbols_parse_parse_failed, instance)",
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

local parseQueueSizePanel = common.makePanel(
  title='Parse Queue Size',
  targets=[
    prometheus.target(
      'symbols_parse_parse_queue_size{instance=~"$instance"}',
      legendFormat='Parse Queue Size',
    ),
  ]
);

local parseFailedRatePanel = common.makePanel(
  title='Parse Failed Rate',
  targets=[
    prometheus.target(
      'rate(symbols_parse_parse_failed{instance=~"$instance"}[$interval])',
      legendFormat='Parse Failed Rate',
    ),
  ]
);

local parseQueueTimeoutsRatePanel = common.makePanel(
  title='Parse Queue Timeouts Rate',
  targets=[
    prometheus.target(
      'rate(symbols_parse_parse_queue_timeouts{instance=~"$instance"}[$interval])',
      legendFormat='Parse Queue Timeouts Rate',
    ),
  ]
);

local parseParsingPanel = common.makePanel(
  title='Parsing',
  targets=[
    prometheus.target(
      'symbols_parse_parsing{instance=~"$instance"}',
      legendFormat='Parsing',
    ),
  ]
);

local storeCacheSizeBytesPanel = common.makePanel(
  title='Store Cache Size',
  extra={
      yaxes: common.makeYAxes({
        format: 'bytes',
      }),
  },
  targets=[
    prometheus.target(
      'symbols_store_cache_size_bytes{instance=~"$instance"}',
      legendFormat='Store Cache Size',
    ),
  ]
);

local storeEvictionRatePanel = common.makePanel(
  title='Store Eviction Rate',
  targets=[
    prometheus.target(
      'rate(symbols_store_evictions{instance=~"$instance"}[$interval])',
      legendFormat='Store Eviction Rate',
    ),
  ]
);

local storeFetchFailedRatePanel = common.makePanel(
  title='Store Fetch Failed Rate',
  targets=[
    prometheus.target(
      'rate(symbols_store_fetch_failed{instance=~"$instance"}[$interval])',
      legendFormat='Store Fetch Failed Rate',
    ),
  ]
);

local storeFetchQueueSizePanel = common.makePanel(
  title='Store Fetch Queue Size',
  targets=[
    prometheus.target(
      'symbols_store_fetch_queue_size{instance=~"$instance"}',
      legendFormat='Store Fetch Queue Size',
    ),
  ]
);

local storeFetchingPanel = common.makePanel(
  title='Store Fetching',
  targets=[
    prometheus.target(
      'symbols_store_fetching{instance=~"$instance"}',
      legendFormat='Store Fetching',
    ),
  ]
);

//
// Dashboard Construction

common.makeDashboard(title='Symbols (ctags)', extra=dashboardTemplatingVars)
.addRow(title='Parsing', panels=[parseQueueSizePanel, parseFailedRatePanel, parseQueueTimeoutsRatePanel])
.addRow(title='Parsing', panels=[parseParsingPanel])
.addRow(title='Store', panels=[storeCacheSizeBytesPanel, storeEvictionRatePanel, storeFetchFailedRatePanel])
.addRow(title='Store', panels=[storeFetchQueueSizePanel, storeFetchingPanel])
