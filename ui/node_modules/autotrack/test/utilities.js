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


module.exports = {

  /**
   * @param {string} expectedUrl The URL to match.
   * @return {Function} A function that, when invoked, returns a promise
   *     that is fulfilled when the URL in the browsers address bar matches
   *     the passed URL.
   */
  urlMatches: function(expectedUrl) {
    return function() {
      var result = browser.url();
      var actualUrl = result.value;
      return actualUrl.indexOf(expectedUrl) > -1;
    };
  },


  /**
   * Prevents the default form submit action allowing forms to be interacted
   * with without navigating away from the current page.
   */
  stopSubmitEvents: function() {
    window.__stopFormSubmits__ = function(event) {
      event.preventDefault();
    };

    document.addEventListener('submit', window.__stopFormSubmits__);
  },


  /**
   * Restores normal form submit behavior.
   */
  unstopSubmitEvents: function() {
    document.removeEventListener('submit', window.__stopFormSubmits__);
  },


  /**
   * Sets all form element submit methods to a noop.
   */
  disableProgramaticFormSubmits: function() {
    for (var i = 0, form; form = document.forms[i]; i++) {
      form.submit = function() {};
    }
  },


  /**
   * Prevents the default link click action allowing links to be interacted
   * with without navigating away from the current page.
   */
  stopClickEvents: function() {
    window.__stopClicks__ = function(event) {
      event.preventDefault();
    };

    document.addEventListener('click', window.__stopClicks__);
  },


  /**
   * Restores normal link click behavior.
   */
  unstopClickEvents: function() {
    document.removeEventListener('click', window.__stopClicks__);
  },


  /**
   * Assigns a function to navigator.sendBeacon so analytics.js assumes support
   * for the beacon transport mechanism.
   */
  stubBeacon: function() {
    navigator.sendBeacon = function() {
      return true;
    };
  },


  /**
   * Unsets navigator.sendBeacon so analytics.js assumes no support for the
   * beacon transport mechanism.
   */
  stubNoBeacon: function() {
    navigator.sendBeacon = undefined;
  },


  /**
   * Wraps `console.error` and tracks calls to it.
   */
  trackConsoleErrors: function() {
    if (!window.console) return;
    window.__consoleErrors__ = [];
    window.__originalConsoleError__ = window.console.error;
    window.console.error = function() {
      window.__consoleErrors__.push(arguments);
      window.__originalConsoleError__.apply(window.console, arguments);
    };
  },


  /**
   * Restores the original `console.error`.
   */
  untrackConsoleErrors: function() {
    if (!window.console) return;
    delete window.__consoleErrors__;
    window.console.error = window.__originalConsoleError__ ||
        window.console.error;
  },


  /**
   * Returns all console error call arguments since tracking started.
   * @return {Array} The list of console error call arguments.
   */
  getConsoleErrors: function() {
    return window.__consoleErrors__;
  }

};
