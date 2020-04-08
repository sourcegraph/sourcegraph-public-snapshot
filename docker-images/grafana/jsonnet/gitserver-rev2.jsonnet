local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local common = import './common.libsonnet';

local bucketColors = ['#96d98d', '#56a64b', '#37872d', '#e0b400', '#f2cc0c', '#ffee52', '#fa6400', '#c4162a'];

local execDurationSecondsBuckets = ['0.2', '0.5', '1', '2', '5', '10', '30', '+Inf'];

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
            "definition": "label_values(src_gitserver_disk_space_available, instance)",
            "hide": 0,
            "includeAll": true,
            "label": null,
            "multi": true,
            "name": "instance",
            "options": [],
            "query": "label_values(src_gitserver_disk_space_available, instance)",
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

local deadlineExceededRatePanel = common.makePanel(
    title='Deadline Exceeded Rate',
    targets=[
      prometheus.target(
        "rate(src_gitserver_client_deadline_exceeded[$interval])",
        legendFormat='Deadline Exceeded Rate',
      ),
    ]
);


local cloneQueueSizePanel = common.makePanel(
  title='Clone Queue Size',
  targets=[
    prometheus.target(
      'src_gitserver_clone_queue{instance=~"$instance"}',
      legendFormat='Clone Queue Size',
    ),
  ]
);

local diskSpaceAvailableBytesPanel = common.makePanel(
  title='Available Disk Space',
  extra={
      yaxes: common.makeYAxes({
        format: 'bytes',
      }),
  },
  targets=[
    prometheus.target(
      'src_gitserver_disk_space_available{instance=~"$instance"}',
      legendFormat='Disk space available',
    ),
  ]
);

local echoDurationSecondsPanel = common.makePanel(
  title='Echo Duration Seconds',
  extra={
      yaxes: common.makeYAxes({
        format: 's',
      }),
  },
  targets=[
    prometheus.target(
      'src_gitserver_echo_duration_seconds{instance=~"$instance"}',
      legendFormat='Echo duration seconds',
    ),
  ]
);

local commandLatenciesSecondsPanel = common.makePanel(
       title='Latencies for commands (seconds)',
       extra={
         seriesOverrides: common.makeBucketSeriesOverrides(execDurationSecondsBuckets, bucketColors),
       },
       targets=[
         prometheus.target(
           'rate(src_gitserver_exec_duration_seconds_bucket{instance=~"$instance"}[$interval])',
           legendFormat='≤ {{le}}s',
         ),
       ]
     );

local execRunningPanel = common.makePanel(
  title='Commands running concurrently',
  targets=[
    prometheus.target(
      'src_gitserver_exec_running{instance=~"$instance"}',
      legendFormat='Commands running concurrently',
    ),
  ]
);

local lsRemotePanel = common.makePanel(
  title='Num repos waiting on git ls-remote',
  targets=[
    prometheus.target(
      'src_gitserver_lsremote_queue{instance=~"$instance"}',
      legendFormat='Repos waiting on git ls-remote',
    ),
  ]
);

local reposClonedRatePanel = common.makePanel(
  title='Repos Cloned Rate',
  targets=[
    prometheus.target(
      'rate(src_gitserver_repo_cloned{instance=~"$instance"}[$interval])',
      legendFormat='Repos cloned rate',
    ),
  ]
);

local reposReclonedRatePanel = common.makePanel(
  title='Repos Recloned Rate',
  targets=[
    prometheus.target(
      'rate(src_gitserver_repos_recloned{instance=~"$instance"}[$interval])',
      legendFormat='Repos recloned rate',
    ),
  ]
);

local reposRemovedRatePanel = common.makePanel(
  title='Repos Removed Rate',
  targets=[
    prometheus.target(
      'rate(src_gitserver_repos_removed{instance=~"$instance"}[$interval])',
      legendFormat='Repos removed rate',
    ),
  ]
);

common.makeDashboard(title='Gitserver rev2', extra=dashboardTemplatingVars)
.addRow(title='', panels=[deadlineExceededRatePanel, cloneQueueSizePanel, diskSpaceAvailableBytesPanel])
.addRow(title='', panels=[echoDurationSecondsPanel, commandLatenciesSecondsPanel, execRunningPanel])
.addRow(title='', panels=[lsRemotePanel, reposClonedRatePanel, reposReclonedRatePanel, reposRemovedRatePanel])
