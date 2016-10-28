/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule writeRelayGraphModeResponse
 * 
 */

'use strict';

var _defineProperty3 = _interopRequireDefault(require('babel-runtime/helpers/defineProperty'));

var _extends4 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var ID = require('./RelayConnectionInterface').ID;

var NODE = require('./RelayConnectionInterface').NODE;

var CACHE_KEY = require('./RelayGraphModeInterface').CACHE_KEY;

var DEFERRED_FRAGMENTS = require('./RelayGraphModeInterface').DEFERRED_FRAGMENTS;

var FRAGMENTS = require('./RelayGraphModeInterface').FRAGMENTS;

var REF_KEY = require('./RelayGraphModeInterface').REF_KEY;

var PUT_EDGES = require('./RelayGraphModeInterface').PUT_EDGES;

var PUT_NODES = require('./RelayGraphModeInterface').PUT_NODES;

var PUT_ROOT = require('./RelayGraphModeInterface').PUT_ROOT;

var TYPENAME = require('./RelayNodeInterface').TYPENAME;

var EXISTENT = require('./RelayRecordState').EXISTENT;

var PATH = require('./RelayRecord').MetadataKey.PATH;

/**
 * Writes a GraphMode payload into a Relay store.
 */


function writeRelayGraphModeResponse(store, writer, payload, options) {
  var graphWriter = new RelayGraphModeWriter(store, writer, options);
  graphWriter.write(payload);
  return graphWriter.getChangeTracker();
}

