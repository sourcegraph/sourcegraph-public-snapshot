/**
 * Copyright 2016 Google Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Imports sub-plugins.
require('./event-tracker');
require('./media-query-tracker');
require('./outbound-form-tracker');
require('./outbound-link-tracker');
require('./social-tracker');
require('./url-change-tracker');


// Imports dependencies.
var provide = require('../provide');


/**
 *
 * Requires all sub-plugins via a single plugin.
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function Autotrack(tracker, opts) {
  var ga = window[window.GoogleAnalyticsObject || 'ga'];
  var name = tracker.get('name');

  // Registers the plugin on the global gaplugins object.
  window.gaplugins = window.gaplugins || {};
  gaplugins.Autotrack = Autotrack;

  ga(name + '.require', 'eventTracker', opts);
  ga(name + '.require', 'mediaQueryTracker', opts);
  ga(name + '.require', 'outboundFormTracker', opts);
  ga(name + '.require', 'outboundLinkTracker', opts);
  ga(name + '.require', 'socialTracker', opts);
  ga(name + '.require', 'urlChangeTracker', opts);
}


provide('autotrack', Autotrack);
