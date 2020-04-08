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

local githubRequestsPanel = makeHttpRequestsPanel(titleValue='Github requests', metricValue='src_github_request');
local githubErrorRatePanel = makeHttpErrorRatePanel(titleValue='Github', metricValue='src_github_request');
local githubDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='Github request', metricValue='src_github_request');

local bitbucketRequestsPanel = makeHttpRequestsPanel(titleValue='Bitbucket requests', metricValue='src_bitbucket_request');
local bitbucketErrorRatePanel = makeHttpErrorRatePanel(titleValue='Bitbucket', metricValue='src_bitbucket_request');
local bitbucketDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='Bitbucket request', metricValue='src_bitbucket_request');

local gitlabRequestsPanel = makeHttpRequestsPanel(titleValue='Gitlab requests', metricValue='src_gitlab_request');
local gitlabErrorRatePanel = makeHttpErrorRatePanel(titleValue='Gitlab', metricValue='src_gitlab_request');
local gitlabDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='Gitlab request', metricValue='src_gitlab_request');

local bitbucketCloudRequestsPanel = makeHttpRequestsPanel(titleValue='bitbucket cloud requests', metricValue='src_bitbucket_cloud_request');
local bitbucketCloudErrorRatePanel = makeHttpErrorRatePanel(titleValue='bitbucket cloud', metricValue='src_bitbucket_cloud_request');
local bitbucketCloudDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='bitbucket cloud request', metricValue='src_bitbucket_cloud_request');

local phabricatorRequestsPanel = makeHttpRequestsPanel(titleValue='phabricator requests', metricValue='src_phabricator_request');
local phabricatorErrorRatePanel = makeHttpErrorRatePanel(titleValue='phabricator', metricValue='src_phabricator_request');
local phabricatorDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='phabricator request', metricValue='src_phabricator_request');

local handlerRequestsPanel = makeHttpRequestsPanel(titleValue='Http handler calls', metricValue='src_repoupdater_http_handler');
local handlerErrorRatePanel = makeHttpErrorRatePanel(titleValue='Http handler', metricValue='src_repoupdater_http_handler');
local handlerDurationPercentilesPanel = makeHttpDurationPercentilesPanel(titleValue='Http handler call', metricValue='src_repoupdater_http_handler');


//
// Dashboard Construction

common.makeDashboard(title='Repoupdater to external services')
.addRow(title='Http handler calls', panels=[handlerRequestsPanel, handlerErrorRatePanel, handlerDurationPercentilesPanel])
.addRow(title='Github requests', panels=[githubRequestsPanel, githubErrorRatePanel, githubDurationPercentilesPanel])
.addRow(title='Gitlab requests', panels=[gitlabRequestsPanel, gitlabErrorRatePanel, gitlabDurationPercentilesPanel])
.addRow(title='Bitbucket requests', panels=[bitbucketRequestsPanel, bitbucketErrorRatePanel, bitbucketDurationPercentilesPanel])
.addRow(title='Bitbucket Cloud requests', panels=[bitbucketCloudRequestsPanel, bitbucketCloudErrorRatePanel, bitbucketCloudDurationPercentilesPanel])
.addRow(title='Phabricator requests', panels=[phabricatorRequestsPanel, phabricatorErrorRatePanel, phabricatorDurationPercentilesPanel])
