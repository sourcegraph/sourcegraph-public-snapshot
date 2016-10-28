/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayRecordWriter
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var EMPTY = '';

var APPEND = require('./GraphQLMutatorConstants').APPEND;

var PREPEND = require('./GraphQLMutatorConstants').PREPEND;

var REMOVE = require('./GraphQLMutatorConstants').REMOVE;

var CURSOR = require('./RelayConnectionInterface').CURSOR;

var NODE = require('./RelayConnectionInterface').NODE;

var _RelayRecord$Metadata = require('./RelayRecord').MetadataKey;

var FILTER_CALLS = _RelayRecord$Metadata.FILTER_CALLS;
var FORCE_INDEX = _RelayRecord$Metadata.FORCE_INDEX;
var MUTATION_IDS = _RelayRecord$Metadata.MUTATION_IDS;
var PATH = _RelayRecord$Metadata.PATH;
var RANGE = _RelayRecord$Metadata.RANGE;
var RESOLVED_FRAGMENT_MAP = _RelayRecord$Metadata.RESOLVED_FRAGMENT_MAP;
var RESOLVED_FRAGMENT_MAP_GENERATION = _RelayRecord$Metadata.RESOLVED_FRAGMENT_MAP_GENERATION;
var STATUS = _RelayRecord$Metadata.STATUS;


/**
 * @internal
 *
 * `RelayRecordWriter` is the helper module to write data into RelayRecordStore.
 */

