local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;

// The standard options that are applied to a new Grafana dashboard
// This sets the schema version, shared crosshairs, and a default time
// frame and refresh rate. These values can be overridden by supplying
// extra values to makeDashboard.
local standardDashboardOptions = {
  schemaVersion: 18,
  tags: [],
  editable: true,
  graphTooltip: 1,  // shared crosshair
  refresh: '30s',
  time: {
    from: 'now-3h',
    to: 'now',
  },
  timezone: 'utc',
};

// The standard options that are applied to the yaxes of a new panel.
// This mainly exists to set a lower bound of zero on panels that have
// little or no data (an appear as a weird line through the center).
// These values can be overridden by supplying `yaxes` in the extra
// values to makePanel (use the return value from makeYaxes for this).
local standardYAxisOptions = {
  format: 'short',
  label: null,
  logBase: 1,
  max: null,
  min: 0,
  show: true,
};

// The standard options that are applied to a new panel. This turns off
// the legend, sets transparency, a sane default yaxes and sort order for
// tooltips. These values can be overridden by supplying extra values to
// makePanel.
local standardPanelOptions = {
  min: 0,
  legend_show: false,
  legend: {
    show: false,
  },
  format: 'short',
  transparent: true,
  yaxes: [
    standardYAxisOptions,
    standardYAxisOptions,
  ],
  tooltip: {
    sort: 2,  // descending
  },
};

// Create a value suitable for the yaxes value of a panel. These override
// the default yaxes options. For example, to update the yaxes of a panel
// for percentage data, do the following:
//
//    common.makePanel(..., extra={
//      yaxes: common.makeYAxes({ max: 1, format: 'percentunit' }),
//    }, ...)
local makeYAxes(extra={}) = [
  standardYAxisOptions + extra,
  standardYAxisOptions,
];

local Dashboard = {
  n:: 0,  // track panel ids
  y:: 0,  // track vertical offset for new rows

  // Add a titled and collapsible row to the dashboard. The row is 24x6,
  // and will stretch or shrink graphs to make the provided panels fit.
  // Each panel will have an equal width and constant height.
  addRow(title, panels, collapsed=false)::
    local n = self.n;
    local y = self.y;

    // Get width of each panel
    local panelWidth = 24 / std.length(panels);

    // Update values of n and y to reflect the new height and the
    // new number of panels in the graph.
    local updatedVars = {
      n: n + std.length(panels) + 1,
      y: y + 7,
    };

    // Set the grid position and a unique ID for each panel
    local mappedPanels = std.mapWithIndex(
      function(i, panel) panel {
        id: n + 1 + i,
        gridPos: { x: i * panelWidth, y: y + 1, w: panelWidth, h: 6 },
      },
      panels,
    );

    local rowPanel = {
      id: n,
      type: 'row',
      title: title,
      collapsed: false,
      panels: [],
    };

    if collapsed then
      // If we're collapsing, then the panels belong to the row
      self.addPanel(
        rowPanel { collapsed: true, panels: mappedPanels },
        gridPos={ x: 0, y: y, w: 24, h: 1 }
      ) + updatedVars
    else
      // If we're not collapsing, the panels under the row should be
      // empty, and each un-collapsed panel should be a sibling of the
      // row in the top level.

      std.foldl(
        function(dashboard, panel) dashboard.addPanel(panel, gridPos=panel.gridPos),
        mappedPanels,
        self.addPanel(rowPanel, gridPos={ x: 0, y: y, w: 24, h: 1 })
      ) + updatedVars,
};

// Create a new dashboard with the given title. This also adds the function
// addRow, defined in the type above. Extras will override default values.
local makeDashboard(title, extra={}) = dashboard.new(title) + standardDashboardOptions + Dashboard + extra;

// Create a new panel with the given title and targets. Extras will override
// default values.
local makePanel(title, targets, extra={}) = (graphPanel.new(title=title) + standardPanelOptions + extra).addTargets(targets);

// Make a seriesOverrides array suitable for use with percentile graphs.
// Percentiles should be an array of increasing percentile values, e.g.
// ['0.5', 0.9', '0.99']. Colors should be an array at least as large as
// the percentiles array. Colors are zipped to percentiles in order.
//
// This creates legends such as 0.5p, and higher percentiles will fill
// down to the next percentile. Colors should generally be increasing in
// alarm (green or blue to red) so that the higher percentiles stand out.
local makePercentileSeriesOverrides(percentiles, colors) = std.mapWithIndex(
  function(i, percentile) {
    alias: '%sp' % percentile,
    color: colors[i],
    fillBelowTo: if i == 0 then '' else '%sp' % percentiles[i - 1],
    fill: if i == 0 then 3 else 1,
    zindex: std.length(percentiles) - i,
  },
  percentiles
);

