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
var delegate = require('delegate');
var provide = require('../provide');


/**
 * Registers social tracking on tracker object.
 * Supports both declarative social tracking via HTML attributes as well as
 * tracking for social events when using official Twitter or Facebook widgets.
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function SocialTracker(tracker, opts) {

  // Registers the plugin on the global gaplugins object.
  window.gaplugins = window.gaplugins || {};
  gaplugins.SocialTracker = SocialTracker;

  // Feature detects to prevent errors in unsupporting browsers.
  if (!window.addEventListener) return;

  this.opts = defaults(opts, {
    attributePrefix: 'data-'
  });

  this.tracker = tracker;

  var prefix = this.opts.attributePrefix;
  var selector = '[' + prefix + 'social-network]' +
                 '[' + prefix + 'social-action]' +
                 '[' + prefix + 'social-target]';

  delegate(document, selector, 'click', this.handleSocialClicks.bind(this));

  this.detectLibraryLoad('FB', 'facebook-jssdk',
      this.addFacebookEventHandlers.bind(this));

  this.detectLibraryLoad('twttr', 'twitter-wjs',
      this.addTwitterEventHandlers.bind(this));
}


/**
 * Handles all clicks on elements with social tracking attributes.
 * @param {Event} event The DOM click event.
 */
SocialTracker.prototype.handleSocialClicks = function(event) {

  var link = event.delegateTarget;
  var prefix = this.opts.attributePrefix;

  this.tracker.send('social', {
    socialNetwork: link.getAttribute(prefix + 'social-network'),
    socialAction: link.getAttribute(prefix + 'social-action'),
    socialTarget: link.getAttribute(prefix + 'social-target')
  });
};


/**
 * A utility method that determines when a social library is finished loading.
 * @param {string} namespace The global property name added to window.
 * @param {string} domId The ID of a script element in the DOM.
 * @param {Function} done A callback to be invoked when done.
 */
SocialTracker.prototype.detectLibraryLoad = function(namespace, domId, done) {
  if (window[namespace]) {
    done();
  }
  else {
    var script = document.getElementById(domId);
    if (script) {
      script.onload = done;
    }
  }
};


/**
 * Adds event handlers for the "tweet" and "follow" events emitted by the
 * official tweet and follow buttons. Note: this does not capture tweet or
 * follow events emitted by other Twitter widgets (tweet, timeline, etc.).
 */
SocialTracker.prototype.addTwitterEventHandlers = function() {
  try {
    twttr.ready(function() {

      twttr.events.bind('tweet', function(event) {
        // Ignore tweets from widgets that aren't the tweet button.
        if (event.region != 'tweet') return;

        var url = event.data.url || event.target.getAttribute('data-url') ||
            location.href;

        this.tracker.send('social', 'Twitter', 'tweet', url);
      }.bind(this));

      twttr.events.bind('follow', function(event) {
        // Ignore follows from widgets that aren't the follow button.
        if (event.region != 'follow') return;

        var screenName = event.data.screen_name ||
            event.target.getAttribute('data-screen-name');

        this.tracker.send('social', 'Twitter', 'follow', screenName);
      }.bind(this));
    }.bind(this));
  } catch(err) {}
};


/**
 * Adds event handlers for the "like" and "unlike" events emitted by the
 * official Facebook like button.
 */
SocialTracker.prototype.addFacebookEventHandlers = function() {
  try {
    FB.Event.subscribe('edge.create', function(url) {
      this.tracker.send('social', 'Facebook', 'like', url);
    }.bind(this));

    FB.Event.subscribe('edge.remove', function(url) {
      this.tracker.send('social', 'Facebook', 'unlike', url);
    }.bind(this));
  } catch(err) {}
};


provide('socialTracker', SocialTracker);
