/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule serializeRelayQueryCall
 * 
 */

'use strict';

var _stringify2 = _interopRequireDefault(require('babel-runtime/core-js/json/stringify'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @internal
 *
 * Serializes a query "call" (a legacy combination of field and argument value).
 */
function serializeRelayQueryCall(call) {
  var value = call.value;

  var valueString = void 0;
  if (Array.isArray(value)) {
    valueString = require('fbjs/lib/flattenArray')(value).map(serializeCallValue).join(',');
  } else {
    valueString = serializeCallValue(value);
  }
  return '.' + call.name + '(' + valueString + ')';
}

function serializeCallValue(value) {
  if (value == null) {
    return '';
  } else if (typeof value !== 'string') {
    return (0, _stringify2['default'])(value);
  } else {
    return value;
  }
}

module.exports = serializeRelayQueryCall;