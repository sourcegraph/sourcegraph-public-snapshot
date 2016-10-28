/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule writeRelayQueryPayload
 * 
 */

'use strict';

var ID = require('./RelayNodeInterface').ID;

/**
 * @internal
 *
 * Traverses a query and payload in parallel, writing the results into the
 * store.
 */


function writeRelayQueryPayload(writer, query, payload) {
  var store = writer.getRecordStore();
  var recordWriter = writer.getRecordWriter();
  var path = require('./RelayQueryPath').create(query);

  require('./RelayNodeInterface').getResultsFromPayload(query, payload).forEach(function (_ref) {
    var result = _ref.result;
    var rootCallInfo = _ref.rootCallInfo;
    var storageKey = rootCallInfo.storageKey;
    var identifyingArgKey = rootCallInfo.identifyingArgKey;


    var dataID = void 0;
    if (typeof result === 'object' && result && typeof result[ID] === 'string') {
      dataID = result[ID];
    }

    if (dataID == null) {
      dataID = store.getDataID(storageKey, identifyingArgKey) || require('./generateClientID')();
    }

    recordWriter.putDataID(storageKey, identifyingArgKey, dataID);
    writer.writePayload(query, dataID, result, path);
  });
}

module.exports = require('./RelayProfiler').instrument('writeRelayQueryPayload', writeRelayQueryPayload);