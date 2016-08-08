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


var debounce = require('debounce');
var defaults = require('../utilities').defaults;
var isObject = require('../utilities').isObject;
var toArray = require('../utilities').toArray;
var provide = require('../provide');


/**
 * Sets the string to use when no custom dimension value is available.
 */
var NULL_DIMENSION = '(not set)';


/**
 * Declares the MediaQueryListener instance cache.
 */
var mediaMap = {};


/**
 * Registers media query tracking.
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function MediaQueryTracker(tracker, opts) {

  // Registers the plugin on the global gaplugins object.
  window.gaplugins = window.gaplugins || {};
  gaplugins.MediaQueryTracker = MediaQueryTracker;

  // Feature detects to prevent errors in unsupporting browsers.
  if (!window.matchMedia) return;

  this.opts = defaults(opts, {
    mediaQueryDefinitions: false,
    mediaQueryChangeTemplate: this.changeTemplate,
    mediaQueryChangeTimeout: 1000
  });

  // Exits early if media query data doesn't exist.
  if (!isObject(this.opts.mediaQueryDefinitions)) return;

  this.opts.mediaQueryDefinitions = toArray(this.opts.mediaQueryDefinitions);
  this.tracker = tracker;
  this.timeouts = {};

  this.processMediaQueries();
}


/**
 * Loops through each media query definition, sets the custom dimenion data,
 * and adds the change listeners.
 */
MediaQueryTracker.prototype.processMediaQueries = function() {
  this.opts.mediaQueryDefinitions.forEach(function(dimension) {

    if (!dimension.dimensionIndex) {
      throw new Error('Media query definitions must have a name.');
    }

    if (!dimension.dimensionIndex) {
      throw new Error('Media query definitions must have a dimension index.');
    }

    var name = this.getMatchName(dimension);
    this.tracker.set('dimension' + dimension.dimensionIndex, name);

    this.addChangeListeners(dimension);
  }.bind(this));
};


/**
 * Takes a dimension object and return the name of the matching media item.
 * If no match is found, the NULL_DIMENSION value is returned.
 * @param {Object} dimension A set of named media queries associated
 *     with a single custom dimension.
 * @return {string} The name of the matched media or NULL_DIMENSION.
 */
MediaQueryTracker.prototype.getMatchName = function(dimension) {
  var match;

  dimension.items.forEach(function(item) {
    if (getMediaListener(item.media).matches) {
      match = item;
    }
  });
  return match ? match.name : NULL_DIMENSION;
};


/**
 * Adds change listeners to each media query in the dimension list.
 * Debounces the changes to prevent unnecessary hits from being sent.
 * @param {Object} dimension A set of named media queries associated
 *     with a single custom dimension
 */
MediaQueryTracker.prototype.addChangeListeners = function(dimension) {
  dimension.items.forEach(function(item) {
    var mql = getMediaListener(item.media);
    mql.addListener(debounce(function() {
      this.handleChanges(dimension);
    }.bind(this), this.opts.mediaQueryChangeTimeout));
  }.bind(this));
};


/**
 * Handles changes to the matched media. When the new value differs from
 * the old value, a change event is sent.
 * @param {Object} dimension A set of named media queries associated
 *     with a single custom dimension
 */
MediaQueryTracker.prototype.handleChanges = function(dimension) {
  var newValue = this.getMatchName(dimension);
  var oldValue = this.tracker.get('dimension' + dimension.dimensionIndex);

  if (newValue !== oldValue) {
    this.tracker.set('dimension' + dimension.dimensionIndex, newValue);
    this.tracker.send('event', dimension.name, 'change',
        this.opts.mediaQueryChangeTemplate(oldValue, newValue));
  }
};


/**
 * Sets the default formatting of the change event label.
 * This can be overridden by setting the `mediaQueryChangeTemplate` option.
 * @param {string} oldValue
 * @param {string} newValue
 * @return {string} The formatted event label.
 */
MediaQueryTracker.prototype.changeTemplate = function(oldValue, newValue) {
  return oldValue + ' => ' + newValue;
};



/**
 * Accepts a media query and returns a MediaQueryListener object.
 * Caches the values to avoid multiple unnecessary instances.
 * @param {string} media A media query value.
 * @return {MediaQueryListener}
 */
function getMediaListener(media) {
  // Returns early if the media is cached.
  if (mediaMap[media]) return mediaMap[media];

  mediaMap[media] = window.matchMedia(media);
  return mediaMap[media];
}


provide('mediaQueryTracker', MediaQueryTracker);
