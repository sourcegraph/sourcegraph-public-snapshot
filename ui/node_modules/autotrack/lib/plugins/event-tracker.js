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
var delegate = require('dom-utils/lib/delegate');
var provide = require('../provide');
var usage = require('../usage');
var createFieldsObj = require('../utilities').createFieldsObj;
var getAttributeFields = require('../utilities').getAttributeFields;


/**
 * Registers declarative event tracking.
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function EventTracker(tracker, opts) {

  usage.track(tracker, usage.plugins.EVENT_TRACKER);

  // Feature detects to prevent errors in unsupporting browsers.
  if (!window.addEventListener) return;

  this.opts = assign({
    events: ['click'],
    fieldsObj: {},
    attributePrefix: 'ga-',
    hitFilter: null
  }, opts);

  this.tracker = tracker;

  // Binds methods.
  this.handleEvents = this.handleEvents.bind(this);

  var selector = '[' + this.opts.attributePrefix + 'on]';

  // Creates a mapping of events to their delegates
  this.delegates = {};
  this.opts.events.forEach(function(event) {
    this.delegates[event] = delegate(document, event, selector,
        this.handleEvents, {deep: true, useCapture: true});
  }.bind(this));
}


/**
 * Handles all clicks on elements with event attributes.
 * @param {Event} event The DOM click event.
 * @param {Element} element The delegated DOM element target.
 */
EventTracker.prototype.handleEvents = function(event, element) {

  var prefix = this.opts.attributePrefix;

  // Ensures the event type matches the one specified on the element.
  if (event.type != element.getAttribute(prefix + 'on')) return;

  var defaultFields = {transport: 'beacon'};
  var attributeFields = getAttributeFields(element, prefix);
  var userFields = assign({}, this.opts.fieldsObj, attributeFields);
  var hitType = attributeFields.hitType || 'event';

  this.tracker.send(hitType, createFieldsObj(
      defaultFields, userFields, this.tracker, this.opts.hitFilter, element));
};


/**
 * Removes all event listeners and instance properties.
 */
EventTracker.prototype.remove = function() {
  Object.keys(this.delegates).forEach(function(key) {
    this.delegates[key].destroy();
  }.bind(this));
};


provide('eventTracker', EventTracker);
