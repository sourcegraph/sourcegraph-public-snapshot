/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayOptimisticMutationUtils
 * @typechecks
 * 
 */

'use strict';

var _defineProperty3 = _interopRequireDefault(require('babel-runtime/helpers/defineProperty'));

var _extends5 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var ARGUMENTS = /^(\w+)(?:\((.+?)\))?$/;
var ARGUMENT_NAME = /(\w+)(?=\s*:)/;
var DEPRECATED_CALLS = /^\w+(?:\.\w+\(.*?\))+$/;
var DEPRECATED_CALL = /^(\w+)\((.*?)\)$/;

var NODE = require('./RelayConnectionInterface').NODE;

var EDGES = require('./RelayConnectionInterface').EDGES;

var ANY_TYPE = require('./RelayNodeInterface').ANY_TYPE;

var ID = require('./RelayNodeInterface').ID;

var idField = require('./RelayQuery').Field.build({
  fieldName: ID,
  type: 'String'
});
var cursorField = require('./RelayQuery').Field.build({
  fieldName: 'cursor',
  type: 'String'
});

/**
 * @internal
 */
var RelayOptimisticMutationUtils = {
  /**
   * Given a record-like object, infers fields that could be used to fetch them.
   * Properties that are fetched via fields with arguments can be encoded by
   * serializing the arguments in property keys.
   */
  inferRelayFieldsFromData: function inferRelayFieldsFromData(data) {
    var fields = [];
    require('fbjs/lib/forEachObject')(data, function (value, key) {
      if (!require('./RelayRecord').isMetadataKey(key)) {
        fields.push(inferField(value, key));
      }
    });
    return fields;
  },
  /**
   * Given a record-like object, infer the proper payload to be used to store
   * them. Properties that are fetched via fields with arguments will be
   * encoded by serializing the arguments in property keys.
   */
  inferRelayPayloadFromData: function inferRelayPayloadFromData(data) {
    var payload = data;
    require('fbjs/lib/forEachObject')(data, function (value, key) {
      if (!require('./RelayRecord').isMetadataKey(key)) {
        var _inferPayload = inferPayload(value, key);

        var newValue = _inferPayload.newValue;
        var _newKey = _inferPayload.newKey;

        if (_newKey !== key) {
          payload = (0, _extends5['default'])({}, payload, (0, _defineProperty3['default'])({}, _newKey, newValue));
          delete payload[key];
        } else if (newValue !== value) {
          payload = (0, _extends5['default'])({}, payload, (0, _defineProperty3['default'])({}, key, newValue));
        }
      }
    });
    return payload;
  }
};

function inferField(value, key) {
  var metadata = {
    canHaveSubselections: true,
    isPlural: false
  };
  var children = void 0;
  if (Array.isArray(value)) {
    var element = value[0];
    if (element && typeof element === 'object') {
      children = RelayOptimisticMutationUtils.inferRelayFieldsFromData(element);
    } else {
      metadata.canHaveSubselections = false;
      children = [];
    }
    metadata.isPlural = true;
  } else if (typeof value === 'object' && value !== null) {
    children = RelayOptimisticMutationUtils.inferRelayFieldsFromData(value);
  } else {
    metadata.canHaveSubselections = false;
    children = [];
  }
  if (key === NODE) {
    children.push(idField);
  } else if (key === EDGES) {
    children.push(cursorField);
  }
  return buildField(key, children, metadata);
}

function inferPayload(value, key) {
  var metadata = {
    canHaveSubselections: true,
    isPlural: false
  };
  var newValue = value;
  if (Array.isArray(value) && Array.isArray(newValue)) {
    for (var ii = 0; ii < value.length; ii++) {
      var element = value[ii];
      if (element && typeof element === 'object') {
        var newElement = RelayOptimisticMutationUtils.inferRelayPayloadFromData(element);
        if (newElement !== element) {
          newValue = newValue.slice();
          newValue[ii] = newElement;
        }
      } else {
        metadata.canHaveSubselections = false;
      }
    }
    metadata.isPlural = true;
  } else if (typeof value === 'object' && value !== null) {
    newValue = RelayOptimisticMutationUtils.inferRelayPayloadFromData(value);
  } else {
    metadata.canHaveSubselections = false;
  }

  var field = buildField(key, [], metadata);
  return { newValue: newValue, newKey: field.getSerializationKey() };
}

function buildField(key, children, metadata) {
  var fieldName = key;
  var calls = null;
  if (DEPRECATED_CALLS.test(key)) {
    require('fbjs/lib/warning')(false, 'RelayOptimisticMutationUtils: Encountered an optimistic payload with ' + 'a deprecated field call string, `%s`. Use valid GraphQL OSS syntax.', key);
    var parts = key.split('.');
    if (parts.length > 1) {
      fieldName = parts.shift();
      calls = parts.map(function (callString) {
        var captures = callString.match(DEPRECATED_CALL);
        require('fbjs/lib/invariant')(captures, 'RelayOptimisticMutationUtils: Malformed data key, `%s`.', key);
        var value = captures[2].split(',');
        return {
          name: captures[1],
          value: value.length === 1 ? value[0] : value
        };
      });
    }
  } else {
    var captures = key.match(ARGUMENTS);
    require('fbjs/lib/invariant')(captures, 'RelayOptimisticMutationUtils: Malformed data key, `%s`.', key);
    fieldName = captures[1];
    if (captures[2]) {
      try {
        (function () {
          // Relay does not currently have a GraphQL argument parser, so...
          var args = JSON.parse('{' + captures[2].replace(ARGUMENT_NAME, '"$1"') + '}');
          calls = (0, _keys2['default'])(args).map(function (name) {
            return { name: name, value: args[name] };
          });
        })();
      } catch (error) {
        require('fbjs/lib/invariant')(false, 'RelayOptimisticMutationUtils: Malformed or unsupported data key, ' + '`%s`. Only booleans, strings, and numbers are currently supported, ' + 'and commas are required. Parse failure reason was `%s`.', key, error.message);
      }
    }
  }
  return require('./RelayQuery').Field.build({
    calls: calls,
    children: children,
    fieldName: fieldName,
    metadata: metadata,
    type: ANY_TYPE
  });
}

module.exports = RelayOptimisticMutationUtils;