var RelayGraphModeWriter = function () {
  function RelayGraphModeWriter(store, writer, options) {
    (0, _classCallCheck3['default'])(this, RelayGraphModeWriter);

    this._cacheKeyMap = new (require('fbjs/lib/Map'))();
    this._changeTracker = new (require('./RelayChangeTracker'))();
    this._forceIndex = options && options.forceIndex || null;
    this._store = store;
    this._writer = writer;
  }

  RelayGraphModeWriter.prototype.getChangeTracker = function getChangeTracker() {
    return this._changeTracker;
  };

  RelayGraphModeWriter.prototype.write = function write(payload) {
    var _this = this;

    payload.forEach(function (operation) {
      if (operation.op === PUT_ROOT) {
        _this._writeRoot(operation);
      } else if (operation.op === PUT_NODES) {
        _this._writeNodes(operation);
      } else if (operation.op === PUT_EDGES) {
        _this._writeEdges(operation);
      } else {
        require('fbjs/lib/invariant')(false, 'writeRelayGraphModeResponse(): Invalid operation type `%s`, ' + 'expected `root`, `nodes`, or `edges`.', operation.op);
      }
    });
  };

  RelayGraphModeWriter.prototype._writeRoot = function _writeRoot(operation) {
    var field = operation.field;
    var identifier = operation.identifier;
    var root = operation.root;

    var identifyingArgKey = getIdentifyingArgKey(identifier);
    var prevID = this._store.getDataID(field, identifyingArgKey);
    var nextID = void 0;
    if (root != null) {
      nextID = getID(root, prevID);
    } else {
      nextID = prevID || require('./generateClientID')();
    }
    if (root == null) {
      this._writeRecord(nextID, root);
    } else {
      var clientRecord = getGraphRecord(root);
      if (clientRecord) {
        this._writeRecord(nextID, clientRecord);
      }
    }
    this._writer.putDataID(field, identifyingArgKey, nextID);
  };

  RelayGraphModeWriter.prototype._writeNodes = function _writeNodes(operation) {
    var _this2 = this;

    var nodes = operation.nodes;

    require('fbjs/lib/forEachObject')(nodes, function (record, dataID) {
      _this2._writeRecord(dataID, record);
    });
  };

  RelayGraphModeWriter.prototype._writeEdges = function _writeEdges(operation) {
    var _this3 = this;

    var range = operation.range;
    var args = operation.args;
    var edges = operation.edges;
    var pageInfo = operation.pageInfo;

    var rangeID = this._cacheKeyMap.get(range[CACHE_KEY]);
    require('fbjs/lib/invariant')(rangeID, 'writeRelayGraphModeResponse(): Cannot find a record for cache key ' + '`%s`.', range[CACHE_KEY]);
    require('fbjs/lib/invariant')(require('./RelayConnectionInterface').hasRangeCalls(args), 'writeRelayGraphModeResponse(): Cannot write edges for connection on ' + 'record `%s` without `first`, `last`, or `find` argument.', rangeID);
    if (!this._writer.hasRange(rangeID) || this._forceIndex != null && this._forceIndex > this._store.getRangeForceIndex(rangeID)) {
      this._changeTracker.updateID(rangeID);
      this._writer.putRange(rangeID, args, this._forceIndex);
    }
    var rangeInfo = this._store.getRangeMetadata(rangeID, args);
    var filteredEdges = rangeInfo && rangeInfo.filteredEdges || [];
    var fetchedEdgeIDs = [];
    var isUpdate = false;
    var nextIndex = 0;
    edges.forEach(function (edgeData) {
      if (edgeData == null) {
        return;
      }
      var nodeData = edgeData[NODE];
      if (nodeData == null) {
        return;
      }
      require('fbjs/lib/invariant')(typeof nodeData === 'object', 'RelayQueryWriter: Expected node to be an object for `%s`.', edgeData);

      // For consistency, edge IDs are calculated from the connection & node ID.
      // A node ID is only generated if the node does not have an id and
      // there is no existing edge.
      var prevEdge = filteredEdges[nextIndex++];
      var prevNodeID = prevEdge && _this3._store.getLinkedRecordID(prevEdge.edgeID, NODE);
      var nextNodeID = getID(nodeData, prevNodeID);
      var edgeID = require('./generateClientEdgeID')(rangeID, nextNodeID);
      fetchedEdgeIDs.push(edgeID);

      _this3._writeRecord(edgeID, (0, _extends4['default'])({}, edgeData, (0, _defineProperty3['default'])({}, NODE, (0, _defineProperty3['default'])({}, REF_KEY, nextNodeID))));
      var clientRecord = getGraphRecord(nodeData);
      if (clientRecord) {
        _this3._writeRecord(nextNodeID, clientRecord);
      }
      if (nextNodeID !== prevNodeID) {
        _this3._changeTracker.updateID(edgeID);
      }
      isUpdate = isUpdate || !prevEdge || edgeID !== prevEdge.edgeID;
    });

    this._writer.putRangeEdges(rangeID, args, pageInfo || require('./RelayConnectionInterface').getDefaultPageInfo(), fetchedEdgeIDs);

    if (isUpdate) {
      this._changeTracker.updateID(rangeID);
    }
  };

  RelayGraphModeWriter.prototype._writeRecord = function _writeRecord(dataID, record) {
    var _this4 = this;

    var recordState = this._store.getRecordState(dataID);
    if (record === undefined) {
      return;
    } else if (record === null) {
      if (recordState === EXISTENT) {
        this._changeTracker.updateID(dataID);
      }
      this._writer.deleteRecord(dataID);
      return;
    }
    var cacheKey = getCacheKey(record);
    if (cacheKey) {
      this._cacheKeyMap.set(cacheKey, dataID);
    }
    if (recordState !== EXISTENT) {
      this._changeTracker.createID(dataID);
    }
    var path = record[PATH] || null;
    var typeName = record[TYPENAME] || null;
    // TODO #10481948: Construct paths lazily
    this._writer.putRecord(dataID, typeName, path);

    require('fbjs/lib/forEachObject')(record, function (nextValue, storageKey) {
      if (storageKey === CACHE_KEY || storageKey === PATH || storageKey === REF_KEY) {
        return;
      } else if (storageKey === FRAGMENTS) {
        _this4._writeFragments(dataID, nextValue, false);
      } else if (storageKey === DEFERRED_FRAGMENTS) {
        _this4._writeFragments(dataID, nextValue, true);
      } else if (nextValue === undefined) {
        return;
      } else if (nextValue === null) {
        _this4._writeScalar(dataID, storageKey, nextValue);
      } else if (Array.isArray(nextValue)) {
        _this4._writePlural(dataID, storageKey, nextValue);
      } else if (typeof nextValue === 'object') {
        _this4._writeLinkedRecord(dataID, storageKey, nextValue);
      } else {
        _this4._writeScalar(dataID, storageKey, nextValue);
      }
    });
  };

  RelayGraphModeWriter.prototype._writeFragments = function _writeFragments(dataID, fragments, isDeferred) {
    var _this5 = this;

    if (isDeferred) {
      // Changes are recorded for deferred fragments to ensure that parent
      // components re-render with the new data.
      this._changeTracker.updateID(dataID);
      require('fbjs/lib/forEachObject')(fragments, function (_, fragmentHash) {
        _this5._writer.setHasDeferredFragmentData(dataID, fragmentHash);
      });
    } else {
      // Other fragments are for diff optimization only and do not require a
      // re-render.
      require('fbjs/lib/forEachObject')(fragments, function (_, fragmentHash) {
        _this5._writer.setHasFragmentData(dataID, fragmentHash);
      });
    }
  };

  RelayGraphModeWriter.prototype._writeScalar = function _writeScalar(dataID, storageKey, nextValue) {
    var prevValue = this._store.getField(dataID, storageKey);
    if (prevValue !== nextValue) {
      this._changeTracker.updateID(dataID);
    }
    this._writer.putField(dataID, storageKey, nextValue);
  };

  RelayGraphModeWriter.prototype._writePlural = function _writePlural(dataID, storageKey, nextValue) {
    var _this6 = this;

    var prevValue = this._store.getField(dataID, storageKey);
    var prevArray = Array.isArray(prevValue) ? prevValue : null;
    var nextIDs = null;
    var nextScalars = null;
    var isUpdate = false;
    var nextIndex = 0;
    nextValue.forEach(function (nextItem) {
      if (nextItem == null) {
        return;
      } else if (typeof nextItem === 'object') {
        require('fbjs/lib/invariant')(!nextScalars, 'writeRelayGraphModeResponse(): Expected items for field `%s` to ' + 'all be objects or all be scalars, got both.', storageKey);
        var prevItem = prevArray && prevArray[nextIndex++];
        var prevID = typeof prevItem === 'object' && prevItem != null ? require('./RelayRecord').getDataIDForObject(prevItem) : null;
        var nextID = getID(nextItem, prevID);
        var clientRecord = getGraphRecord(nextItem);
        if (clientRecord) {
          _this6._writeRecord(nextID, clientRecord);
        }
        isUpdate = isUpdate || nextID !== prevID;
        nextIDs = nextIDs || [];
        nextIDs.push(nextID);
      } else {
        // array of scalars
        require('fbjs/lib/invariant')(!nextIDs, 'writeRelayGraphModeResponse(): Expected items for field `%s` to ' + 'all be objects or all be scalars, got both.', storageKey);
        var _prevItem = prevArray && prevArray[nextIndex++];
        isUpdate = isUpdate || nextItem !== _prevItem;
        nextScalars = nextScalars || [];
        nextScalars.push(nextItem);
      }
    });
    nextScalars = nextScalars || [];
    var nextArray = nextIDs || nextScalars;
    if (isUpdate || !prevArray || nextArray.length !== prevArray.length) {
      this._changeTracker.updateID(dataID);
    }
    if (nextIDs) {
      this._writer.putLinkedRecordIDs(dataID, storageKey, nextIDs);
    } else {
      this._writer.putField(dataID, storageKey, nextScalars || []);
    }
  };

  RelayGraphModeWriter.prototype._writeLinkedRecord = function _writeLinkedRecord(dataID, storageKey, nextValue) {
    var prevID = this._store.getLinkedRecordID(dataID, storageKey);
    var nextID = getID(nextValue, prevID);

    var clientRecord = getGraphRecord(nextValue);
    if (clientRecord) {
      this._writeRecord(nextID, clientRecord);
    }
    if (nextID !== prevID) {
      this._changeTracker.updateID(dataID);
    }
    this._writer.putLinkedRecordID(dataID, storageKey, nextID);
  };

  return RelayGraphModeWriter;
}();

function getCacheKey(record) {
  if (record.hasOwnProperty(CACHE_KEY) && typeof record[CACHE_KEY] === 'string') {
    return record[CACHE_KEY];
  }
  return null;
}

function getID(record, prevID) {
  if (record.hasOwnProperty(REF_KEY) && typeof record[REF_KEY] === 'string') {
    return record[REF_KEY];
  } else if (record.hasOwnProperty(ID) && typeof record[ID] === 'string') {
    return record[ID];
  } else if (prevID != null) {
    return prevID;
  } else {
    return require('./generateClientID')();
  }
}

function getIdentifyingArgKey(value) {
  if (value == null) {
    return null;
  } else {
    return typeof value === 'string' ? value : require('./stableStringify')(value);
  }
}

function getGraphRecord(record) {
  if (!record.hasOwnProperty(REF_KEY)) {
    return record;
  }
  return null;
}

module.exports = writeRelayGraphModeResponse;