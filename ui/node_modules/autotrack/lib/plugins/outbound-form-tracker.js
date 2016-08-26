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
var parseUrl = require('dom-utils/lib/parse-url');
var provide = require('../provide');
var usage = require('../usage');
var createFieldsObj = require('../utilities').createFieldsObj;
var getAttributeFields = require('../utilities').getAttributeFields;
var withTimeout = require('../utilities').withTimeout;


/**
 * Registers outbound form tracking.
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function OutboundFormTracker(tracker, opts) {

  usage.track(tracker, usage.plugins.OUTBOUND_FORM_TRACKER);

  // Feature detects to prevent errors in unsupporting browsers.
  if (!window.addEventListener) return;

  this.opts = assign({
    formSelector: 'form',
    shouldTrackOutboundForm: this.shouldTrackOutboundForm,
    fieldsObj: {},
    attributePrefix: 'ga-',
    hitFilter: null
  }, opts);

  this.tracker = tracker;

  this.delegate = delegate(document, 'submit', 'form',
      this.handleFormSubmits.bind(this), {deep: true, useCapture: true});
}


/**
 * Handles all submits on form elements. A form submit is considered outbound
 * if its action attribute starts with http and does not contain
 * location.hostname.
 * When the beacon transport method is not available, the event's default
 * action is prevented and re-emitted after the hit is sent.
 * @param {Event} event The DOM submit event.
 * @param {Element} form The delegated event target.
 */
OutboundFormTracker.prototype.handleFormSubmits = function(event, form) {

  var action = parseUrl(form.action).href;
  var defaultFields = {
    transport: 'beacon',
    eventCategory: 'Outbound Form',
    eventAction: 'submit',
    eventLabel: action
  };

  if (this.opts.shouldTrackOutboundForm(form, parseUrl)) {

    if (!navigator.sendBeacon) {
      // Stops the submit and waits until the hit is complete (with timeout)
      // for browsers that don't support beacon.
      event.preventDefault();
      defaultFields.hitCallback = withTimeout(function() {
        form.submit();
      });
    }

    var userFields = assign({}, this.opts.fieldsObj,
        getAttributeFields(form, this.opts.attributePrefix));

    this.tracker.send('event', createFieldsObj(
        defaultFields, userFields, this.tracker, this.opts.hitFilter, form));
  }
};


/**
 * Determines whether or not the tracker should send a hit when a form is
 * submitted. By default, forms with an action attribute that starts with
 * "http" and doesn't contain the current hostname are tracked.
 * @param {Element} form The form that was submitted.
 * @param {Function} parseUrl A cross-browser utility method for url parsing.
 * @return {boolean} Whether or not the form should be tracked.
 */
OutboundFormTracker.prototype.shouldTrackOutboundForm =
    function(form, parseUrl) {

  var url = parseUrl(form.action);
  return url.hostname != location.hostname &&
      url.protocol.slice(0, 4) == 'http';
};


/**
 * Removes all event listeners and instance properties.
 */
OutboundFormTracker.prototype.remove = function() {
  this.delegate.destroy();
};


provide('outboundFormTracker', OutboundFormTracker);
