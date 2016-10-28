/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayRecordStore
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var EMPTY = '';

var NODE = require('./RelayConnectionInterface').NODE;

var _RelayRecord$Metadata = require('./RelayRecord').MetadataKey;

var FILTER_CALLS = _RelayRecord$Metadata.FILTER_CALLS;
var FORCE_INDEX = _RelayRecord$Metadata.FORCE_INDEX;
var MUTATION_IDS = _RelayRecord$Metadata.MUTATION_IDS;
var PATH = _RelayRecord$Metadata.PATH;
var RANGE = _RelayRecord$Metadata.RANGE;
var RESOLVED_FRAGMENT_MAP = _RelayRecord$Metadata.RESOLVED_FRAGMENT_MAP;

/**
 * @internal
 *
 * `RelayRecordStore` is the central repository for all data fetched by the
 * client. Data is stored as a map from IDs to Records. Records are shapes of
 * data with globally unique "data IDs".
 */

var RelayRecordStore = function () {
  function RelayRecordStore(records, rootCallMaps, nodeConnectionMap) {
    (0, _classCallCheck3['default'])(this, RelayRecordStore);

    this._cachedRecords = records.cachedRecords;
    this._cachedRootCallMap = rootCallMaps && rootCallMaps.cachedRootCallMap || {};
    this._queuedRecords = records.queuedRecords;
    this._nodeConnectionMap = nodeConnectionMap || {};
    this._records = records.records;
    this._rootCallMap = rootCallMaps && rootCallMaps.rootCallMap || {};
    this._storage = [];
    if (this._queuedRecords) {
      this._storage.push(this._queuedRecords);
    }
    if (this._records) {
      this._storage.push(this._records);
    }
    if (this._cachedRecords) {
      this._storage.push(this._cachedRecords);
    }
  }

  /**
   * Get the data ID associated with a storage key (and optionally an
   * identifying argument value) for a root query.
   */


  RelayRecordStore.prototype.getDataID = function getDataID(storageKey, identifyingArgValue) {
    if (require('./RelayNodeInterface').isNodeRootCall(storageKey)) {
      require('fbjs/lib/invariant')(identifyingArgValue != null, 'RelayRecordStore.getDataID(): Argument to `%s()` ' + 'cannot be null or undefined.', storageKey);
      return identifyingArgValue;
    }
    if (identifyingArgValue == null) {
      identifyingArgValue = EMPTY;
    }
    if (this._rootCallMap.hasOwnProperty(storageKey) && this._rootCallMap[storageKey].hasOwnProperty(identifyingArgValue)) {
      return this._rootCallMap[storageKey][identifyingArgValue];
    } else if (this._cachedRootCallMap.hasOwnProperty(storageKey)) {
      return this._cachedRootCallMap[storageKey][identifyingArgValue];
    }
  };

  /**
   * Returns the status of the record stored at `dataID`.
   */


  RelayRecordStore.prototype.getRecordState = function getRecordState(dataID) {
    var record = this._getRecord(dataID);
    if (record === null) {
      return 'NONEXISTENT';
    } else if (record === undefined) {
      return 'UNKNOWN';
    }
    return 'EXISTENT';
  };

  /**
   * Returns the path to a non-refetchable record.
   */


  RelayRecordStore.prototype.getPathToRecord = function getPathToRecord(dataID) {
    var path = this._getField(dataID, PATH);
    return path;
  };

  /**
   * Returns whether a given record is affected by an optimistic update.
   */


  RelayRecordStore.prototype.hasOptimisticUpdate = function hasOptimisticUpdate(dataID) {
    var queuedRecords = this._queuedRecords;
    return queuedRecords ? queuedRecords.hasOwnProperty(dataID) : false;
  };

  /**
   * Returns a list of client mutation IDs for queued mutations whose optimistic
   * updates are affecting the record corresponding the given dataID. Returns
   * null if the record isn't affected by any optimistic updates.
   */


  RelayRecordStore.prototype.getClientMutationIDs = function getClientMutationIDs(dataID) {
    var queuedRecords = this._queuedRecords;
    if (queuedRecords) {
      var record = queuedRecords[dataID];
      if (record) {
        return record[MUTATION_IDS];
      }
    }
    return null;
  };

  /**
   * Check whether a given record has received data for a deferred fragment.
   */


  RelayRecordStore.prototype.hasFragmentData = function hasFragmentData(dataID, fragmentID) {
    var resolvedFragmentMap = this._getField(dataID, RESOLVED_FRAGMENT_MAP);
    require('fbjs/lib/invariant')(typeof resolvedFragmentMap === 'object' || resolvedFragmentMap == null, 'RelayRecordStore.hasFragmentData(): Expected the map of ' + 'resolved deferred fragments associated with record `%s` to be null or ' + 'an object. Found a(n) `%s`.', dataID, typeof resolvedFragmentMap);
    return !!(resolvedFragmentMap && resolvedFragmentMap[fragmentID]);
  };

  RelayRecordStore.prototype.getType = function getType(dataID) {
    // `__typename` property is typed as `string`
    return this._getField(dataID, '__typename');
  };

  /**
   * Returns the value of the field for the given dataID.
   */


  RelayRecordStore.prototype.getField = function getField(dataID, storageKey) {
    return this._getField(dataID, storageKey);
  };

  /**
   * Returns the Data ID of a linked record (eg the ID of the `address` record
   * in `actor{address}`).
   */


  RelayRecordStore.prototype.getLinkedRecordID = function getLinkedRecordID(dataID, storageKey) {
    var field = this._getField(dataID, storageKey);
    if (field == null) {
      return field;
    }
    var record = require('./RelayRecord').getRecord(field);
    require('fbjs/lib/invariant')(record, 'RelayRecordStore.getLinkedRecordID(): Expected field `%s` for record ' + '`%s` to have a linked record.', storageKey, dataID);
    return require('./RelayRecord').getDataID(record);
  };

  /**
   * Returns an array of Data ID for a plural linked field (eg the actor IDs of
   * the `likers` in `story{likers}`).
   */


  RelayRecordStore.prototype.getLinkedRecordIDs = function getLinkedRecordIDs(dataID, storageKey) {
    var field = this._getField(dataID, storageKey);
    if (field == null) {
      return field;
    }
    require('fbjs/lib/invariant')(Array.isArray(field), 'RelayRecordStore.getLinkedRecordIDs(): Expected field `%s` for ' + 'record `%s` to have an array of linked records.', storageKey, dataID);
    return field.map(function (element, ii) {
      var record = require('./RelayRecord').getRecord(element);
      require('fbjs/lib/invariant')(record, 'RelayRecordStore.getLinkedRecordIDs(): Expected element at index %s ' + 'in field `%s` for record `%s` to be a linked record.', ii, storageKey, dataID);
      return require('./RelayRecord').getDataID(record);
    });
  };

  /**
   * Gets the connectionIDs for all the connections that contain the given
   * record as a `node`, or null if the record does not appear as a `node` in
   * any connection.
   */


  RelayRecordStore.prototype.getConnectionIDsForRecord = function getConnectionIDsForRecord(dataID) {
    var connectionIDs = this._nodeConnectionMap[dataID];
    if (connectionIDs) {
      return (0, _keys2['default'])(connectionIDs);
    }
    return null;
  };

  /**
   * Gets the connectionIDs for all variations of calls for the given base
   * schema name (Ex: `posts.orderby(recent)` and `posts.orderby(likes)`).
   */


  RelayRecordStore.prototype.getConnectionIDsForField = function getConnectionIDsForField(dataID, schemaName) {
    // ignore queued records because not all range fields may be present there
    var record = this._records[dataID];
    if (record == null) {
      return record;
    }
    var connectionIDs = void 0;
    require('fbjs/lib/forEachObject')(record, function (datum, key) {
      if (datum && getFieldNameFromKey(key) === schemaName) {
        var connectionID = require('./RelayRecord').getDataIDForObject(datum);
        if (connectionID) {
          connectionIDs = connectionIDs || [];
          connectionIDs.push(connectionID);
        }
      }
    });
    return connectionIDs;
  };

  /**
   * Get the force index associated with the range at `connectionID`.
   */


  RelayRecordStore.prototype.getRangeForceIndex = function getRangeForceIndex(connectionID) {
    var forceIndex = this._getField(connectionID, FORCE_INDEX);
    if (forceIndex === null) {
      return -1;
    }
    return forceIndex || 0;
  };

  /**
   * Get the condition calls that were used to fetch the given connection.
   * Ex: for a field `photos.orderby(recent)`, this would be
   * [{name: 'orderby', value: 'recent'}]
   */


  RelayRecordStore.prototype.getRangeFilterCalls = function getRangeFilterCalls(connectionID) {
    return this._getField(connectionID, FILTER_CALLS);
  };

  /**
   * Returns range information for the given connection field:
   * - `filteredEdges`: any edges already fetched for the given `calls`.
   * - `diffCalls`: an array of calls describing the difference
   *   between the given `calls` and already fetched data. Includes conditional
   *   calls (`orderby`) and range/offset calls (`first`, `after`).
   * - `filterCalls`: the subset of `calls` that are condition calls
   *   (`orderby`).
   */


  RelayRecordStore.prototype.getRangeMetadata = function getRangeMetadata(connectionID, calls) {
    var _this = this;

    if (connectionID == null) {
      return connectionID;
    }
    var range = this._getField(connectionID, RANGE);
    if (range == null) {
      if (range === null) {
        require('fbjs/lib/warning')(false, 'RelayRecordStore.getRangeMetadata(): Expected range to exist if ' + '`edges` has been fetched.');
      }
      return undefined;
    }
    var filterCalls = getFilterCalls(calls);
    // Edges can only be fetched if a range call (first/last/find) is given.
    // Otherwise return diffCalls/filterCalls with empty edges.
    if (calls.length === filterCalls.length) {
      return {
        diffCalls: calls,
        filterCalls: filterCalls,
        pageInfo: undefined,
        requestedEdgeIDs: [],
        filteredEdges: []
      };
    }
    var queuedRecord = this._queuedRecords ? this._queuedRecords[connectionID] : null;
    var rangeInfo = range.retrieveRangeInfoForQuery(calls, queuedRecord);
    var diffCalls = rangeInfo.diffCalls;
    var pageInfo = rangeInfo.pageInfo;
    var requestedEdgeIDs = rangeInfo.requestedEdgeIDs;

    if (diffCalls && diffCalls.length) {
      diffCalls = filterCalls.concat(diffCalls);
    } else {
      diffCalls = [];
    }
    var filteredEdges = void 0;
    if (requestedEdgeIDs) {
      filteredEdges = requestedEdgeIDs.map(function (edgeID) {
        return {
          edgeID: edgeID,
          nodeID: _this.getLinkedRecordID(edgeID, NODE)
        };
      }).filter(function (edge) {
        return _this._getRecord(edge.nodeID);
      });
    } else {
      filteredEdges = [];
    }
    return {
      diffCalls: diffCalls,
      filterCalls: filterCalls,
      pageInfo: pageInfo,
      requestedEdgeIDs: requestedEdgeIDs,
      filteredEdges: filteredEdges
    };
  };

  /**
   * Returns whether there is a range at `connectionID`.
   */


  RelayRecordStore.prototype.hasRange = function hasRange(connectionID) {
    return !!this._getField(connectionID, RANGE);
  };

  /**
   * Completely removes the record identified by `dataID` from the store.
   * This is only used by garbage collection.
   */


  RelayRecordStore.prototype.removeRecord = function removeRecord(dataID) {
    delete this._records[dataID];
    if (this._queuedRecords) {
      delete this._queuedRecords[dataID];
    }
    if (this._cachedRecords) {
      delete this._cachedRecords[dataID];
    }
    delete this._nodeConnectionMap[dataID];
  };

  /**
   * Gets the first version of the record from the available caches.
   */


  RelayRecordStore.prototype._getRecord = function _getRecord(dataID) {
    if (this._queuedRecords && this._queuedRecords.hasOwnProperty(dataID)) {
      return this._queuedRecords[dataID];
    } else if (this._records.hasOwnProperty(dataID)) {
      return this._records[dataID];
    } else if (this._cachedRecords) {
      return this._cachedRecords[dataID];
    }
  };

  /**
   * Get the value of the field from the first version of the record for which
   * the field is defined, returning `null` if the record has been deleted or
   * `undefined` if the record has not been fetched.
   */


  RelayRecordStore.prototype._getField = function _getField(dataID, storageKey) {
    var storage = this._storage;
    for (var ii = 0; ii < storage.length; ii++) {
      var record = storage[ii][dataID];
      if (record === null) {
        return null;
      } else if (record && record.hasOwnProperty(storageKey)) {
        return record[storageKey];
      }
    }
    return undefined;
  };

  return RelayRecordStore;
}();

/**
 * Filter calls to only those that specify conditions on the returned results
 * (ex: `orderby(TOP_STORIES)`), removing generic calls (ex: `first`, `find`).
 */


function getFilterCalls(calls) {
  return calls.filter(function (call) {
    return !require('./RelayConnectionInterface').isConnectionCall(call);
  });
}

/**
 * Returns the field name based on the object key used to store the data in
 * nodeData. It returns the field name without any calls. For example, the
 * field name for 'profile_picture{size:"50"}' will be 'profile_picture'
 */
function getFieldNameFromKey(key) {
  // This is based on the GraphQL spec for what constitutes a valid field name.
  return key.split(/(?![_A-Za-z][_0-9A-Za-z]*)/, 1)[0];
}

module.exports = RelayRecordStore;