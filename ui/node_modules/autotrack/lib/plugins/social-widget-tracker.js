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


/* global FB, twttr */


var assign = require('object-assign');
var provide = require('../provide');
var usage = require('../usage');
var createFieldsObj = require('../utilities').createFieldsObj;


/**
 * Registers social tracking on tracker object.
 * Supports both declarative social tracking via HTML attributes as well as
 * tracking for social events when using official Twitter or Facebook widgets.
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function SocialWidgetTracker(tracker, opts) {

  usage.track(tracker, usage.plugins.SOCIAL_WIDGET_TRACKER);

  // Feature detects to prevent errors in unsupporting browsers.
  if (!window.addEventListener) return;

  this.opts = assign({
    fieldsObj: {},
    hitFilter: null
  }, opts);

  this.tracker = tracker;

  // Binds methods to `this`.
  this.addWidgetListeners = this.addWidgetListeners.bind(this);
  this.addTwitterEventHandlers = this.addTwitterEventHandlers.bind(this);
  this.handleTweetEvents = this.handleTweetEvents.bind(this);
  this.handleFollowEvents = this.handleFollowEvents.bind(this);
  this.handleLikeEvents = this.handleLikeEvents.bind(this);
  this.handleUnlikeEvents = this.handleUnlikeEvents.bind(this);

  if (document.readyState != 'complete') {
    // Adds the widget listeners after the window's `load` event fires.
    // If loading widgets using the officially recommended snippets, they
    // will be available at `window.load`. If not users can call the
    // `addWidgetListeners` method manually.
    window.addEventListener('load', this.addWidgetListeners);
  }
  else {
    this.addWidgetListeners();
  }
}


/**
 * Invokes the methods to add Facebook and Twitter widget event listeners.
 * Ensures the respective global namespaces are present before adding.
 */
SocialWidgetTracker.prototype.addWidgetListeners = function() {
  if (window.FB) this.addFacebookEventHandlers();
  if (window.twttr) this.addTwitterEventHandlers();
};


/**
 * Adds event handlers for the "tweet" and "follow" events emitted by the
 * official tweet and follow buttons. Note: this does not capture tweet or
 * follow events emitted by other Twitter widgets (tweet, timeline, etc.).
 */
SocialWidgetTracker.prototype.addTwitterEventHandlers = function() {
  try {
    twttr.ready(function() {
      twttr.events.bind('tweet', this.handleTweetEvents);
      twttr.events.bind('follow', this.handleFollowEvents);
    }.bind(this));
  } catch(err) {}
};


/**
 * Removes event handlers for the "tweet" and "follow" events emitted by the
 * official tweet and follow buttons.
 */
SocialWidgetTracker.prototype.removeTwitterEventHandlers = function() {
  try {
    twttr.ready(function() {
      twttr.events.unbind('tweet', this.handleTweetEvents);
      twttr.events.unbind('follow', this.handleFollowEvents);
    }.bind(this));
  } catch(err) {}
};


/**
 * Adds event handlers for the "like" and "unlike" events emitted by the
 * official Facebook like button.
 */
SocialWidgetTracker.prototype.addFacebookEventHandlers = function() {
  try {
    FB.Event.subscribe('edge.create', this.handleLikeEvents);
    FB.Event.subscribe('edge.remove', this.handleUnlikeEvents);
  } catch(err) {}
};


/**
 * Removes event handlers for the "like" and "unlike" events emitted by the
 * official Facebook like button.
 */
SocialWidgetTracker.prototype.removeFacebookEventHandlers = function() {
  try {
    FB.Event.unsubscribe('edge.create', this.handleLikeEvents);
    FB.Event.unsubscribe('edge.remove', this.handleUnlikeEvents);
  } catch(err) {}
};


/**
 * Handles `tweet` events emitted by the Twitter JS SDK.
 * @param {Object} event The Twitter event object passed to the handler.
 */
SocialWidgetTracker.prototype.handleTweetEvents = function(event) {
  // Ignores tweets from widgets that aren't the tweet button.
  if (event.region != 'tweet') return;

  var url = event.data.url || event.target.getAttribute('data-url') ||
      location.href;

  var defaultFields = {
    transport: 'beacon',
    socialNetwork: 'Twitter',
    socialAction: 'tweet',
    socialTarget: url
  };
  this.tracker.send('social', createFieldsObj(defaultFields,
      this.opts.fieldsObj, this.tracker, this.opts.hitFilter));
};


/**
 * Handles `follow` events emitted by the Twitter JS SDK.
 * @param {Object} event The Twitter event object passed to the handler.
 */
SocialWidgetTracker.prototype.handleFollowEvents = function(event) {
  // Ignore follows from widgets that aren't the follow button.
  if (event.region != 'follow') return;

  var screenName = event.data.screen_name ||
      event.target.getAttribute('data-screen-name');

  var defaultFields = {
    transport: 'beacon',
    socialNetwork: 'Twitter',
    socialAction: 'follow',
    socialTarget: screenName
  };
  this.tracker.send('social', createFieldsObj(defaultFields,
      this.opts.fieldsObj, this.tracker, this.opts.hitFilter));
};


/**
 * Handles `like` events emitted by the Facebook JS SDK.
 * @param {string} url The URL corresponding to the like event.
 */
SocialWidgetTracker.prototype.handleLikeEvents = function(url) {
  var defaultFields = {
    transport: 'beacon',
    socialNetwork: 'Facebook',
    socialAction: 'like',
    socialTarget: url
  };
  this.tracker.send('social', createFieldsObj(defaultFields,
      this.opts.fieldsObj, this.tracker, this.opts.hitFilter));
};


/**
 * Handles `unlike` events emitted by the Facebook JS SDK.
 * @param {string} url The URL corresponding to the unlike event.
 */
SocialWidgetTracker.prototype.handleUnlikeEvents = function(url) {
  var defaultFields = {
    transport: 'beacon',
    socialNetwork: 'Facebook',
    socialAction: 'unlike',
    socialTarget: url
  };
  this.tracker.send('social', createFieldsObj(defaultFields,
      this.opts.fieldsObj, this.tracker, this.opts.hitFilter));
};


/**
 * Removes all event listeners and instance properties.
 */
SocialWidgetTracker.prototype.remove = function() {
  window.removeEventListener('load', this.addWidgetListeners);
  this.removeFacebookEventHandlers();
  this.removeTwitterEventHandlers();
};


provide('socialWidgetTracker', SocialWidgetTracker);
