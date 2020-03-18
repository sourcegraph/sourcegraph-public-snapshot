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
local makeRequestsPanel(titleValue, metricValue) = common.makeRequestsPanel(titleValue, metricValue, timeRange=timeRange, buckets=buckets, colors=bucketColors);
local makeDurationPercentilesPanel(titleValue, metricValue) = common.makeDurationPercentilesPanel(titleValue, metricValue, timeRange=timeRange, percentiles=percentiles, colors=percentileColors);
local makeErrorRatePanel(titleValue, metricValue) = common.makeErrorRatePanel(titleValue, metricValue, timeRange=timeRange);

local requestsDurationsPanel = makeRequestsPanel(
    titleValue = 'sync requests',
    metricValue = 'src_repoupdater_authz_perms_sync'
);
local requestsDurationPercentilesPanel = makeDurationPercentilesPanel(
    titleValue = 'sync requests',
    metricValue = 'src_repoupdater_authz_perms_sync'
);
local requestsErrorRatePanel = makeErrorRatePanel(
    titleValue = 'sync requests',
    metricValue ='src_repoupdater_authz_perms_sync_errors_total'
);

local noPermsUsersPanel = common.makePanel(
    title='Number of users with no permissions',
    targets=[
        prometheus.target(
            "src_repoupdater_authz_no_perms_users",
            legendFormat='Number of users'
        ),
    ]
);
local noPermsReposPanel = common.makePanel(
    title='Number of repositories with no permissions',
    targets=[
        prometheus.target(
            "src_repoupdater_authz_no_perms_repos",
            legendFormat='Number of repositories'
        ),
    ]
);

//
// Dashboard Construction

common.makeDashboard(
    title = 'Permissions Sync'
).addRow(
    title = 'Sync requests',
    panels = [requestsDurationsPanel, requestsDurationPercentilesPanel, requestsErrorRatePanel]
).addRow(
    title = 'Stats',
    panels = [noPermsUsersPanel, noPermsReposPanel]
)
