local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local common = import './common.libsonnet';

local bucketColors = ['#96d98d', '#56a64b', '#37872d', '#e0b400', '#f2cc0c', '#ffee52', '#fa6400', '#c4162a'];

local indexRepoSecondsBuckets = ['0.1', '1.0', '10', '100', '1000', '10000', '100000', '+Inf'];
local resolveRevisionSecondsBuckets = ['0.25', '0.5', '1', '2', '+Inf'];
local resolveAllRevisionsSecondsBuckets = ['1', '10', '100', '1000', '10000', '100000', '+Inf'];

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
            "definition": "label_values(index_repo_seconds_bucket, instance)",
            "hide": 0,
            "includeAll": true,
            "label": null,
            "multi": true,
            "name": "instance",
            "options": [],
            "query": "label_values(index_repo_seconds_bucket, instance)",
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

local repoIndexingSecondsPanel = common.makePanel(
       title='Latencies for indexing a repo (seconds)',
       extra={
         seriesOverrides: common.makeBucketSeriesOverrides(indexRepoSecondsBuckets, bucketColors),
       },
       targets=[
         prometheus.target(
           'rate(index_repo_seconds_bucket{state="success", instance=~"$instance"}[$interval])',
           legendFormat='≤ {{le}}s',
         ),
       ]
     );

local repoRevisionSecondsPanel = common.makePanel(
       title='Latencies for resolving a repo revision (seconds)',
       extra={
         seriesOverrides: common.makeBucketSeriesOverrides(resolveRevisionSecondsBuckets, bucketColors),
       },
       targets=[
         prometheus.target(
           'rate(resolve_revision_seconds_bucket{success="true", instance=~"$instance"}[$interval])',
           legendFormat='≤ {{le}}s',
         ),
       ]
     );

local repoAllRevisionsSecondsPanel = common.makePanel(
       title='Latencies for resolving all revisions (seconds)',
       extra={
         seriesOverrides: common.makeBucketSeriesOverrides(resolveAllRevisionsSecondsBuckets, bucketColors),
       },
       targets=[
         prometheus.target(
           'rate(resolve_revisions_seconds_bucket{instance=~"$instance"}[$interval])',
           legendFormat='≤ {{le}}s',
         ),
       ]
     );

common.makeDashboard(title='Zoekt indexserver', extra=dashboardTemplatingVars)
.addRow(title='indexing', panels=[repoIndexingSecondsPanel])
.addRow(title='resolving', panels=[repoRevisionSecondsPanel, repoAllRevisionsSecondsPanel])
