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
var debounce = require('debounce');
var constants = require('../constants');
var provide = require('../provide');
var usage = require('../usage');
var createFieldsObj = require('../utilities').createFieldsObj;
var isObject = require('../utilities').isObject;
var toArray = require('../utilities').toArray;


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

  usage.track(tracker, usage.plugins.MEDIA_QUERY_TRACKER);

  // Feature detects to prevent errors in unsupporting browsers.
  if (!window.matchMedia) return;

  this.opts = assign({
    definitions: null,
    changeTemplate: this.changeTemplate,
    changeTimeout: 1000,
    fieldsObj: {},
    hitFilter: null
  }, opts);

  // Exits early if media query data doesn't exist.
  if (!isObject(this.opts.definitions)) return;

  this.opts.definitions = toArray(this.opts.definitions);
  this.tracker = tracker;
  this.changeListeners = [];

  this.processMediaQueries();
}


/**
 * Loops through each media query definition, sets the custom dimenion data,
 * and adds the change listeners.
 */
MediaQueryTracker.prototype.processMediaQueries = function() {
  this.opts.definitions.forEach(function(definition) {
    // Only processes definitions with a name and index.
    if (definition.name && definition.dimensionIndex) {
      var mediaName = this.getMatchName(definition);
      this.tracker.set('dimension' + definition.dimensionIndex, mediaName);

      this.addChangeListeners(definition);
    }
  }.bind(this));
};


/**
 * Takes a definition object and return the name of the matching media item.
 * If no match is found, the NULL_DIMENSION value is returned.
 * @param {Object} definition A set of named media queries associated
 *     with a single custom dimension.
 * @return {string} The name of the matched media or NULL_DIMENSION.
 */
MediaQueryTracker.prototype.getMatchName = function(definition) {
  var match;

  definition.items.forEach(function(item) {
    if (getMediaListener(item.media).matches) {
      match = item;
    }
  });
  return match ? match.name : constants.NULL_DIMENSION;
};


/**
 * Adds change listeners to each media query in the definition list.
 * Debounces the changes to prevent unnecessary hits from being sent.
 * @param {Object} definition A set of named media queries associated
 *     with a single custom dimension
 */
MediaQueryTracker.prototype.addChangeListeners = function(definition) {
  definition.items.forEach(function(item) {
    var mql = getMediaListener(item.media);
    var fn = debounce(function() {
      this.handleChanges(definition);
    }.bind(this), this.opts.changeTimeout);

    mql.addListener(fn);
    this.changeListeners.push({mql: mql, fn: fn});
  }.bind(this));
};


/**
 * Handles changes to the matched media. When the new value differs from
 * the old value, a change event is sent.
 * @param {Object} definition A set of named media queries associated
 *     with a single custom dimension
 */
MediaQueryTracker.prototype.handleChanges = function(definition) {
  var newValue = this.getMatchName(definition);
  var oldValue = this.tracker.get('dimension' + definition.dimensionIndex);

  if (newValue !== oldValue) {
    this.tracker.set('dimension' + definition.dimensionIndex, newValue);

    var defaultFields = {
      eventCategory: definition.name,
      eventAction: 'change',
      eventLabel: this.opts.changeTemplate(oldValue, newValue)
    };
    this.tracker.send('event', createFieldsObj(
        defaultFields, this.opts.fieldsObj, this.tracker, this.opts.hitFilter));
  }
};


/**
 * Removes all event listeners and instance properties.
 */
MediaQueryTracker.prototype.remove = function() {
  for (var i = 0, listener; listener = this.changeListeners[i]; i++) {
    listener.mql.removeListener(listener.fn);
  }
};


/**
 * Sets the default formatting of the change event label.
 * This can be overridden by setting the `changeTemplate` option.
 * @param {string} oldValue The value of the media query prior to the change.
 * @param {string} newValue The value of the media query after the change.
 * @return {string} The formatted event label.
 */
MediaQueryTracker.prototype.changeTemplate = function(oldValue, newValue) {
  return oldValue + ' => ' + newValue;
};


/**
 * Accepts a media query and returns a MediaQueryListener object.
 * Caches the values to avoid multiple unnecessary instances.
 * @param {string} media A media query value.
 * @return {MediaQueryListener} The matched media.
 */
function getMediaListener(media) {
  // Returns early if the media is cached.
  if (mediaMap[media]) return mediaMap[media];

  mediaMap[media] = window.matchMedia(media);
  return mediaMap[media];
}


provide('mediaQueryTracker', MediaQueryTracker);