// Make a seriesOverrides array suitable fo ruse with a duration histogram.
// Buckets should be an array of increasing durations,
// e.g. ['0.2', '0.5', '+Inf']. Colors should be an array at least as large
// as the buckets array. Colors are zipped to buckets in order.
//
// This creates legends such as ≤ 0.5s, and higher buckets will fill down
// to the next bucket. Colors should generally be increasing in alarm (green
// or blue to red) so that the "warmer" a graph is, the more frequent the
// requests are to higher buckets. z-indices are chosen so that the lower
// buckets will appear in front of the higher ones (thus an equal number of
// requests in the 0.5s and the 1s buckets will not appear as if all the
// requests belong to the larger bucket).
local makeBucketSeriesOverrides(buckets, colors) = std.mapWithIndex(
  function(i, bucket) {
    alias: '≤ %ss' % bucket,
    color: colors[i],
    fillBelowTo: if i == 0 then '' else '≤ %ss' % buckets[i - 1],
    zindex: std.length(buckets) - i,
  },
  buckets
);

// Make a panel to display the total number of operations performed and a
// histogram of their duration. This assumes a histogram metric is emitted
// by Prometheus.
//
// - titleValue: fill the placeholder in 'Number of %s by duration'
// - metricValue: fills the placeholder in '%s_duration_seconds_bucket'
// - metricFilter: if supplied, applies a filter expression to the metric name
// - buckets: the buckets that match the histogram 'le' labels
local makeRequestsPanel(titleValue, metricValue, timeRange, buckets, colors, metricFilter='') = makePanel(
  title='Number of %s by duration' % titleValue,
  extra={
    seriesOverrides: makeBucketSeriesOverrides(buckets, colors),
  },
  targets=[
    prometheus.target(
      'rate(%s_duration_seconds_bucket%s[%s])' % [
        metricValue,
        if metricFilter == '' then '' else '{%s}' % metricFilter,
        timeRange,
      ],
      legendFormat='≤ {{le}}s',
    ),
  ]
);

// Make a panel to display the total number of HTTP requests performed regardless
// of the status code. This assumes a histogram metric is emitted by Prometheus.
// See makeHttpRequestsPanel for the histogram version.
local makeHttpTotalRequestsPanel(titleValue, metricValue, timeRange, metricFilter='') = makePanel(
  title='Total number of %s' % titleValue,
  targets=[
    prometheus.target(
      "rate(%s_duration_seconds_bucket{le='+Inf'%s}[%s])" % [
        metricValue,
        if metricFilter == '' then '' else ',%s' % metricFilter,
        timeRange,
      ],
      legendFormat='Total requests',
    ),
  ]
);

// Make a panel to display the total number of HTTP requests performed and a
// histogram of their duration. This assumes a histogram metric is emitted
// by Prometheus with a code label. See makeRequestsPanel for reference.
// This will only display the requests that return a 200-level response.
local makeHttpRequestsPanel(titleValue, metricValue, timeRange, buckets, colors, metricFilter='') = makeRequestsPanel(
  titleValue=titleValue,
  metricValue=metricValue,
  metricFilter='code=~"2.."' + (if metricFilter == '' then '' else ',%s' % metricFilter),
  timeRange=timeRange,
  buckets=buckets,
  colors=colors,
);

// Make a panel to display the percentage of operations that resulted in an
// error. This assumes a histogram metric as well as a separate error counter
// metric (of the form %s_errors_total) is emitted by Prometheus.
//
// - titleValue: fill the placeholder in '%s error rate'
// - metricValue: fills the placeholder in '%s_duration_seconds_bucket'
//                and '%s_errors_total'
local makeErrorRatePanel(titleValue, metricValue, timeRange, metricFilter='') = makePanel(
  title='%s error rate' % titleValue,
  extra={
    lines: false,
    bars: true,
    stack: true,
    max: 1,
    yaxes: makeYAxes({ max: 1, format: 'percentunit' }),
    seriesOverrides: [{
      alias: '% of failing operations',
      color: '#c4162a',
    }],
    tooltip: {
      common: false,
    },
  },
  targets=[
    prometheus.target(
      'rate(%s_errors_total%s[%s]) / rate(%s_duration_seconds_count%s[%s])' % [
        metricValue,
        if metricFilter == '' then '' else '{%s}' % metricFilter,
        timeRange,
        metricValue,
        if metricFilter == '' then '' else '{%s}' % metricFilter,
        timeRange,
      ],
      legendFormat='% of failing operations',
    ),
  ]
);

