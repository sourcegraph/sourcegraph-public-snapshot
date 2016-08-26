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
var getAttributes = require('dom-utils/lib/get-attributes');


var utilities = {


  /**
   * Accepts default and user override fields and an optional tracker, hit
   * filter, and target element and returns a single object that can be used in
   * `ga('send', ...)` commands.
   * @param {Object} defaultFields The default fields to return.
   * @param {Object} userFields Fields set by the user to override the defaults.
   * @param {Object} opt_tracker The tracker object to apply the hit filter to.
   * @param {Function} opt_hitFilter A filter function that gets
   *     called with the tracker model right before the `buildHitTask`. It can
   *     be used to modify the model for the current hit only.
   * @param {Element} opt_target If the hit originated from an interaction
   *     with a DOM element, hitFilter is invoked with that element as the
   *     second argument.
   * @return {Object} The final fields object.
   */
  createFieldsObj: function(
      defaultFields, userFields, opt_tracker, opt_hitFilter, opt_target) {

    if (typeof opt_hitFilter == 'function') {
      var originalBuildHitTask = opt_tracker.get('buildHitTask');
      return {
        buildHitTask: function(model) {
          model.set(defaultFields, null, true);
          model.set(userFields, null, true);
          opt_hitFilter(model, opt_target);
          originalBuildHitTask(model);
        }
      };
    }
    else {
      return assign({}, defaultFields, userFields);
    }
  },


  /**
   * Retrieves the attributes from an DOM element and returns a fields object
   * for all attributes matching the passed prefix string.
   * @param {Element} element The DOM element to get attributes from.
   * @param {string} prefix An attribute prefix. Only the attributes matching
   *     the prefix will be returned on the fields object.
   * @return {Object} An object of analytics.js fields and values
   */
  getAttributeFields: function(element, prefix) {
    var attributes = getAttributes(element);
    var attributeFields = {};

    Object.keys(attributes).forEach(function(attribute) {

      // The `on` prefix is used for event handling but isn't a field.
      if (attribute.indexOf(prefix) === 0 && attribute != prefix + 'on') {

        var value = attributes[attribute];

        // Detects Boolean value strings.
        if (value == 'true') value = true;
        if (value == 'false') value = false;

        var field = utilities.camelCase(attribute.slice(prefix.length));
        attributeFields[field] = value;
      }
    });

    return attributeFields;
  },


  domReady: function(callback) {
    if (document.readyState == 'loading') {
      document.addEventListener('DOMContentLoaded', function fn() {
        document.removeEventListener('DOMContentLoaded', fn);
        callback();
      });
    } else {
      callback();
    }
  },


  /**
   * Accepts a function and returns a wrapped version of the function that is
   * expected to be called elsewhere in the system. If it's not called
   * elsewhere after the timeout period, it's called regardless. The wrapper
   * function also prevents the callback from being called more than once.
   * @param {Function} callback The function to call.
   * @param {number} wait How many milliseconds to wait before invoking
   *     the callback.
   * @returns {Function} The wrapped version of the passed function.
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
   * Accepts a string containing hyphen or underscore word separators and
   * converts it to camelCase.
   * @param {string} str The string to camelCase.
   * @return {string} The camelCased version of the string.
   */
  camelCase: function(str) {
    return str.replace(/[\-\_]+(\w?)/g, function(match, p1) {
      return p1.toUpperCase();
    });
  },


  /**
   * Capitalizes the first letter of a string.
   * @param {string} str The input string.
   * @return {string} The capitalized string
   */
  capitalize: function(str) {
    return str.charAt(0).toUpperCase() + str.slice(1);
  },


  /**
   * Indicates whether the passed variable is a JavaScript object.
   * @param {*} value The input variable to test.
   * @return {boolean} Whether or not the test is an object.
   */
  isObject: function(value) {
    return typeof value == 'object' && value !== null;
  },


  /**
   * Indicates whether the passed variable is a JavaScript array.
   * @param {*} value The input variable to test.
   * @return {boolean} Whether or not the value is an array.
   */
  isArray: Array.isArray || function(value) {
    return Object.prototype.toString.call(value) === '[object Array]';
  },


  /**
   * Accepts a value that may or may not be an array. If it is not an array,
   * it is returned as the first item in a single-item array.
   * @param {*} value The value to convert to an array if it is not.
   * @return {Array} The array-ified value.
   */
  toArray: function(value) {
    return utilities.isArray(value) ? value : [value];
  }
};

module.exports = utilities;
