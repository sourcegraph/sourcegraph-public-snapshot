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
var provide = require('../provide');
var usage = require('../usage');
var createFieldsObj = require('../utilities').createFieldsObj;
var isObject = require('../utilities').isObject;


/**
 * Adds handler for the history API methods
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function UrlChangeTracker(tracker, opts) {

  usage.track(tracker, usage.plugins.URL_CHANGE_TRACKER);

  // Feature detects to prevent errors in unsupporting browsers.
  if (!history.pushState || !window.addEventListener) return;

  this.opts = assign({
    shouldTrackUrlChange: this.shouldTrackUrlChange,
    fieldsObj: {},
    hitFilter: null
  }, opts);

  this.tracker = tracker;

  // Sets the initial page field.
  // Don't set this on the tracker yet so campaign data can be retreived
  // from the location field.
  this.path = getPath();

  this.updateTrackerData = this.updateTrackerData.bind(this);

  // Overrides history.pushState.
  this.originalPushState = history.pushState;
  history.pushState = function(state, title) {
    // Sets the document title for reference later.
    // TODO(philipwalton): consider using WeakMap for this to not conflict
    // with any user-defined property also called "title".
    if (isObject(state) && title) state.title = title;

    this.originalPushState.apply(history, arguments);
    this.updateTrackerData();
  }.bind(this);

  // Overrides history.repaceState.
  this.originalReplaceState = history.replaceState;
  history.replaceState = function(state, title) {
    // Sets the document title for reference later.
    // TODO(philipwalton): consider using WeakMap for this to not conflict
    // with any user-defined property also called "title".
    if (isObject(state) && title) state.title = title;

    this.originalReplaceState.apply(history, arguments);
    this.updateTrackerData(false);
  }.bind(this);

  // Handles URL changes via user interaction.
  window.addEventListener('popstate', this.updateTrackerData);
}


/**
 * Updates the page and title fields on the tracker if necessary and
 * optionally sends a pageview.
 * @param {boolean} shouldSendPageview Indicates whether the tracker should
 *     send a pageview after updating the URL.
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

      if (shouldSendPageview) {
        var defaultFields = {transport: 'beacon'};
        this.tracker.send('pageview', createFieldsObj(defaultFields,
            this.opts.fieldsObj, this.tracker, this.opts.hitFilter));
      }
    }
  }.bind(this), 0);
};


/**
 * Determines whether or not the tracker should send a hit with the new page
 * data. This default implementation can be overrided in the config options.
 * @param {string} newPath The path prior to the URL change.
 * @param {string} oldPath The path after the URL change.
 * @return {boolean} Whether or not the URL change should be tracked.
 */
UrlChangeTracker.prototype.shouldTrackUrlChange = function(newPath, oldPath) {
  return newPath && oldPath;
};


/**
 * Removes all event listeners and instance properties.
 */
UrlChangeTracker.prototype.remove = function() {
  window.removeEventListener('popstate', this.updateTrackerData);
  history.replaceState = this.originalReplaceState;
  history.pushState = this.originalPushState;

  this.tracker = null;
  this.opts = null;
  this.path = null;

  this.updateTrackerData = null;
  this.originalReplaceState = null;
  this.originalPushState = null;
};


/**
 * @return {string} The path value of the current URL.
 */
function getPath() {
  return location.pathname + location.search;
}


provide('urlChangeTracker', UrlChangeTracker);
