/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule callsToGraphQL
 * 
 */

'use strict';

/**
 * @internal
 *
 * Convert from plain object `{name, value}` calls to GraphQL call nodes.
 */
function callsToGraphQL(calls) {
  return calls.map(function (_ref) {
    var name = _ref.name;
    var type = _ref.type;
    var value = _ref.value;

    var concreteValue = null;
    if (Array.isArray(value)) {
      concreteValue = value.map(require('./QueryBuilder').createCallValue);
    } else if (value != null) {
      concreteValue = require('./QueryBuilder').createCallValue(value);
    }
    return require('./QueryBuilder').createCall(name, concreteValue, type);
  });
}

module.exports = callsToGraphQL;