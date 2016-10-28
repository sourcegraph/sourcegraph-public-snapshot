/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule filterExclusiveKeys
 * 
 */

'use strict';

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var hasOwnProperty = Object.prototype.hasOwnProperty;

/**
 * Returns two arrays of keys that contain each object's exclusive keys.
 */
function filterExclusiveKeys(a, b) {
  var keysA = a ? (0, _keys2['default'])(a) : [];
  var keysB = b ? (0, _keys2['default'])(b) : [];

  if (keysA.length === 0 || keysB.length === 0) {
    return [keysA, keysB];
  }
  return [keysA.filter(function (key) {
    return !hasOwnProperty.call(b, key);
  }), keysB.filter(function (key) {
    return !hasOwnProperty.call(a, key);
  })];
}

module.exports = filterExclusiveKeys;