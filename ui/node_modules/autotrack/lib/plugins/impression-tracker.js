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
var domReady = require('../utilities').domReady;
var getAttributeFields = require('../utilities').getAttributeFields;


/**
 * Registers impression tracking.
 * @constructor
 * @param {Object} tracker Passed internally by analytics.js
 * @param {?Object} opts Passed by the require command.
 */
function ImpressionTracker(tracker, opts) {

  usage.track(tracker, usage.plugins.IMPRESSION_TRACKER);

  // Feature detects to prevent errors in unsupporting browsers.
  if (!(window.IntersectionObserver && window.MutationObserver)) return;

  this.opts = assign({
    elements: [],
    rootMargin: '0px',
    fieldsObj: {},
    attributePrefix: 'ga-',
    hitFilter: null
  }, opts);

  this.tracker = tracker;

  // Binds methods.
  this.handleDomMutations = this.handleDomMutations.bind(this);
  this.walkNodeTree = this.walkNodeTree.bind(this);
  this.handleIntersectionChanges = this.handleIntersectionChanges.bind(this);
  this.startObserving = this.startObserving.bind(this);
  this.observeElement = this.observeElement.bind(this);
  this.handleDomElementRemoved = this.handleDomElementRemoved.bind(this);

  var data = this.deriveDataFromConfigOptions();

  // The primary list of elements to observe. Each item contains the
  // element ID, threshold, and whether it's currently in-view.
  this.items = data.items;

  // A hash map of elements contained in the items array.
  this.elementMap = data.elementMap;

  // A sorted list of threshold values contained in the items array.
  this.threshold = data.threshold;

  this.intersectionObserver = this.initIntersectionObserver();
  this.mutationObserver = this.initMutationObserver();

  // Once the DOM is ready, start observing for changes.
  domReady(this.startObserving);
}


/**
 * Loops through each element in the `elements` configuration option and
 * creates a map of element IDs currently being observed, a list of "items"
 * (which contains each element's `threshold` and `trackFirstImpressionOnly`
 * property), and a list of `threshold` values to pass to the
 * `IntersectionObserver` instance.
 * @return {Object} An object with the properties `items`, `elementMap`
 *     and `threshold`.
 */
ImpressionTracker.prototype.deriveDataFromConfigOptions = function() {
  var items = [];
  var threshold = [];

  // A map of element IDs in the `items` array to DOM elements in the document.
  // The presence of a key indicates that the element ID is in the `items`
  // array, and the presence of an element value indicates that the element
  // is in the DOM.
  var elementMap = {};

  this.opts.elements.forEach(function(item) {
    // The item can be just a string if it's OK with all the defaults.
    if (typeof item == 'string') item = {id: item};

    items.push(item = assign({
      threshold: 0,
      trackFirstImpressionOnly: true
    }, item));

    elementMap[item.id] = null;
    threshold.push(item.threshold);
  });

  return {
    items: items,
    elementMap: elementMap,
    threshold: threshold
  };
};


/**
 * Initializes a new `MutationObsever` instance and registers the callback.
 * @return {MutationObserver} The new MutationObserver instance.
 */
ImpressionTracker.prototype.initMutationObserver = function() {
  return new MutationObserver(this.handleDomMutations);
};


/**
 * Initializes a new `IntersectionObsever` instance with the appropriate
 * options and registers the callback.
 * @return {IntersectionObserver} The newly created instance.
 */
ImpressionTracker.prototype.initIntersectionObserver = function() {
  return new IntersectionObserver(this.handleIntersectionChanges, {
    rootMargin: this.opts.rootMargin,
    threshold: this.threshold
  });
};


/**
 * Starts observing each eleemnt to intersections as well as the entire DOM
 * for node changes.
 */
ImpressionTracker.prototype.startObserving = function() {
  // Start observing elements for intersections.
  Object.keys(this.elementMap).forEach(this.observeElement);

  // Start observing the DOM for added and removed elements.
  this.mutationObserver.observe(document.body, {
    childList: true,
    subtree: true
  });

  // TODO(philipwalton): Remove temporary hack to force a new frame
  // immediately after adding observers.
  // https://bugs.chromium.org/p/chromium/issues/detail?id=612323
  requestAnimationFrame(function() {});
};


/**
 * Adds an element to the `elementMap` map and registers it for observation
 * on `this.intersectionObserver`.
 * @param {string} id The ID of the element to observe.
 */
ImpressionTracker.prototype.observeElement = function(id) {
  var element = this.elementMap[id] ||
      (this.elementMap[id] = document.getElementById(id));

  if (element) this.intersectionObserver.observe(element);
};


