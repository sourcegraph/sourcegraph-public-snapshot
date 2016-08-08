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


var delegate = require('delegate');
var defaults = require('../utilities').defaults;
var provide = require('../provide');


/**
 * Registers declarative event tracking.
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function EventTracker(tracker, opts) {

  // Registers the plugin on the global gaplugins object.
  window.gaplugins = window.gaplugins || {};
  gaplugins.EventTracker = EventTracker;

  // Feature detects to prevent errors in unsupporting browsers.
  if (!window.addEventListener) return;

  this.opts = defaults(opts, {
    attributePrefix: 'data-'
  });

  this.tracker = tracker;

  var prefix = this.opts.attributePrefix;
  var selector = '[' + prefix + 'event-category][' + prefix + 'event-action]';

  delegate(document, selector, 'click', this.handleEventClicks.bind(this));
}


/**
 * Handles all clicks on elements with event attributes.
 * @param {Event} event The DOM click event.
 */
EventTracker.prototype.handleEventClicks = function(event) {

  var link = event.delegateTarget;
  var prefix = this.opts.attributePrefix;

  this.tracker.send('event', {
    eventCategory: link.getAttribute(prefix + 'event-category'),
    eventAction: link.getAttribute(prefix + 'event-action'),
    eventLabel: link.getAttribute(prefix + 'event-label'),
    eventValue: link.getAttribute(prefix + 'event-value')
  });
};


provide('eventTracker', EventTracker);
