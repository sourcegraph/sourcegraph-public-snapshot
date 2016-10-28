/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule callsFromGraphQL
 * 
 */

'use strict';

var _stringify2 = _interopRequireDefault(require('babel-runtime/core-js/json/stringify'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @internal
 *
 * Convert from GraphQL call nodes to plain object `{name,value}` calls.
 */
function callsFromGraphQL(concreteCalls, variables) {
  // $FlowIssue: ConcreteCall should flow into CallOrDirective
  var callsOrDirectives = concreteCalls;
  var orderedCalls = [];
  for (var ii = 0; ii < callsOrDirectives.length; ii++) {
    var callOrDirective = callsOrDirectives[ii];
    var _value = callOrDirective.value;

    if (_value != null) {
      if (Array.isArray(_value)) {
        _value = _value.map(function (arg) {
          return getCallValue(arg, variables);
        });
      } else if (_value.kind === 'BatchCallVariable') {
        // Batch calls are handled separately
        _value = null;
      } else {
        _value = getCallValue(_value, variables);
      }
    }
    var _metadata = callOrDirective.metadata;
    var _name = callOrDirective.name;

    var orderedCall = { name: _name, value: _value };
    if (_metadata && _metadata.type) {
      orderedCall.type = _metadata.type;
    }
    orderedCalls.push(orderedCall);
  }
  return orderedCalls;
}

function getCallValue(concreteValue, variables) {
  var callValue = void 0;
  if (concreteValue.kind === 'CallValue') {
    callValue = concreteValue.callValue;
  } else {
    var variableName = concreteValue.callVariableName;
    require('fbjs/lib/invariant')(variables.hasOwnProperty(variableName), 'callsFromGraphQL(): Expected a declared value for variable, `$%s`.', variableName);
    callValue = variables[variableName];
  }
  // Perform a shallow check to ensure the value conforms to `CallValue` type:
  // For performance reasons, skip recursively testing array/object values.
  var valueType = typeof callValue;
  require('fbjs/lib/invariant')(callValue == null || valueType === 'boolean' || valueType === 'number' || valueType === 'string' || valueType === 'object', 'callsFromGraphQL(): Expected argument value `%s` to either be null or a ' + 'boolean, number, string, or array/object.', (0, _stringify2['default'])(callValue));
  return callValue;
}

module.exports = callsFromGraphQL;