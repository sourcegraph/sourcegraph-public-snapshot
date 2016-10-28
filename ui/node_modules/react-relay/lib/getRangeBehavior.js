/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule getRangeBehavior
 * 
 */

'use strict';

/**
 * Return the action (prepend/append) to use when adding an item to
 * the range with the specified calls.
 *
 * Ex:
 * rangeBehaviors: `{'orderby(recent)': 'append'}`
 * calls: `[{name: 'orderby', value: 'recent'}]`
 *
 * Returns `'append'`
 */
function getRangeBehavior(rangeBehaviors, calls) {
  if (typeof rangeBehaviors === 'function') {
    var rangeFilterCalls = getObjectFromCalls(calls);
    return rangeBehaviors(rangeFilterCalls);
  } else {
    var rangeBehaviorKey = calls.map(require('./serializeRelayQueryCall')).sort().join('').slice(1);
    var behavior = rangeBehaviors[rangeBehaviorKey];
    if (behavior == null) {
      return null;
    }
    require('fbjs/lib/invariant')(typeof behavior === 'string', 'getRangeBehavior(): Expected range behavior for key `%s` to be a ' + 'string, got `%s`.', rangeBehaviorKey, behavior);
    return behavior;
  }
}

/**
 * Returns an object representation of the rangeFilterCalls that
 * will be passed to config.rangeBehaviors
 *
 * Example:
 * calls: `[{name: 'orderby', value: 'recent'}]`
 *
 * Returns:
 * `{orderby: 'recent'}`
*/
function getObjectFromCalls(calls) {
  var behaviors = {};
  calls.forEach(function (call) {
    behaviors[call.name] = call.value;
  });
  return behaviors;
}

module.exports = getRangeBehavior;