var RelayRecordWriter = function () {
  function RelayRecordWriter(records, rootCallMap, isOptimistic, nodeConnectionMap, cacheWriter, clientMutationID) {
    (0, _classCallCheck3['default'])(this, RelayRecordWriter);

    this._cacheWriter = cacheWriter;
    this._clientMutationID = clientMutationID;
    this._isOptimisticWrite = isOptimistic;
    this._nodeConnectionMap = nodeConnectionMap || {};
    this._records = records;
    this._rootCallMap = rootCallMap;
  }

  /**
   * Get the data ID associated with a storage key (and optionally an
   * identifying argument value) for a root query.
   */


  RelayRecordWriter.prototype.getDataID = function getDataID(storageKey, identifyingArgValue) {
    if (require('./RelayNodeInterface').isNodeRootCall(storageKey)) {
      require('fbjs/lib/invariant')(identifyingArgValue != null, 'RelayRecordWriter.getDataID(): Argument to `%s()` ' + 'cannot be null or undefined.', storageKey);
      return identifyingArgValue;
    }
    if (identifyingArgValue == null) {
      identifyingArgValue = EMPTY;
    }
    if (this._rootCallMap.hasOwnProperty(storageKey) && this._rootCallMap[storageKey].hasOwnProperty(identifyingArgValue)) {
      return this._rootCallMap[storageKey][identifyingArgValue];
    }
  };

  /**
   * Associate a data ID with a storage key (and optionally an identifying
   * argument value) for a root query.
   */


  RelayRecordWriter.prototype.putDataID = function putDataID(storageKey, identifyingArgValue, dataID) {
    if (require('./RelayNodeInterface').isNodeRootCall(storageKey)) {
      require('fbjs/lib/invariant')(identifyingArgValue != null, 'RelayRecordWriter.putDataID(): Argument to `%s()` ' + 'cannot be null or undefined.', storageKey);
      return;
    }
    if (identifyingArgValue == null) {
      identifyingArgValue = EMPTY;
    }
    this._rootCallMap[storageKey] = this._rootCallMap[storageKey] || {};
    this._rootCallMap[storageKey][identifyingArgValue] = dataID;
    if (this._cacheWriter) {
      this._cacheWriter.writeRootCall(storageKey, identifyingArgValue, dataID);
    }
  };

  /**
   * Returns the status of the record stored at `dataID`.
   */


  RelayRecordWriter.prototype.getRecordState = function getRecordState(dataID) {
    var record = this._records[dataID];
    if (record === null) {
      return 'NONEXISTENT';
    } else if (record === undefined) {
      return 'UNKNOWN';
    }
    return 'EXISTENT';
  };

  /**
   * Create an empty record at `dataID` if a record does not already exist.
   */


  RelayRecordWriter.prototype.putRecord = function putRecord(dataID, typeName, path) {
    var prevRecord = this._getRecordForWrite(dataID);
    if (prevRecord) {
      return;
    }
    var nextRecord = require('./RelayRecord').createWithFields(dataID, {
      __typename: typeName
    });
    if (this._isOptimisticWrite) {
      this._setClientMutationID(nextRecord);
    }
    if (require('./RelayRecord').isClientID(dataID) && path) {
      nextRecord[PATH] = path;
    }
    this._records[dataID] = nextRecord;
    var cacheWriter = this._cacheWriter;
    if (!this._isOptimisticWrite && cacheWriter) {
      cacheWriter.writeField(dataID, '__dataID__', dataID, typeName);
    }
  };

  /**
   * Returns the path to a non-refetchable record.
   */


  RelayRecordWriter.prototype.getPathToRecord = function getPathToRecord(dataID) {
    return this._getField(dataID, PATH);
  };

  /**
   * Check whether a given record has received data for a deferred fragment.
   */


  RelayRecordWriter.prototype.hasFragmentData = function hasFragmentData(dataID, fragmentID) {
    var resolvedFragmentMap = this._getField(dataID, RESOLVED_FRAGMENT_MAP);
    require('fbjs/lib/invariant')(typeof resolvedFragmentMap === 'object' || resolvedFragmentMap == null, 'RelayRecordWriter.hasFragmentData(): Expected the map of ' + 'resolved deferred fragments associated with record `%s` to be null or ' + 'an object. Found a(n) `%s`.', dataID, typeof resolvedFragmentMap);
    return !!(resolvedFragmentMap && resolvedFragmentMap[fragmentID]);
  };

  /**
   * Mark a given record as having received data for a deferred fragment.
   */


  RelayRecordWriter.prototype.setHasDeferredFragmentData = function setHasDeferredFragmentData(dataID, fragmentID) {
    this._setHasFragmentData(dataID, fragmentID, true);
  };

  /**
   * Mark a given record as having received data for a fragment.
   */


  RelayRecordWriter.prototype.setHasFragmentData = function setHasFragmentData(dataID, fragmentID) {
    this._setHasFragmentData(dataID, fragmentID, false);
  };

  RelayRecordWriter.prototype._setHasFragmentData = function _setHasFragmentData(dataID, fragmentID, updateFragmentGeneration) {
    var record = this._getRecordForWrite(dataID);
    require('fbjs/lib/invariant')(record, 'RelayRecordWriter.setHasFragmentData(): Expected record `%s` ' + 'to exist before marking it as having received data for the deferred ' + 'fragment with id `%s`.', dataID, fragmentID);
    var resolvedFragmentMap = record[RESOLVED_FRAGMENT_MAP];
    if (typeof resolvedFragmentMap !== 'object' || !resolvedFragmentMap) {
      resolvedFragmentMap = {};
    }
    resolvedFragmentMap[fragmentID] = true;
    record[RESOLVED_FRAGMENT_MAP] = resolvedFragmentMap;
    if (updateFragmentGeneration) {
      if (typeof record[RESOLVED_FRAGMENT_MAP_GENERATION] === 'number') {
        record[RESOLVED_FRAGMENT_MAP_GENERATION]++;
      } else {
        record[RESOLVED_FRAGMENT_MAP_GENERATION] = 0;
      }
    }
  };

  /**
   * Delete the record at `dataID`, setting its value to `null`.
   */


  RelayRecordWriter.prototype.deleteRecord = function deleteRecord(dataID) {
    this._records[dataID] = null;

    // Remove any links for this record
    if (!this._isOptimisticWrite) {
      delete this._nodeConnectionMap[dataID];
      if (this._cacheWriter) {
        this._cacheWriter.writeNode(dataID, null);
      }
    }
  };

  RelayRecordWriter.prototype.getType = function getType(dataID) {
    // `__typename` property is typed as `string`
    return this._getField(dataID, '__typename');
  };

  /**
   * Returns the value of the field for the given dataID.
   */


  RelayRecordWriter.prototype.getField = function getField(dataID, storageKey) {
    return this._getField(dataID, storageKey);
  };

  /**
   * Sets the value of a scalar field.
   */


  RelayRecordWriter.prototype.putField = function putField(dataID, storageKey, value) {
    var record = this._getRecordForWrite(dataID);
    require('fbjs/lib/invariant')(record, 'RelayRecordWriter.putField(): Expected record `%s` to exist before ' + 'writing field `%s`.', dataID, storageKey);
    record[storageKey] = value;
    if (!this._isOptimisticWrite && this._cacheWriter) {
      var typeName = record.__typename;
      this._cacheWriter.writeField(dataID, storageKey, value, typeName);
    }
  };

  /**
   * Clears the value of a field by setting it to null/undefined.
   */


  RelayRecordWriter.prototype.deleteField = function deleteField(dataID, storageKey) {
    var record = this._getRecordForWrite(dataID);
    require('fbjs/lib/invariant')(record, 'RelayRecordWriter.deleteField(): Expected record `%s` to exist before ' + 'deleting field `%s`.', dataID, storageKey);
    record[storageKey] = null;
    if (!this._isOptimisticWrite && this._cacheWriter) {
      this._cacheWriter.writeField(dataID, storageKey, null);
    }
  };

  /**
   * Returns the Data ID of a linked record (eg the ID of the `address` record
   * in `actor{address}`).
   */


  RelayRecordWriter.prototype.getLinkedRecordID = function getLinkedRecordID(dataID, storageKey) {
    var field = this._getField(dataID, storageKey);
    if (field == null) {
      return field;
    }
    var record = require('./RelayRecord').getRecord(field);
    require('fbjs/lib/invariant')(record, 'RelayRecordWriter.getLinkedRecordID(): Expected field `%s` for record ' + '`%s` to have a linked record.', storageKey, dataID);
    return require('./RelayRecord').getDataID(record);
  };

  /**
   * Creates/updates a link between two records via the given field.
   */


  RelayRecordWriter.prototype.putLinkedRecordID = function putLinkedRecordID(parentID, storageKey, recordID) {
    var parent = this._getRecordForWrite(parentID);
    require('fbjs/lib/invariant')(parent, 'RelayRecordWriter.putLinkedRecordID(): Expected record `%s` to exist ' + 'before linking to record `%s`.', parentID, recordID);
    var fieldValue = require('./RelayRecord').create(recordID);
    parent[storageKey] = fieldValue;
    if (!this._isOptimisticWrite && this._cacheWriter) {
      this._cacheWriter.writeField(parentID, storageKey, fieldValue);
    }
  };

  /**
   * Returns an array of Data ID for a plural linked field (eg the actor IDs of
   * the `likers` in `story{likers}`).
   */


  RelayRecordWriter.prototype.getLinkedRecordIDs = function getLinkedRecordIDs(dataID, storageKey) {
    var field = this._getField(dataID, storageKey);
    if (field == null) {
      return field;
    }
    require('fbjs/lib/invariant')(Array.isArray(field), 'RelayRecordWriter.getLinkedRecordIDs(): Expected field `%s` for ' + 'record `%s` to have an array of linked records.', storageKey, dataID);
    return field.map(function (element, ii) {
      var record = require('./RelayRecord').getRecord(element);
      require('fbjs/lib/invariant')(record, 'RelayRecordWriter.getLinkedRecordIDs(): Expected element at index ' + '%s in field `%s` for record `%s` to be a linked record.', ii, storageKey, dataID);
      return require('./RelayRecord').getDataID(record);
    });
  };

  /**
   * Creates/updates a one-to-many link between records via the given field.
   */


  RelayRecordWriter.prototype.putLinkedRecordIDs = function putLinkedRecordIDs(parentID, storageKey, recordIDs) {
    var parent = this._getRecordForWrite(parentID);
    require('fbjs/lib/invariant')(parent, 'RelayRecordWriter.putLinkedRecordIDs(): Expected record `%s` to exist ' + 'before linking records.', parentID);
    var records = recordIDs.map(function (recordID) {
      return require('./RelayRecord').create(recordID);
    });
    parent[storageKey] = records;
    if (!this._isOptimisticWrite && this._cacheWriter) {
      this._cacheWriter.writeField(parentID, storageKey, records);
    }
  };

  /**
   * Get the force index associated with the range at `connectionID`.
   */


  RelayRecordWriter.prototype.getRangeForceIndex = function getRangeForceIndex(connectionID) {
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


  RelayRecordWriter.prototype.getRangeFilterCalls = function getRangeFilterCalls(connectionID) {
    return this._getField(connectionID, FILTER_CALLS);
  };

  /**
   * Creates a range at `dataID` with an optional `forceIndex`.
   */


  RelayRecordWriter.prototype.putRange = function putRange(connectionID, calls, forceIndex) {
    require('fbjs/lib/invariant')(!this._isOptimisticWrite, 'RelayRecordWriter.putRange(): Cannot create a queued range.');
    var record = this._getRecordForWrite(connectionID);
    require('fbjs/lib/invariant')(record, 'RelayRecordWriter.putRange(): Expected record `%s` to exist before ' + 'adding a range.', connectionID);
    var range = new (require('./GraphQLRange'))();
    var filterCalls = getFilterCalls(calls);
    forceIndex = forceIndex || 0;
    record[FILTER_CALLS] = filterCalls;
    record[FORCE_INDEX] = forceIndex;
    record[RANGE] = range;

    var cacheWriter = this._cacheWriter;
    if (!this._isOptimisticWrite && cacheWriter) {
      cacheWriter.writeField(connectionID, FILTER_CALLS, filterCalls);
      cacheWriter.writeField(connectionID, FORCE_INDEX, forceIndex);
      cacheWriter.writeField(connectionID, RANGE, range);
    }
  };

  /**
   * Returns whether there is a range at `connectionID`.
   */


  RelayRecordWriter.prototype.hasRange = function hasRange(connectionID) {
    return !!this._getField(connectionID, RANGE);
  };

  /**
   * Adds newly fetched edges to a range.
   */


  RelayRecordWriter.prototype.putRangeEdges = function putRangeEdges(connectionID, calls, pageInfo, edges) {
    var _this = this;

    var range = this._getField(connectionID, RANGE);
    require('fbjs/lib/invariant')(range, 'RelayRecordWriter.putRangeEdges(): Expected record `%s` to exist and ' + 'have a range.', connectionID);
    var edgeRecords = [];
    edges.forEach(function (edgeID) {
      var edgeRecord = _this._getRangeEdgeRecord(edgeID);
      edgeRecords.push(edgeRecord);
      var nodeID = require('./RelayRecord').getDataID(edgeRecord.node);
      _this._addConnectionForNode(connectionID, nodeID);
    });
    range.addItems(calls, edgeRecords, pageInfo);
    if (!this._isOptimisticWrite && this._cacheWriter) {
      this._cacheWriter.writeField(connectionID, RANGE, range);
    }
  };

  /**
   * Prepend, append, or delete edges to/from a range.
   */


  RelayRecordWriter.prototype.applyRangeUpdate = function applyRangeUpdate(connectionID, edgeID, operation) {
    if (this._isOptimisticWrite) {
      this._applyOptimisticRangeUpdate(connectionID, edgeID, operation);
    } else {
      this._applyServerRangeUpdate(connectionID, edgeID, operation);
    }
  };

  /**
   * Get edge data in a format compatibile with `GraphQLRange`.
   * TODO: change `GraphQLRange` to accept `(edgeID, cursor, nodeID)` tuple
   */


  RelayRecordWriter.prototype._getRangeEdgeRecord = function _getRangeEdgeRecord(edgeID) {
    var nodeID = this.getLinkedRecordID(edgeID, NODE);
    require('fbjs/lib/invariant')(nodeID, 'RelayRecordWriter: Expected edge `%s` to have a `node` record.', edgeID);
    return require('./RelayRecord').createWithFields(edgeID, {
      cursor: this.getField(edgeID, CURSOR),
      node: require('./RelayRecord').create(nodeID)
    });
  };

  RelayRecordWriter.prototype._applyOptimisticRangeUpdate = function _applyOptimisticRangeUpdate(connectionID, edgeID, operation) {
    var record = this._getRecordForWrite(connectionID);
    if (!record) {
      record = require('./RelayRecord').create(connectionID);
      this._records[connectionID] = record;
    }
    this._setClientMutationID(record);
    var key = require('./rangeOperationToMetadataKey')[operation];
    var queue = record[key];
    if (!queue) {
      queue = [];
      record[key] = queue;
    }
    if (operation === PREPEND) {
      queue.unshift(edgeID);
    } else {
      queue.push(edgeID);
    }
  };

  RelayRecordWriter.prototype._applyServerRangeUpdate = function _applyServerRangeUpdate(connectionID, edgeID, operation) {
    var range = this._getField(connectionID, RANGE);
    require('fbjs/lib/invariant')(range, 'RelayRecordWriter: Cannot apply `%s` update to non-existent record ' + '`%s`.', operation, connectionID);
    if (operation === REMOVE) {
      range.removeEdgeWithID(edgeID);
      var nodeID = this.getLinkedRecordID(edgeID, 'node');
      if (nodeID) {
        this._removeConnectionForNode(connectionID, nodeID);
      }
    } else {
      var edgeRecord = this._getRangeEdgeRecord(edgeID);
      var _nodeID = require('./RelayRecord').getDataID(edgeRecord.node);
      this._addConnectionForNode(connectionID, _nodeID);
      if (operation === APPEND) {
        range.appendEdge(this._getRangeEdgeRecord(edgeID));
      } else {
        range.prependEdge(this._getRangeEdgeRecord(edgeID));
      }
    }
    if (this._cacheWriter) {
      this._cacheWriter.writeField(connectionID, RANGE, range);
    }
  };

  /**
   * Record that the node is contained in the connection.
   */


  RelayRecordWriter.prototype._addConnectionForNode = function _addConnectionForNode(connectionID, nodeID) {
    var connectionMap = this._nodeConnectionMap[nodeID];
    if (!connectionMap) {
      connectionMap = {};
      this._nodeConnectionMap[nodeID] = connectionMap;
    }
    connectionMap[connectionID] = true;
  };

  /**
   * Record that the given node is no longer part of the connection.
   */


  RelayRecordWriter.prototype._removeConnectionForNode = function _removeConnectionForNode(connectionID, nodeID) {
    var connectionMap = this._nodeConnectionMap[nodeID];
    if (connectionMap) {
      delete connectionMap[connectionID];
      if ((0, _keys2['default'])(connectionMap).length === 0) {
        delete this._nodeConnectionMap[nodeID];
      }
    }
  };

  /**
   * If the record is in the store, gets a version of the record
   * in the store being used for writes.
   */


  RelayRecordWriter.prototype._getRecordForWrite = function _getRecordForWrite(dataID) {
    var record = this._records[dataID];
    if (!record) {
      return record;
    }
    if (this._isOptimisticWrite) {
      this._setClientMutationID(record);
    }
    return record;
  };

  /**
   * Get the value of the field from the first version of the record for which
   * the field is defined, returning `null` if the record has been deleted or
   * `undefined` if the record has not been fetched.
   */


  RelayRecordWriter.prototype._getField = function _getField(dataID, storageKey) {
    var record = this._records[dataID];
    if (record === null) {
      return null;
    } else if (record && record.hasOwnProperty(storageKey)) {
      return record[storageKey];
    } else {
      return undefined;
    }
  };

  /**
   * Injects the client mutation id associated with the record store instance
   * into the given record.
   */


  RelayRecordWriter.prototype._setClientMutationID = function _setClientMutationID(record) {
    var clientMutationID = this._clientMutationID;
    require('fbjs/lib/invariant')(clientMutationID, 'RelayRecordWriter: _clientMutationID cannot be null/undefined.');
    var mutationIDs = record[MUTATION_IDS] || [];
    if (mutationIDs.indexOf(clientMutationID) === -1) {
      mutationIDs.push(clientMutationID);
      record[MUTATION_IDS] = mutationIDs;
    }
    record[STATUS] = require('./RelayRecordStatusMap').setOptimisticStatus(0, true);
  };

  return RelayRecordWriter;
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

module.exports = RelayRecordWriter;