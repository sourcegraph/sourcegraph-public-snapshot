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


var constants = require('./constants');


// Adds the dev ID to the list of dev IDs if any plugin is used.
(window.gaDevIds = window.gaDevIds || []).push(constants.DEV_ID);


/**
 * Provides a plugin for use with analytics.js, accounting for the possibility
 * that the global command queue has been renamed.
 * @param {string} pluginName The plugin name identifier.
 * @param {Function} pluginConstructor The plugin constructor function.
 */
module.exports = function providePlugin(pluginName, pluginConstructor) {
  var w = window;
  var g = w.GoogleAnalyticsObject || 'ga';

  // Creates the global command queue if it's not defined.
  w[g] = w[g] || function(){(w[g].q=w[g].q||[]).push(arguments)};
  w[g].l = w[g].l || +new Date;

  w[g]('provide', pluginName, pluginConstructor);
};
