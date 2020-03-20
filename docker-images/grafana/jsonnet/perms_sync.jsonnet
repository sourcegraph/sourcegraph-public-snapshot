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

// The histogram buckets for sync requests
local buckets = ['1', '2', '5', '10', '30', '60', '120', '+Inf'];

// Colors to pair to buckets defined above (green to red)
local bucketColors = ['#96d98d', '#56a64b', '#37872d', '#e0b400', '#f2cc0c', '#ffee52', '#fa6400', '#c4162a'];

//
// Standard Panels

// Apply defaults defined above to panel constructors
local makeRequestsPanel(title, metric) = common.makeRequestsPanel(title, metric, timeRange=timeRange, buckets=buckets, colors=bucketColors);
local makeDurationPercentilesPanel(title, metric) = common.makeDurationPercentilesPanel(title, metric, timeRange=timeRange, percentiles=percentiles, colors=percentileColors);
local makeErrorRatePanel(title, metric) = common.makeErrorRatePanel(title, metric, timeRange=timeRange);
local makeDurationSecondsPanel(title, metric) = common.makePanel(
  title = title,
  extra = {
      yaxes: common.makeYAxes({
        format: 's',
      }),
  },
  targets = [
    prometheus.target(
      metric,
      legendFormat = 'Seconds',
    ),
  ]
);

local usersPermsGapPanel = makeDurationSecondsPanel(
    title = 'The largest time gap between users',
    metric = 'src_repoupdater_perms_syncer_users_perms_gap_seconds'
);
local usersWithStalePermsPanel = common.makePanel(
    title = 'Number of users with stale permissions',
    targets = [
        prometheus.target(
            "src_repoupdater_perms_syncer_users_with_stale_perms",
            legendFormat = 'Number of users'
        ),
    ]
);
local usersWithNoPermsPanel = common.makePanel(
    title = 'Number of users with no permissions',
    targets = [
        prometheus.target(
            "src_repoupdater_perms_syncer_users_with_no_perms",
            legendFormat = 'Number of users'
        ),
    ]
);

local reposPermsGapPanel = makeDurationSecondsPanel(
    title = 'The largest time gap between repos',
    metric = 'src_repoupdater_perms_syncer_repos_perms_gap_seconds'
);
local reposWithStalePermsPanel = common.makePanel(
    title = 'Number of repositories with stale permissions',
    targets = [
        prometheus.target(
            "src_repoupdater_perms_syncer_repos_with_stale_perms",
            legendFormat = 'Number of repositories'
        ),
    ]
);
local reposWithNoPermsPanel = common.makePanel(
    title = 'Number of repositories with no permissions',
    targets = [
        prometheus.target(
            "src_repoupdater_perms_syncer_repos_with_no_perms",
            legendFormat = 'Number of repositories'
        ),
    ]
);

local usersSyncRequestsPanel = makeRequestsPanel(
    title = 'sync',
    metric = 'src_repoupdater_perms_syncer_users_sync'
);
local usersSyncRequestsDurationPercentilesPanel = makeDurationPercentilesPanel(
    title = 'sync',
    metric = 'src_repoupdater_perms_syncer_users_sync'
);
local usersSyncRequestsErrorRatePanel = makeErrorRatePanel(
    title = 'sync',
    metric ='src_repoupdater_perms_syncer_users_sync_errors_total'
);

local reposSyncRequestsPanel = makeRequestsPanel(
    title = 'sync',
    metric = 'src_repoupdater_perms_syncer_repos_sync'
);
local reposSyncRequestsDurationPercentilesPanel = makeDurationPercentilesPanel(
    title = 'sync',
    metric = 'src_repoupdater_perms_syncer_repos_sync'
);
local reposSyncRequestsErrorRatePanel = makeErrorRatePanel(
    title = 'sync',
    metric ='src_repoupdater_perms_syncer_repos_sync_errors_total'
);

//
// Dashboard Construction

common.makeDashboard(
    title = 'Permissions Sync'
).addRow(
    title = 'Users permissions stats',
    panels = [usersPermsGapPanel, usersWithStalePermsPanel, usersWithNoPermsPanel]
).addRow(
    title = 'Repositories permissions stats',
    panels = [reposPermsGapPanel, reposWithStalePermsPanel, reposWithNoPermsPanel]
).addRow(
    title = 'Users permissions sync',
    panels = [usersSyncRequestsPanel, usersSyncRequestsDurationPercentilesPanel, usersSyncRequestsErrorRatePanel]
).addRow(
    title = 'Repositories permissions sync',
    panels = [reposSyncRequestsPanel, reposSyncRequestsDurationPercentilesPanel, reposSyncRequestsErrorRatePanel]
)
