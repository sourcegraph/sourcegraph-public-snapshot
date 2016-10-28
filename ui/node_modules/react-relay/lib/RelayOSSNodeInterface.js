/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayOSSNodeInterface
 * 
 */

'use strict';

/**
 * @internal
 *
 * Defines logic relevant to the informal "Node" GraphQL interface.
 */
var RelayOSSNodeInterface = {
  ANY_TYPE: '__any',
  ID: 'id',
  ID_TYPE: 'ID!',
  NODE: 'node',
  NODE_TYPE: 'Node',
  NODES: 'nodes',
  TYPENAME: '__typename',

  isNodeRootCall: function isNodeRootCall(fieldName) {
    return fieldName === RelayOSSNodeInterface.NODE || fieldName === RelayOSSNodeInterface.NODES;
  },
  getResultsFromPayload: function getResultsFromPayload(query, payload) {
    var results = [];

    var rootBatchCall = query.getBatchCall();
    if (rootBatchCall) {
      getPayloadRecords(query, payload).forEach(function (result) {
        if (typeof result !== 'object' || !result) {
          return;
        }
        var dataID = result[RelayOSSNodeInterface.ID];
        require('fbjs/lib/invariant')(typeof dataID === 'string', 'RelayOSSNodeInterface.getResultsFromPayload(): Unable to write ' + 'result with no `%s` field for query, `%s`.', RelayOSSNodeInterface.ID, query.getName());
        results.push({
          result: result,
          rootCallInfo: {
            storageKey: RelayOSSNodeInterface.NODE,
            identifyingArgKey: dataID,
            identifyingArgValue: dataID
          }
        });
      });
    } else {
      (function () {
        var records = getPayloadRecords(query, payload);
        var ii = 0;
        var storageKey = query.getStorageKey();
        require('./forEachRootCallArg')(query, function (_ref) {
          var identifyingArgKey = _ref.identifyingArgKey;
          var identifyingArgValue = _ref.identifyingArgValue;

          var result = records[ii++];
          results.push({
            result: result,
            rootCallInfo: { storageKey: storageKey, identifyingArgKey: identifyingArgKey, identifyingArgValue: identifyingArgValue }
          });
        });
      })();
    }

    return results;
  }
};

function getPayloadRecords(query, payload) {
  var fieldName = query.getFieldName();
  var identifyingArg = query.getIdentifyingArg();
  var identifyingArgValue = identifyingArg && identifyingArg.value || null;
  var records = payload[fieldName];
  if (!query.getBatchCall()) {
    if (Array.isArray(identifyingArgValue)) {
      require('fbjs/lib/invariant')(Array.isArray(records), 'RelayOSSNodeInterface: Expected payload for root field `%s` to be ' + 'an array with %s results, instead received a single non-array result.', fieldName, identifyingArgValue.length);
      require('fbjs/lib/invariant')(records.length === identifyingArgValue.length, 'RelayOSSNodeInterface: Expected payload for root field `%s` to be ' + 'an array with %s results, instead received an array with %s results.', fieldName, identifyingArgValue.length, records.length);
    } else if (Array.isArray(records)) {
      require('fbjs/lib/invariant')(false, 'RelayOSSNodeInterface: Expected payload for root field `%s` to be ' + 'a single non-array result, instead received an array with %s results.', fieldName, records.length);
    }
  }
  return Array.isArray(records) ? records : [records || null];
}

module.exports = RelayOSSNodeInterface;