local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local common = import './common.libsonnet';

// How long to look back for rate() queries
local timeRange = '1m';

// The duration percentiles to display
local percentiles = ['0.95'];

// Colors to pair to percentiles above
local percentileColors = ['#7eb26d'];

// The histogram buckets for sync requests
local buckets = ['1', '2', '5', '10', '30', '60', '120', '+Inf'];

// Colors to pair to buckets defined above (green to red)
local bucketColors = ['#96d98d', '#56a64b', '#37872d', '#e0b400', '#f2cc0c', '#ffee52', '#fa6400', '#c4162a'];

//
// Standard Panels

// Apply defaults defined above to panel constructors
local makeRequestsPanel(title, metric) = common.makePanel(
    title,
    extra = {
        seriesOverrides: common.makeBucketSeriesOverrides(buckets, bucketColors),
    },
    targets = [
        prometheus.target(
            'rate(%s[%s])' % [
                metric,
                timeRange,
            ],
            legendFormat='≤ {{le}}s',
        ),
    ]
);
local makeDurationPercentilesPanel(title, metric) = common.makePanel(
    title,
    extra = {
        yaxes: common.makeYAxes({ format: 's' }),
    },
    targets = std.map(
        function(percentile) prometheus.target(
            'histogram_quantile(%s, rate(%s[%s]))' % [
                percentile,
                metric,
                timeRange,
            ],
            legendFormat = '%sp' % percentile,
        ),
        percentiles
    )
);
local makeErrorRatePanel(title, metric, filter) = common.makeErrorRatePanel(title, metric, timeRange, filter);
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
    title = 'Time gap between least and most up to date user permissions',
    metric = 'src_repoupdater_perms_syncer_perms_gap_seconds{type="user"}'
);
local usersWithStalePermsPanel = common.makePanel(
    title = 'Number of users with stale permissions',
    targets = [
        prometheus.target(
            'src_repoupdater_perms_syncer_stale_perms{type="user"}',
            legendFormat = 'Number of users'
        ),
    ]
);
local usersWithNoPermsPanel = common.makePanel(
    title = 'Number of users with no permissions',
    targets = [
        prometheus.target(
            'src_repoupdater_perms_syncer_no_perms{type="user"}',
            legendFormat = 'Number of users'
        ),
    ]
);

local reposPermsGapPanel = makeDurationSecondsPanel(
    title = 'Time gap between least and most up to date repo permissions',
    metric = 'src_repoupdater_perms_syncer_perms_gap_seconds{type="repo"}'
);
local reposWithStalePermsPanel = common.makePanel(
    title = 'Number of repositories with stale permissions',
    targets = [
        prometheus.target(
            'src_repoupdater_perms_syncer_stale_perms{type="repo"}',
            legendFormat = 'Number of repositories'
        ),
    ]
);
local reposWithNoPermsPanel = common.makePanel(
    title = 'Number of repositories with no permissions',
    targets = [
        prometheus.target(
            'src_repoupdater_perms_syncer_no_perms{type="repo"}',
            legendFormat = 'Number of repositories'
        ),
    ]
);

local usersSyncRequestsPanel = makeRequestsPanel(
    title = 'User permissions synced per minute',
    metric = 'src_repoupdater_perms_syncer_sync_duration_seconds_bucket{type="user"}',
);
local usersSyncRequestsDurationPercentilesPanel = makeDurationPercentilesPanel(
    title = 'User permissions sync duration',
    metric = 'src_repoupdater_perms_syncer_sync_duration_seconds_bucket{type="user"}',
);
local usersSyncRequestsErrorRatePanel = makeErrorRatePanel(
    title = 'User permissions sync',
    metric ='src_repoupdater_perms_syncer_sync',
    filter = 'type="user"'
);

local reposSyncRequestsPanel = makeRequestsPanel(
    title = 'Repo permissions synced per minute',
    metric = 'src_repoupdater_perms_syncer_sync_duration_seconds_bucket{type="repo"}',
);
local reposSyncRequestsDurationPercentilesPanel = makeDurationPercentilesPanel(
    title = 'Repo permissions sync duration',
    metric = 'src_repoupdater_perms_syncer_sync_duration_seconds_bucket{type="repo"}',
);
local reposSyncRequestsErrorRatePanel = makeErrorRatePanel(
    title = 'Repo permissions sync',
    metric ='src_repoupdater_perms_syncer_sync',
    filter = 'type="repo"'
);

local queueSizePanel = common.makePanel(
    title = 'Request queue size',
    targets = [
        prometheus.target(
            'src_repoupdater_perms_syncer_queue_size',
            legendFormat = 'Number of sync requests'
        ),
    ]
);
local authzFilterDurationPercentilesPanel = makeDurationPercentilesPanel(
    title = 'Authorization duration',
    metric = 'src_frontend_authz_filter_duration_seconds_bucket{success="true"}',
);

//
// Dashboard Construction

common.makeDashboard(
    title = 'Permissions'
).addRow(
    title = 'Users permissions sync stats',
    panels = [usersPermsGapPanel, usersWithStalePermsPanel, usersWithNoPermsPanel]
).addRow(
    title = '',
    panels = [usersSyncRequestsPanel, usersSyncRequestsDurationPercentilesPanel, usersSyncRequestsErrorRatePanel]
).addRow(
    title = 'Repositories permissions sync stats',
    panels = [reposPermsGapPanel, reposWithStalePermsPanel, reposWithNoPermsPanel]
).addRow(
    title = '',
    panels = [reposSyncRequestsPanel, reposSyncRequestsDurationPercentilesPanel, reposSyncRequestsErrorRatePanel]
).addRow(
    title = 'Syncer stats',
    panels = [queueSizePanel]
).addRow(
    title = 'General stats',
    panels = [authzFilterDurationPercentilesPanel]
)
