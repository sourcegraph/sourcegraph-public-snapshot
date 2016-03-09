// webpack entry point

require("babel-polyfill");

window.jQuery = window.$ = require("jquery");

require("bootstrap-sass/assets/javascripts/bootstrap/transition.js");
require("bootstrap-sass/assets/javascripts/bootstrap/collapse.js");
require("bootstrap-sass/assets/javascripts/bootstrap/button.js");
require("bootstrap-sass/assets/javascripts/bootstrap/dropdown.js");
require("bootstrap-sass/assets/javascripts/bootstrap/modal.js");

require("sourcegraph/init/AppdashRecordPerfTiming");

// Views
require("./componentInjection");

require("../style/web.scss");

require("sourcegraph/util/actionLogger");

// REQUIRED. Configures Sentry error monitoring.
require("sourcegraph/init/Sentry");

// REQUIRED. Enables HTML history API (pushState) tracking in Google Analytics.
// See https://github.com/googleanalytics/autotrack#shouldtrackurlchange.
require("autotrack/lib/plugins/url-change-tracker");
