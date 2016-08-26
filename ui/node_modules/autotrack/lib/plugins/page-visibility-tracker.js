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


var DEFAULT_SESSION_TIMEOUT = 30; // 30 minutes.


/**
 * Registers outbound link tracking on tracker object.
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function PageVisibilityTracker(tracker, opts) {

  usage.track(tracker, usage.plugins.PAGE_VISIBILITY_TRACKER);

  // Feature detects to prevent errors in unsupporting browsers.
  if (!window.addEventListener) return;

  this.opts = assign({
    sessionTimeout: DEFAULT_SESSION_TIMEOUT,
    changeTemplate: this.changeTemplate,
    hiddenMetricIndex: null,
    visibleMetricIndex: null,
    fieldsObj: {},
    hitFilter: null
  }, opts);

  this.tracker = tracker;
  this.visibilityState = document.visibilityState;

  // Consider the plugin creation to be the start of the visibility change
  // time calculations.
  this.lastVisibilityChangeTime = +new Date;

  // Binds methods to `this`.
  this.handleVisibilityStateChange =
      this.handleVisibilityStateChange.bind(this);

  this.overrideTrackerSendMethod();
  this.overrideTrackerSendHitTask();

  document.addEventListener(
      'visibilitychange', this.handleVisibilityStateChange);
}


/**
 * Handles changes to `document.visibilityState`. This method sends events when
 * the visibility state changes during active sessions (active meaning the
 * session has not timed out). If the session has timed out, a return to a
 * visibility state of visible will trigger a new pageview (instead of a
 * visibility change event). Lastly, this method keeps track of the elapsed
 * time a document's visibility state was visible and sends that as the event
 * value for hidden events, allowing you to more accurately derive how long
 * a user spent active during a session.
 */
PageVisibilityTracker.prototype.handleVisibilityStateChange = function() {

  var defaultFields;
  this.prevVisibilityState = this.visibilityState;
  this.visibilityState = document.visibilityState;

  if (this.sessionHasTimedOut()) {
    // Prevents sending 'hidden' state hits when the session has timed out.
    if (this.visibilityState == 'hidden') return;

    if (this.visibilityState == 'visible') {
      // If the session has timed out, a transition to "visible" should be
      // considered a new pageview and a new session.
      defaultFields = {transport: 'beacon'};
      this.tracker.send('pageview', createFieldsObj(defaultFields,
          this.opts.fieldsObj, this.tracker, this.opts.hitFilter));
    }
  }
  else {
    // Rounds the time up to the nearest second. If the rounded value is zero
    // use 1 instead since unset metrics default to 0.
    var timeDeltaInSeconds = Math.round(
        (new Date - this.lastVisibilityChangeTime) / 1000) || 1;

    defaultFields = {
      transport: 'beacon',
      eventCategory: 'Page Visibility',
      eventAction: 'change',
      eventLabel: this.opts.changeTemplate(
          this.prevVisibilityState, this.visibilityState),
      eventValue: timeDeltaInSeconds
    };

    // Changes to hidden are non interaction hits by default
    if (this.visibilityState == 'hidden') defaultFields.nonInteraction = true;

    // If a custom metric was specified for the current visibility state,
    // give it the same as the event value.
    var metric = this.opts[this.prevVisibilityState + 'MetricIndex'];
    if (metric) defaultFields['metric' + metric] = timeDeltaInSeconds;

    this.tracker.send('event', createFieldsObj(defaultFields,
        this.opts.fieldsObj, this.tracker, this.opts.hitFilter));
  }

  // Updates the time the last visibility state change event occurred, so
  // change events can report the delta.
  this.lastVisibilityChangeTime = +new Date;
};


/**
 * Returns true if the session has not timed out. A session timeout occurs when
 * more than `this.opts.sessionTimeout` minutes has elapsed since the
 * tracker sent the previous hit.
 * @return {boolean} True if the session has timed out.
 */
PageVisibilityTracker.prototype.sessionHasTimedOut = function() {
  var minutesSinceLastHit = (new Date - this.lastHitTime) / (60 * 1000);
  return this.opts.sessionTimeout < minutesSinceLastHit;
};


/**
 * Overrides the `tracker.send` method to send a pageview hit before the
 * current hit being sent if the session has timed out and the current hit is
 * not a pageview itself.
 */
PageVisibilityTracker.prototype.overrideTrackerSendMethod = function() {
  this.originalTrackerSendMethod = this.tracker.send;

  this.tracker.send = function() {
    var args = Array.prototype.slice.call(arguments);
    var firstArg = args[0];
    var hitType = isObject(firstArg) ? firstArg.hitType : firstArg;
    var isPageview = hitType == 'pageview';

    if (!isPageview && this.sessionHasTimedOut()) {
      var defaultFields = {transport: 'beacon'};
      this.originalTrackerSendMethod.call(this.tracker, 'pageview',
          createFieldsObj(defaultFields, this.opts.fieldsObj,
              this.tracker, this.opts.hitFilter));
    }

    this.originalTrackerSendMethod.apply(this.tracker, args);
  }.bind(this);
};


/**
 * Overrides the tracker's `sendHitTask` to record the time of the previous
 * hit. This is used to determine whether or not a session has timed out.
 */
PageVisibilityTracker.prototype.overrideTrackerSendHitTask = function() {
  this.originalTrackerSendHitTask = this.tracker.get('sendHitTask');
  this.lastHitTime = +new Date;

  this.tracker.set('sendHitTask', function(model) {
    this.originalTrackerSendHitTask(model);
    this.lastHitTime = +new Date;
  }.bind(this));
};


/**
 * Sets the default formatting of the change event label.
 * This can be overridden by setting the `changeTemplate` option.
 * @param {string} oldValue The value of the media query prior to the change.
 * @param {string} newValue The value of the media query after the change.
 * @return {string} The formatted event label.
 */
PageVisibilityTracker.prototype.changeTemplate = function(oldValue, newValue) {
  return oldValue + ' => ' + newValue;
};


/**
 * Removes all event listeners and instance properties.
 */
 PageVisibilityTracker.prototype.remove = function() {
  this.tracker.set('sendHitTask', this.originalTrackerSendHitTask);
  this.tracker.send = this.originalTrackerSendMethod;

  document.removeEventListener(
      'visibilitychange', this.handleVisibilityStateChange);
};


provide('pageVisibilityTracker', PageVisibilityTracker);
