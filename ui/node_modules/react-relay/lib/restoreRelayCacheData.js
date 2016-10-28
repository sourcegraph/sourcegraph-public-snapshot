/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule restoreRelayCacheData
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @internal
 *
 * Retrieves data for queries or fragments from disk into `cachedRecords`.
 */
function restoreFragmentDataFromCache(dataID, fragment, path, store, cachedRecords, cachedRootCallMap, garbageCollector, cacheManager, changeTracker, callbacks) {
  var restorator = new RelayCachedDataRestorator(cacheManager, store, cachedRecords, cachedRootCallMap, changeTracker, callbacks, garbageCollector);
  restorator.restoreFragmentData(dataID, fragment, path);

  return {
    abort: function abort() {
      restorator.abort();
    }
  };
}

function restoreQueriesDataFromCache(queries, store, cachedRecords, cachedRootCallMap, garbageCollector, cacheManager, changeTracker, callbacks) {
  var restorator = new RelayCachedDataRestorator(cacheManager, store, cachedRecords, cachedRootCallMap, changeTracker, callbacks, garbageCollector);
  restorator.restoreQueriesData(queries);

  return {
    abort: function abort() {
      restorator.abort();
    }
  };
}

var RelayCachedDataRestorator = function (_RelayCacheProcessor) {
  (0, _inherits3['default'])(RelayCachedDataRestorator, _RelayCacheProcessor);

  function RelayCachedDataRestorator(cacheManager, store, cachedRecords, cachedRootCallMap, changeTracker, callbacks, garbageCollector) {
    (0, _classCallCheck3['default'])(this, RelayCachedDataRestorator);

    var _this = (0, _possibleConstructorReturn3['default'])(this, _RelayCacheProcessor.call(this, cacheManager, callbacks));

    _this._cachedRecords = cachedRecords;
    _this._cachedRootCallMap = cachedRootCallMap;
    _this._changeTracker = changeTracker;
    _this._garbageCollector = garbageCollector;
    _this._store = store;
    return _this;
  }

  RelayCachedDataRestorator.prototype.handleNodeVisited = function handleNodeVisited(node, dataID, record, nextState) {
    var recordState = this._store.getRecordState(dataID);
    this._cachedRecords[dataID] = record;
    // Mark records as created/updated as necessary. Note that if the
    // record is known to be deleted in the store then it will have been
    // been marked as created already. Further, it does not need to be
    // updated since no additional data can be read about a deleted node.
    if (recordState === 'UNKNOWN' && record !== undefined) {
      // Register immediately in case anything tries to read and subscribe
      // to this record (which means incrementing reference counts).
      if (this._garbageCollector) {
        this._garbageCollector.register(dataID);
      }
      // Mark as created if the store did not have a record but disk cache
      // did (either a known record or known deletion).
      this._changeTracker.createID(dataID);
    } else if (recordState === 'EXISTENT' && record != null) {
      // Mark as updated only if a record exists in both the store and
      // disk cache.
      this._changeTracker.updateID(dataID);
    }
    if (!record) {
      // We are out of luck if disk doesn't have the node either.
      this.handleFailure();
      return;
    }
    if (require('./RelayRecord').isClientID(dataID)) {
      record.__path__ = nextState.path;
    }
  };

  RelayCachedDataRestorator.prototype.handleIdentifiedRootVisited = function handleIdentifiedRootVisited(query, dataID, identifyingArgKey, nextState) {
    if (dataID == null) {
      // Read from cache and we still don't have a valid `dataID`.
      this.handleFailure();
      return;
    }
    var storageKey = query.getStorageKey();
    this._cachedRootCallMap[storageKey] = this._cachedRootCallMap[storageKey] || {};
    this._cachedRootCallMap[storageKey][identifyingArgKey || ''] = dataID;
    nextState.dataID = dataID;
  };

  RelayCachedDataRestorator.prototype.restoreFragmentData = function restoreFragmentData(dataID, fragment, path) {
    var _this2 = this;

    this.process(function () {
      _this2.visitFragment(fragment, {
        dataID: dataID,
        node: fragment,
        path: path,
        rangeCalls: undefined
      });
    });
  };

  RelayCachedDataRestorator.prototype.restoreQueriesData = function restoreQueriesData(queries) {
    var _this3 = this;

    this.process(function () {
      require('fbjs/lib/forEachObject')(queries, function (query) {
        if (_this3._state === 'COMPLETED') {
          return;
        }
        if (query) {
          _this3.visitRoot(query, {
            dataID: undefined,
            node: query,
            path: require('./RelayQueryPath').create(query),
            rangeCalls: undefined
          });
        }
      });
    });
  };

  RelayCachedDataRestorator.prototype.traverse = function traverse(node, nextState) {
    require('fbjs/lib/invariant')(nextState.dataID != null, 'RelayCachedDataRestorator: Attempted to traverse without a ' + '`dataID`.');

    var _findRelayQueryLeaves = require('./findRelayQueryLeaves')(this._store, this._cachedRecords, nextState.node, nextState.dataID, nextState.path, nextState.rangeCalls);

    var missingData = _findRelayQueryLeaves.missingData;
    var pendingNodeStates = _findRelayQueryLeaves.pendingNodeStates;

    if (missingData) {
      this.handleFailure();
      return;
    }
    for (var ii = 0; ii < pendingNodeStates.length; ii++) {
      if (this._state === 'COMPLETED') {
        return;
      }
      require('fbjs/lib/invariant')(pendingNodeStates[ii].dataID != null, 'RelayCachedDataRestorator: Attempted to visit a node without ' + 'a `dataID`.');
      this.visitNode(pendingNodeStates[ii].node, pendingNodeStates[ii].dataID, pendingNodeStates[ii]);
    }
  };

  RelayCachedDataRestorator.prototype.visitIdentifiedRoot = function visitIdentifiedRoot(query, identifyingArgKey, nextState) {
    var dataID = this._store.getDataID(query.getStorageKey(), identifyingArgKey);
    if (dataID == null) {
      _RelayCacheProcessor.prototype.visitIdentifiedRoot.call(this, query, identifyingArgKey, nextState);
    } else {
      this.traverse(query, {
        dataID: dataID,
        node: query,
        path: require('./RelayQueryPath').create(query),
        rangeCalls: undefined
      });
    }
  };

  return RelayCachedDataRestorator;
}(require('./RelayCacheProcessor'));

require('./RelayProfiler').instrumentMethods(RelayCachedDataRestorator.prototype, {
  handleIdentifiedRootVisited: 'RelayCachedDataRestorator.handleIdentifiedRootVisited',
  handleNodeVisited: 'RelayCachedDataRestorator.handleNodeVisited',
  queueIdentifiedRoot: 'RelayCachedDataRestorator.queueRoot',
  queueNode: 'RelayCachedDataRestorator.queueNode',
  restoreFragmentData: 'RelayCachedDataRestorator.readFragment',
  restoreQueriesData: 'RelayCachedDataRestorator.read',
  traverse: 'RelayCachedDataRestorator.traverse',
  visitNode: 'RelayCachedDataRestorator.visitNode',
  visitRoot: 'RelayCachedDataRestorator.visitRoot'
});

module.exports = {
  restoreFragmentDataFromCache: restoreFragmentDataFromCache,
  restoreQueriesDataFromCache: restoreQueriesDataFromCache
};