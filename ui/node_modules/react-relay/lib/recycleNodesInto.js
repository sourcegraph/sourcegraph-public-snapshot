/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule recycleNodesInto
 * 
 */

'use strict';

/**
 * Recycles subtrees from `prevData` by replacing equal subtrees in `nextData`.
 */

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

function recycleNodesInto(prevData, nextData) {
  if (prevData === nextData || typeof prevData !== 'object' || !prevData || typeof nextData !== 'object' || !nextData) {
    return nextData;
  }
  var canRecycle = false;

  // Assign local variables to preserve Flow type refinement.
  var prevArray = Array.isArray(prevData) ? prevData : null;
  var nextArray = Array.isArray(nextData) ? nextData : null;
  if (prevArray && nextArray) {
    canRecycle = nextArray.reduce(function (wasEqual, nextItem, ii) {
      nextArray[ii] = recycleNodesInto(prevArray[ii], nextItem);
      return wasEqual && nextArray[ii] === prevArray[ii];
    }, true) && prevArray.length === nextArray.length;
  } else if (!prevArray && !nextArray) {
    (function () {
      // Assign local variables to preserve Flow type refinement.
      var prevObject = prevData;
      var nextObject = nextData;
      var prevKeys = (0, _keys2['default'])(prevObject);
      var nextKeys = (0, _keys2['default'])(nextObject);
      canRecycle = nextKeys.reduce(function (wasEqual, key) {
        var nextValue = nextObject[key];
        nextObject[key] = recycleNodesInto(prevObject[key], nextValue);
        return wasEqual && nextObject[key] === prevObject[key];
      }, true) && prevKeys.length === nextKeys.length;
    })();
  }
  return canRecycle ? prevData : nextData;
}

module.exports = recycleNodesInto;