// Make a panel to display the percentage of HTTP requests that resulted in an
// error of a particular type. This assumes a histogram metric is emitted by
// Prometheus with a code label. See makeErrorRatePanel for reference.
//
// - patterns: A list of status code regular expressions that will partition
//             the error types (e.g. ['4..', '5..']).
// - colors: An array of colors that create series overrides for the given
//           error pattern types.
local makeHttpErrorRatePanel(titleValue, metricValue, timeRange, patterns, colors, metricFilter='') = makePanel(
  title='%s error rate' % titleValue,
  extra={
    lines: false,
    bars: true,
    stack: true,
    max: 1,
    yaxes: makeYAxes({ max: 1, format: 'percentunit' }),
    seriesOverrides: std.mapWithIndex(
      function(i, statusCodePattern) {
        alias: '%% of %s responses' % statusCodePattern,
        color: colors[i],
      },
      patterns
    ),
    tooltip: {
      common: false,
    },
  },
  targets=std.map(
    function(statusCodePattern) prometheus.target(
      'sum(rate(%s_duration_seconds_count{code=~"%s"%s}[%s])) / sum(rate(%s_duration_seconds_count%s[%s]))' % [
        metricValue,
        statusCodePattern,
        if metricFilter == '' then '' else ',%s' % metricFilter,
        timeRange,
        metricValue,
        if metricFilter == '' then '' else '{%s}' % metricFilter,
        timeRange,
      ],
      legendFormat='%% of %s responses' % statusCodePattern,
    ),
    patterns
  ),
);

// Make a panel to display duration percentiles. This assumes a histogram metric
// is emitted by Prometheus.
//
// - titleValue: fill the placeholder in '%s duration percentiles'
// - metricValue: fills the placeholder in '%s_duration_seconds_bucket'
// - percentiles: The list of percentiles to graph
local makeDurationPercentilesPanel(titleValue, metricValue, timeRange, percentiles, colors, metricFilter='') =
  makePanel(
    title='%s duration percentiles' % titleValue,
    extra={
      yaxes: makeYAxes({ format: 's' }),
      seriesOverrides: makePercentileSeriesOverrides(percentiles, colors),
    },
    targets=std.map(
      function(percentile) prometheus.target(
        'histogram_quantile(%s, rate(%s_duration_seconds_bucket%s[%s]))' % [
          percentile,
          metricValue,
          if metricFilter == '' then '' else '{%s}' % metricFilter,
          timeRange,
        ],
        legendFormat='%sp' % percentile,
      ),
      percentiles
    )
  );

// Make a panel to display duration percentiles for HTTP requests. This assumes
// a histogram metric is emitted by Prometheus with a code label. See
// makeDurationPercentilesPanel for reference. This will only display the requests
// that return a 200-level response.
local makeHttpDurationPercentilesPanel(titleValue, metricValue, timeRange, percentiles, colors, metricFilter='') = makeDurationPercentilesPanel(
  titleValue=titleValue,
  metricValue=metricValue,
  metricFilter='code=~"2.."' + (if metricFilter == '' then '' else ',%s' % metricFilter),
  timeRange=timeRange,
  percentiles=percentiles,
  colors=colors,
);

//
// Exports

{
  makeYAxes:: makeYAxes,
  makeDashboard:: makeDashboard,
  makePanel:: makePanel,
  makeRequestsPanel:: makeRequestsPanel,
  makeHttpTotalRequestsPanel:: makeHttpTotalRequestsPanel,
  makeHttpRequestsPanel:: makeHttpRequestsPanel,
  makeErrorRatePanel:: makeErrorRatePanel,
  makeHttpErrorRatePanel:: makeHttpErrorRatePanel,
  makeDurationPercentilesPanel:: makeDurationPercentilesPanel,
  makeHttpDurationPercentilesPanel:: makeHttpDurationPercentilesPanel,
  makeBucketSeriesOverrides:: makeBucketSeriesOverrides,
}
