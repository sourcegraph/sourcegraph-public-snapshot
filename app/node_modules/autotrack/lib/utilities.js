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


var utilities = {

  /**
   * Accepts a function and returns a wrapped version of the function that is
   * expected to be called elsewhere in the system. If it's not called
   * elsewhere after the timeout period, it's called regardless. The wrapper
   * function also prevents the callback from being called more than once.
   * @param {Function} callback The function to call.
   * @param {number} wait How many milliseconds to wait before invoking
   *     the callback.
   */
  withTimeout: function(callback, wait) {
    var called = false;
    setTimeout(callback, wait || 2000);
    return function() {
      if (!called) {
        called = true;
        callback();
      }
    };
  },


  /**
   * Accepts an object of overrides and defaults and returns a new object
   * with the values merged. For each key in defaults, if there's a
   * corresponding value in overrides, it gets used.
   * @param {Object} overrides.
   * @param {?Object} defaults.
   */
  defaults: function(overrides, defaults) {
    var result = {};

    if (typeof overrides != 'object') {
      overrides = {};
    }

    if (typeof defaults != 'object') {
      defaults = {};
    }

    for (var key in defaults) {
      if (defaults.hasOwnProperty(key)) {
        result[key] = overrides.hasOwnProperty(key) ?
            overrides[key] : defaults[key];
      }
    }
    return result;
  },


  isObject: function(obj) {
    return typeof obj == 'object' && obj !== null;
  },


  isArray: Array.isArray || function(arr) {
    return Object.prototype.toString.call(arr) === '[object Array]';
  },


  toArray: function(obj) {
    return utilities.isArray(obj) ? obj : [obj];
  }
};

module.exports = utilities;