/**
 * Handles nodes being added or removed from the DOM. This function is passed
 * as the callback to `this.mutationObserver`.
 * @param {Array} mutations A list of `MutationRecord` instances
 */
ImpressionTracker.prototype.handleDomMutations = function(mutations) {
  for (var i = 0, mutation; mutation = mutations[i]; i++) {
    // Handles removed elements.
    for (var k = 0, removedEl; removedEl = mutation.removedNodes[k]; k++) {
      this.walkNodeTree(removedEl, this.handleDomElementRemoved);
    }
    // Handles added elements.
    for (var j = 0, addedEl; addedEl = mutation.addedNodes[j]; j++) {
      this.walkNodeTree(addedEl, this.observeElement);
    }
  }
};


/**
 * Iterates through all descendents of a DOM node and invokes the passed
 * callback if any of them match an elememt in `elementMap`.
 * @param {Node} node The DOM node to walk.
 * @param {Function} callback A function to be invoked if a match is found.
 */
ImpressionTracker.prototype.walkNodeTree = function(node, callback) {
  if (node.nodeType == 1 && node.id in this.elementMap) {
    callback(node.id);
  }
  for (var i = 0, child; child = node.childNodes[i]; i++) {
    this.walkNodeTree(child, callback);
  }
};


/**
 * Handles intersection changes. This function is passed as the callback to
 * `this.intersectionObserver`
 * @param {Array} records A list of `IntersectionObserverEntry` records.
 */
ImpressionTracker.prototype.handleIntersectionChanges = function(records) {
  for (var i = 0, record; record = records[i]; i++) {
    for (var j = 0, item; item = this.items[j]; j++) {
      if (record.target.id !== item.id) continue;

      if (isTargetVisible(item.threshold, record)) {
        this.handleImpression(item.id);

        if (item.trackFirstImpressionOnly) {
          this.items.splice(j, 1);
          j--;
          this.possiblyUnobserveElement(item.id);
        }
      }
    }
  }

  // If all items have been removed, remove the plugin.
  if (this.items.length === 0) this.remove();
};


/**
 * Sends a hit to Google Analytics with the impression data.
 * @param {string} id The ID of the element making the impression.
 */
ImpressionTracker.prototype.handleImpression = function(id) {
  var element = document.getElementById(id);

  var defaultFields = {
    transport: 'beacon',
    eventCategory: 'Viewport',
    eventAction: 'impression',
    eventLabel: id
  };

  var userFields = assign({}, this.opts.fieldsObj,
      getAttributeFields(element, this.opts.attributePrefix));

  this.tracker.send('event', createFieldsObj(defaultFields,
      userFields, this.tracker, this.opts.hitFilter, element));
};


/**
 * Inspects the `items` array after an item was removed. If the removed
 * item's element ID is not found in any other item, the element stops being
 * observed for intersection changes and is removed from `elementMap`.
 * @param {string} id The element ID to check for possible unobservation.
 */
ImpressionTracker.prototype.possiblyUnobserveElement = function(id) {
  if (!this.itemsIncludesId(id)) {
    this.intersectionObserver.unobserve(this.elementMap[id]);
    delete this.elementMap[id];
  }
};


/**
 * Handles an element currently being observed for intersections being removed
 * from the DOM.
 * @param {string} id The ID of the element that was removed.
 */
ImpressionTracker.prototype.handleDomElementRemoved = function(id) {
  this.intersectionObserver.unobserve(this.elementMap[id]);
  this.elementMap[id] = null;
};


/**
 * Scans the `items` array for the presense of an item with the passed ID.
 * @param {string} id The ID of the element to search for.
 * @return {boolean} True if the element ID was found in one of the items.
 */
ImpressionTracker.prototype.itemsIncludesId = function(id) {
  return this.items.some(function(item) {
    return id == item.id;
  });
};


/**
 * Removes all listeners and observers.
 */
ImpressionTracker.prototype.remove = function() {
  this.mutationObserver.disconnect();
  this.intersectionObserver.disconnect();
};


provide('impressionTracker', ImpressionTracker);


/**
 * Detects whether or not an intersection record represents a visible target
 * given a particular threshold.
 * @param {number} threshold The threshold the target is visible above.
 * @param {IntersectionObserverEntry} record The most recent record entry.
 * @return {boolean} True if the target is visible.
 */
function isTargetVisible(threshold, record) {
  if (threshold === 0) {
    var i = record.intersectionRect;
    return i.top > 0 || i.bottom > 0 || i.left > 0 || i.right > 0;
  }
  else {
    return record.intersectionRatio >= threshold;
  }
}
