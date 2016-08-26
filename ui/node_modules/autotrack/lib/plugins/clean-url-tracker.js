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


var assign = require('object-assign');
var parseUrl = require('dom-utils/lib/parse-url');
var constants = require('../constants');
var provide = require('../provide');
var usage = require('../usage');


/**
 * Registers clean URL tracking on a tracker object. The clean URL tracker
 * removes query parameters from the page value reported to Google Analytics.
 * It also helps to prevent tracking similar URLs, e.g. sometimes ending a URL
 * with a slash and sometimes not.
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function CleanUrlTracker(tracker, opts) {

  usage.track(tracker, usage.plugins.CLEAN_URL_TRACKER);

  this.opts = assign({
    stripQuery: false,
    queryDimensionIndex: null,
    indexFilename: null,
    trailingSlash: null
  }, opts);

  this.tracker = tracker;

  this.overrideTrackerBuildHitTask();
}


/**
 * Cleans the URL based on the preferences set in the configuration options.
 * @param {Object} model An analytics.js Model object.
 */
CleanUrlTracker.prototype.cleanUrlTask = function(model) {

  var location = model.get('location');
  var page = model.get('page');
  var url = parseUrl(page || location);

  var oldPath = url.pathname;
  var newPath = oldPath;

  // If an index filename was provided, remove it if it appears at the end
  // of the URL.
  if (this.opts.indexFilename) {
    var parts = newPath.split('/');
    if (this.opts.indexFilename == parts[parts.length - 1]) {
      parts[parts.length - 1] = '';
      newPath = parts.join('/');
    }
  }

  // Ensure the URL ends with or doesn't end with slash based on the
  // `trailingSlash` option. Note that filename URLs should never contain
  // a trailing slash.
  if (this.opts.trailingSlash == 'remove') {
      newPath = newPath.replace(/\/+$/, '');
  }
  else if (this.opts.trailingSlash == 'add') {
    var isFilename = /\.\w+$/.test(newPath);
    if (!isFilename && newPath.substr(-1) != '/') {
      newPath = newPath + '/';
    }
  }

  // If a query dimensions index was provided, set the query string portion
  // of the URL on that dimension. If no query string exists on the URL use
  // the NULL_DIMENSION.
  if (this.opts.stripQuery && this.opts.queryDimensionIndex) {
    model.set('dimension' + this.opts.queryDimensionIndex,
        url.query || constants.NULL_DIMENSION);
  }

  model.set('page', newPath + (!this.opts.stripQuery ? url.search : ''));
};


/**
 * Overrides the tracker's `buildHitTask` to check for proper URL formatting
 * on every hit (not just the initial pageview).
 */
CleanUrlTracker.prototype.overrideTrackerBuildHitTask = function() {
  this.originalTrackerBuildHitTask = this.tracker.get('buildHitTask');

  this.tracker.set('buildHitTask', function(model) {
    this.cleanUrlTask(model);
    this.originalTrackerBuildHitTask(model);
  }.bind(this));
};


/**
 * Restores all overridden tasks and methods.
 */
CleanUrlTracker.prototype.remove = function() {
  this.tracker.set('sendHitTask', this.originalTrackerSendHitTask);
};


provide('cleanUrlTracker', CleanUrlTracker);
