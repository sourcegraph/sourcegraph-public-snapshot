/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule printRelayOSSQuery
 * 
 */

'use strict';

var _stringify2 = _interopRequireDefault(require('babel-runtime/core-js/json/stringify'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var oneIndent = '';
var newLine = '';

if (process.env.NODE_ENV !== 'production') {
  oneIndent = '  ';
  newLine = '\n';
}

/**
 * @internal
 *
 * `printRelayOSSQuery(query)` returns a string representation of the query. The
 * supplied `node` must be flattened (and not contain fragments).
 */
function printRelayOSSQuery(node) {
  var fragmentTexts = [];
  var variableMap = new (require('fbjs/lib/Map'))();
  var printerState = {
    fragmentCount: 0,
    fragmentNameByHash: {},
    fragmentNameByText: {},
    fragmentTexts: fragmentTexts,
    variableCount: 0,
    variableMap: variableMap
  };
  var queryText = null;
  if (node instanceof require('./RelayQuery').Root) {
    queryText = printRoot(node, printerState);
  } else if (node instanceof require('./RelayQuery').Operation) {
    queryText = printOperation(node, printerState);
  } else if (node instanceof require('./RelayQuery').Fragment) {
    queryText = printFragment(node, printerState);
  }
  require('fbjs/lib/invariant')(queryText, 'printRelayOSSQuery(): Unsupported node type.');
  var variables = {};
  variableMap.forEach(function (variablesForType) {
    variablesForType.forEach(function (_ref) {
      var value = _ref.value;
      var variableID = _ref.variableID;

      variables[variableID] = value;
    });
  });

  return {
    text: [queryText].concat(fragmentTexts).join(newLine.length ? newLine : ' '),
    variables: variables
  };
}

function printRoot(node, printerState) {
  require('fbjs/lib/invariant')(!node.getBatchCall(), 'printRelayOSSQuery(): Deferred queries are not supported.');
  var identifyingArg = node.getIdentifyingArg();
  var identifyingArgName = identifyingArg && identifyingArg.name || null;
  var identifyingArgType = identifyingArg && identifyingArg.type || null;
  var identifyingArgValue = identifyingArg && identifyingArg.value || null;
  var fieldName = node.getFieldName();
  if (identifyingArgValue != null) {
    require('fbjs/lib/invariant')(identifyingArgName, 'printRelayOSSQuery(): Expected an argument name for root field `%s`.', fieldName);
    var rootArgString = printArgument(identifyingArgName, identifyingArgValue, identifyingArgType, printerState);
    if (rootArgString) {
      fieldName += '(' + rootArgString + ')';
    }
  }
  // Note: children must be traversed before printing variable definitions
  var children = printChildren(node, printerState, oneIndent);
  var queryString = node.getName() + printVariableDefinitions(printerState);
  fieldName += printDirectives(node);

  return 'query ' + queryString + ' {' + newLine + oneIndent + fieldName + children + newLine + '}';
}

function printOperation(node, printerState) {
  var operationKind = node instanceof require('./RelayQuery').Mutation ? 'mutation' : 'subscription';
  var call = node.getCall();
  var inputString = printArgument(node.getCallVariableName(), call.value, node.getInputType(), printerState);
  require('fbjs/lib/invariant')(inputString, 'printRelayOSSQuery(): Expected %s `%s` to have a value for `%s`.', operationKind, node.getName(), node.getCallVariableName());
  // Note: children must be traversed before printing variable definitions
  var children = printChildren(node, printerState, oneIndent);
  var operationString = node.getName() + printVariableDefinitions(printerState);
  var fieldName = call.name + '(' + inputString + ')';

  return operationKind + ' ' + operationString + ' {' + newLine + oneIndent + fieldName + children + newLine + '}';
}

function printVariableDefinitions(_ref2) {
  var variableMap = _ref2.variableMap;

  var argStrings = null;
  variableMap.forEach(function (variablesForType, type) {
    variablesForType.forEach(function (_ref3) {
      var variableID = _ref3.variableID;

      argStrings = argStrings || [];
      argStrings.push('$' + variableID + ':' + type);
    });
  });
  if (argStrings) {
    return '(' + argStrings.join(',') + ')';
  }
  return '';
}

function printNonNullType(type) {
  if (type.endsWith('!')) {
    return type;
  }
  return type + '!';
}

function printFragment(node, printerState) {
  var directives = printDirectives(node);
  return 'fragment ' + node.getDebugName() + ' on ' + node.getType() + directives + printChildren(node, printerState, '');
}

function printChildren(node, printerState, indent) {
  var childrenText = [];
  var children = node.getChildren();
  var fragments = void 0;
  for (var ii = 0; ii < children.length; ii++) {
    var child = children[ii];
    if (child instanceof require('./RelayQuery').Field) {
      var fieldText = child.getSchemaName();
      var fieldCalls = child.getCallsWithValues();
      if (fieldCalls.length) {
        fieldText = child.getSerializationKey() + ':' + fieldText;
        var argTexts = [];
        for (var jj = 0; jj < fieldCalls.length; jj++) {
          var _fieldCalls$jj = fieldCalls[jj];
          var name = _fieldCalls$jj.name;
          var _value = _fieldCalls$jj.value;

          var argText = printArgument(name, _value, child.getCallType(name), printerState);
          if (argText) {
            argTexts.push(argText);
          }
        }
        if (argTexts.length) {
          fieldText += '(' + argTexts.join(',') + ')';
        }
      }
      fieldText += printDirectives(child);
      if (child.getChildren().length) {
        fieldText += printChildren(child, printerState, indent + oneIndent);
      }
      childrenText.push(fieldText);
    } else if (child instanceof require('./RelayQuery').Fragment) {
      if (child.getChildren().length) {
        var _fragmentNameByHash = printerState.fragmentNameByHash;
        var _fragmentNameByText = printerState.fragmentNameByText;
        var _fragmentTexts = printerState.fragmentTexts;

        // Avoid walking fragments if we have printed the same one before.

        var _fragmentHash = child.getCompositeHash();

        var fragmentName = void 0;
        if (_fragmentNameByHash.hasOwnProperty(_fragmentHash)) {
          fragmentName = _fragmentNameByHash[_fragmentHash];
        } else {
          // Avoid reprinting a fragment that is identical to another fragment.
          var _fragmentText = child.getType() + printDirectives(child) + printChildren(child, printerState, '');
          if (_fragmentNameByText.hasOwnProperty(_fragmentText)) {
            fragmentName = _fragmentNameByText[_fragmentText];
          } else {
            fragmentName = 'F' + require('fbjs/lib/base62')(printerState.fragmentCount++);
            _fragmentNameByHash[_fragmentHash] = fragmentName;
            _fragmentNameByText[_fragmentText] = fragmentName;
            _fragmentTexts.push('fragment ' + fragmentName + ' on ' + _fragmentText);
          }
        }
        if (!fragments || !fragments.hasOwnProperty(fragmentName)) {
          fragments = fragments || {};
          fragments[fragmentName] = true;
          childrenText.push('...' + fragmentName);
        }
      }
    } else {
      require('fbjs/lib/invariant')(false, 'printRelayOSSQuery(): Expected a field or fragment, got `%s`.', child.constructor.name);
    }
  }
  if (!childrenText) {
    return '';
  }
  return childrenText.length ? ' {' + newLine + indent + oneIndent + childrenText.join(',' + newLine + indent + oneIndent) + newLine + indent + '}' : '';
}

function printDirectives(node) {
  var directiveStrings = void 0;
  node.getDirectives().forEach(function (directive) {
    var dirString = '@' + directive.name;
    if (directive.args.length) {
      dirString += '(' + directive.args.map(printDirective).join(',') + ')';
    }
    directiveStrings = directiveStrings || [];
    directiveStrings.push(dirString);
  });
  if (!directiveStrings) {
    return '';
  }
  return ' ' + directiveStrings.join(' ');
}

function printDirective(_ref4) {
  var name = _ref4.name;
  var value = _ref4.value;

  require('fbjs/lib/invariant')(typeof value === 'boolean' || typeof value === 'number' || typeof value === 'string', 'printRelayOSSQuery(): Relay only supports directives with scalar values ' + '(boolean, number, or string), got `%s: %s`.', name, value);
  return name + ':' + (0, _stringify2['default'])(value);
}

function printArgument(name, value, type, printerState) {
  if (value == null) {
    return value;
  }
  var stringValue = void 0;
  if (type != null) {
    var _variableID = createVariable(name, value, type, printerState);
    stringValue = '$' + _variableID;
  } else {
    stringValue = (0, _stringify2['default'])(value);
  }
  return name + ':' + stringValue;
}

function createVariable(name, value, type, printerState) {
  require('fbjs/lib/invariant')(value != null, 'printRelayOSSQuery: Expected a non-null value for variable `%s`.', name);
  var valueKey = (0, _stringify2['default'])(value);
  var nonNullType = printNonNullType(type);
  var variablesForType = printerState.variableMap.get(nonNullType);
  if (!variablesForType) {
    variablesForType = new (require('fbjs/lib/Map'))();
    printerState.variableMap.set(nonNullType, variablesForType);
  }
  var existingVariable = variablesForType.get(valueKey);
  if (existingVariable) {
    return existingVariable.variableID;
  } else {
    var _variableID2 = name + '_' + require('fbjs/lib/base62')(printerState.variableCount++);
    variablesForType.set(valueKey, {
      value: value,
      variableID: _variableID2
    });
    return _variableID2;
  }
}

module.exports = require('./RelayProfiler').instrument('printRelayQuery', printRelayOSSQuery);