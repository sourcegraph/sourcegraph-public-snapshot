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


/**
 * Registers outbound link tracking on a tracker object.
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function OutboundLinkTracker(tracker, opts) {

  usage.track(tracker, usage.plugins.OUTBOUND_LINK_TRACKER);

  // Feature detects to prevent errors in unsupporting browsers.
  if (!window.addEventListener) return;

  this.opts = assign({
    events: ['click'],
    linkSelector: 'a',
    shouldTrackOutboundLink: this.shouldTrackOutboundLink,
    fieldsObj: {},
    attributePrefix: 'ga-',
    hitFilter: null
  }, opts);

  this.tracker = tracker;

  // Binds methods.
  this.handleLinkInteractions = this.handleLinkInteractions.bind(this);

  // Creates a mapping of events to their delegates
  this.delegates = {};
  this.opts.events.forEach(function(event) {
    this.delegates[event] = delegate(document, event, this.opts.linkSelector,
        this.handleLinkInteractions, {deep: true, useCapture: true});
  }.bind(this));
}


/**
 * Handles all interactions on link elements. A link is considered an outbound
 * link if its hostname property does not match location.hostname. When the
 * beacon transport method is not available, the links target is set to
 * "_blank" to ensure the hit can be sent.
 * @param {Event} event The DOM click event.
 * @param {Element} link The delegated event target.
 */
OutboundLinkTracker.prototype.handleLinkInteractions = function(event, link) {

  if (this.opts.shouldTrackOutboundLink(link, parseUrl)) {
    // Opens outbound links in a new tab if the browser doesn't support
    // the beacon transport method.
    if (!navigator.sendBeacon) {
      link.target = '_blank';
    }

    var defaultFields = {
      transport: 'beacon',
      eventCategory: 'Outbound Link',
      eventAction: event.type,
      eventLabel: link.href
    };

    var userFields = assign({}, this.opts.fieldsObj,
        getAttributeFields(link, this.opts.attributePrefix));

    this.tracker.send('event', createFieldsObj(
        defaultFields, userFields, this.tracker, this.opts.hitFilter, link));
  }
};


/**
 * Determines whether or not the tracker should send a hit when a link is
 * clicked. By default links with a hostname property not equal to the current
 * hostname are tracked.
 * @param {Element} link The link that was clicked on.
 * @param {Function} parseUrl A cross-browser utility method for url parsing.
 * @return {boolean} Whether or not the link should be tracked.
 */
OutboundLinkTracker.prototype.shouldTrackOutboundLink =
    function(link, parseUrl) {

  var url = parseUrl(link.href);
  return url.hostname != location.hostname &&
      url.protocol.slice(0, 4) == 'http';
};


/**
 * Removes all event listeners and instance properties.
 */
OutboundLinkTracker.prototype.remove = function() {
  Object.keys(this.delegates).forEach(function(key) {
    this.delegates[key].destroy();
  }.bind(this));
};


provide('outboundLinkTracker', OutboundLinkTracker);
