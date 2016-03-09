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


var defaults = require('../utilities').defaults;
var isObject = require('../utilities').isObject;
var provide = require('../provide');


/**
 * Adds handler for the history API methods
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function UrlChangeTracker(tracker, opts) {

  // Registers the plugin on the global gaplugins object.
  window.gaplugins = window.gaplugins || {};
  gaplugins.UrlChangeTracker = UrlChangeTracker;

  // Feature detects to prevent errors in unsupporting browsers.
    if (!history.pushState || !window.addEventListener) return;

  this.opts = defaults(opts, {
    shouldTrackUrlChange: this.shouldTrackUrlChange
  });

  this.tracker = tracker;

  // Sets the initial page field.
  // Don't set this on the tracker yet so campaign data can be retreived
  // from the location field.
  this.path = getPath();

  // Overrides history.pushState.
  var originalPushState = history.pushState;
  history.pushState = function(state, title, url) {
    // Sets the document title for reference later.
    // TODO(philipwalton): consider using WeakMap for this to not conflict
    // with any user-defined property also called "title".
    if (isObject(state) && title) state.title = title;

    originalPushState.call(history, state, title, url);
    this.updateTrackerData();
  }.bind(this);

  // Overrides history.repaceState.
  var originalReplaceState = history.replaceState;
  history.replaceState = function(state, title, url) {
    // Sets the document title for reference later.
    // TODO(philipwalton): consider using WeakMap for this to not conflict
    // with any user-defined property also called "title".
    if (isObject(state) && title) state.title = title;

    originalReplaceState.call(history, state, title, url);
    this.updateTrackerData(false);
  }.bind(this);

  // Handles URL changes via user interaction.
  window.addEventListener('popstate', this.updateTrackerData.bind(this));
}


/**
 * Updates the page and title fields on the tracker if necessary and
 * optionally sends a pageview.
 */
UrlChangeTracker.prototype.updateTrackerData = function(shouldSendPageview) {

  // Sets the default.
  shouldSendPageview = shouldSendPageview === false ? false : true;

  // Calls the update logic asychronously to help ensure user callbacks
  // happen first.
  setTimeout(function() {

    var oldPath = this.path;
    var newPath = getPath();

    if (oldPath != newPath &&
        this.opts.shouldTrackUrlChange.call(this, newPath, oldPath)) {

      this.path = newPath;
      this.tracker.set({
        page: newPath,
        title: isObject(history.state) && history.state.title || document.title
      });

      if (shouldSendPageview) this.tracker.send('pageview');
    }
  }.bind(this), 0);
};


/**
 * Determines whether or not the tracker should send a hit with the new page
 * data. This default implementation can be overrided in the config options.
 * @param {string} newPath
 * @param {string} oldPath
 * @return {boolean}
 */
UrlChangeTracker.prototype.shouldTrackUrlChange = function(newPath, oldPath) {
  return true;
};


/**
 * Returns the path value of the current URL.
 * @return {string}
 */
function getPath() {
  return location.pathname + location.search;
}


provide('urlChangeTracker', UrlChangeTracker);
