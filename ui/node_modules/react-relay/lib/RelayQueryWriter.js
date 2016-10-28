/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayQueryWriter
 * 
 */

'use strict';

var _extends3 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var ANY_TYPE = require('./RelayNodeInterface').ANY_TYPE;

var ID = require('./RelayNodeInterface').ID;

var TYPENAME = require('./RelayNodeInterface').TYPENAME;

var EDGES = require('./RelayConnectionInterface').EDGES;

var NODE = require('./RelayConnectionInterface').NODE;

var PAGE_INFO = require('./RelayConnectionInterface').PAGE_INFO;

var EXISTENT = require('./RelayRecordState').EXISTENT;

/**
 * @internal
 *
 * Helper for writing the result of one or more queries/operations into the
 * store, updating tracked queries, and recording changed record IDs.
 */


var RelayQueryWriter = function (_RelayQueryVisitor) {
  (0, _inherits3['default'])(RelayQueryWriter, _RelayQueryVisitor);

  function RelayQueryWriter(store, writer, queryTracker, changeTracker, options) {
    (0, _classCallCheck3['default'])(this, RelayQueryWriter);

    var _this = (0, _possibleConstructorReturn3['default'])(this, _RelayQueryVisitor.call(this));

    _this._changeTracker = changeTracker;
    _this._forceIndex = options && options.forceIndex ? options.forceIndex : 0;
    _this._isOptimisticUpdate = !!(options && options.isOptimisticUpdate);
    _this._store = store;
    _this._queryTracker = queryTracker;
    _this._updateTrackedQueries = !!(options && options.updateTrackedQueries);
    _this._writer = writer;
    return _this;
  }

  RelayQueryWriter.prototype.getRecordStore = function getRecordStore() {
    return this._store;
  };

  RelayQueryWriter.prototype.getRecordWriter = function getRecordWriter() {
    return this._writer;
  };

  RelayQueryWriter.prototype.getRecordTypeName = function getRecordTypeName(node, recordID, payload) {
    if (this._isOptimisticUpdate) {
      // Optimistic queries are inferred. Reuse existing type if available.
      return this._store.getType(recordID);
    }
    var typeName = payload[TYPENAME];
    if (typeName == null) {
      if (!node.isAbstract()) {
        typeName = node.getType();
      } else {
        typeName = this._store.getType(recordID);
      }
    }
    require('fbjs/lib/warning')(typeName && typeName !== ANY_TYPE, 'RelayQueryWriter: Could not find a type name for record `%s`.', recordID);
    return typeName || null;
  };

  /**
   * Traverses a query and payload in parallel, writing the results into the
   * store.
   */


  RelayQueryWriter.prototype.writePayload = function writePayload(node, recordID, responseData, path) {
    var _this2 = this;

    var state = {
      nodeID: null,
      path: path,
      recordID: recordID,
      responseData: responseData
    };

    if (node instanceof require('./RelayQuery').Field && node.canHaveSubselections()) {
      // for non-scalar fields, the recordID is the parent
      node.getChildren().forEach(function (child) {
        _this2.visit(child, state);
      });
      return;
    }

    this.visit(node, state);
  };

  /**
   * Records are "created" whenever an entry did not previously exist for the
   * `recordID`, including cases when a `recordID` is created with a null value.
   */


  RelayQueryWriter.prototype.recordCreate = function recordCreate(recordID) {
    this._changeTracker.createID(recordID);
  };

  /**
   * Records are "updated" if any field changes (including being set to null).
   * Updates are not recorded for newly created records.
   */


  RelayQueryWriter.prototype.recordUpdate = function recordUpdate(recordID) {
    this._changeTracker.updateID(recordID);
  };

  /**
   * Determine if the record was created by this write operation.
   */


  RelayQueryWriter.prototype.isNewRecord = function isNewRecord(recordID) {
    return this._changeTracker.isNewRecord(recordID);
  };

  /**
   * Helper to create a record and the corresponding notification.
   */


  RelayQueryWriter.prototype.createRecordIfMissing = function createRecordIfMissing(node, recordID, path, payload) {
    var recordState = this._store.getRecordState(recordID);
    var typeName = payload && this.getRecordTypeName(node, recordID, payload);
    this._writer.putRecord(recordID, typeName, path);
    if (recordState !== EXISTENT) {
      this.recordCreate(recordID);
    }
    if (this._queryTracker && (this.isNewRecord(recordID) || this._updateTrackedQueries) && (!require('./RelayRecord').isClientID(recordID) || require('./RelayQueryPath').isRootPath(path))) {
      this._queryTracker.trackNodeForID(node, recordID);
    }
  };

  RelayQueryWriter.prototype.visitRoot = function visitRoot(root, state) {
    var path = state.path;
    var recordID = state.recordID;
    var responseData = state.responseData;

    var recordState = this._store.getRecordState(recordID);

    // GraphQL should never return undefined for a field
    if (responseData == null) {
      require('fbjs/lib/invariant')(responseData !== undefined, 'RelayQueryWriter: Unexpectedly encountered `undefined` in payload. ' + 'Cannot set root record `%s` to undefined.', recordID);
      this._writer.deleteRecord(recordID);
      if (recordState === EXISTENT) {
        this.recordUpdate(recordID);
      }
      return;
    }
    require('fbjs/lib/invariant')(typeof responseData === 'object' && responseData !== null, 'RelayQueryWriter: Cannot update record `%s`, expected response to be ' + 'an array or object.', recordID);
    this.createRecordIfMissing(root, recordID, path, responseData);
    this.traverse(root, state);
  };

  RelayQueryWriter.prototype.visitFragment = function visitFragment(fragment, state) {
    var recordID = state.recordID;

    if (fragment.isDeferred()) {
      var hash = fragment.getSourceCompositeHash() || fragment.getCompositeHash();

      this._writer.setHasDeferredFragmentData(recordID, hash);

      this.recordUpdate(recordID);
    }
    // Skip fragments that do not match the record's concrete type. Fragments
    // cannot be skipped for optimistic writes because optimistically created
    // records *may* have a default `Node` type.
    if (this._isOptimisticUpdate || require('./isCompatibleRelayFragmentType')(fragment, this._store.getType(recordID))) {
      if (!this._isOptimisticUpdate && fragment.isTrackingEnabled()) {
        this._writer.setHasFragmentData(recordID, fragment.getCompositeHash());
      }
      var _path = require('./RelayQueryPath').getPath(state.path, fragment, recordID);
      this.traverse(fragment, (0, _extends3['default'])({}, state, {
        path: _path
      }));
    }
  };

  RelayQueryWriter.prototype.visitField = function visitField(field, state) {
    var recordID = state.recordID;
    var responseData = state.responseData;

    require('fbjs/lib/invariant')(this._writer.getRecordState(recordID) === EXISTENT, 'RelayQueryWriter: Cannot update a non-existent record, `%s`.', recordID);
    require('fbjs/lib/invariant')(typeof responseData === 'object' && responseData !== null, 'RelayQueryWriter: Cannot update record `%s`, expected response to be ' + 'an object.', recordID);
    var serializationKey = field.getSerializationKey();

    var fieldData = responseData[serializationKey];
    // Queried fields that are `undefined` are stored as nulls.
    if (fieldData == null) {
      if (fieldData === undefined) {
        if (responseData.hasOwnProperty(serializationKey)) {
          require('fbjs/lib/warning')(false, 'RelayQueryWriter: Encountered an explicit `undefined` field `%s` ' + 'on record `%s`, expected response to not contain `undefined`.', field.getDebugName(), recordID);
          return;
        } else if (this._isOptimisticUpdate) {

          return;
        }
      }

      var storageKey = field.getStorageKey();
      var prevValue = this._store.getField(recordID, storageKey);
      // Always write to ensure data is stored in the correct recordStore.
      this._writer.deleteField(recordID, storageKey);
      if (prevValue !== null) {
        this.recordUpdate(recordID);
      }
      return;
    }

    if (!field.canHaveSubselections()) {
      this._writeScalar(field, state, recordID, fieldData);
    } else if (field.isConnection()) {
      this._writeConnection(field, state, recordID, fieldData);
    } else if (field.isPlural()) {
      this._writePluralLink(field, state, recordID, fieldData);
    } else {
      this._writeLink(field, state, recordID, fieldData);
    }
  };

  /**
   * Writes the value for a 'scalar' field such as `id` or `name`. The response
   * data is expected to be scalar values or arrays of scalar values.
   */


  RelayQueryWriter.prototype._writeScalar = function _writeScalar(field, state, recordID, nextValue) {
    var storageKey = field.getStorageKey();
    var prevValue = this._store.getField(recordID, storageKey);

    // always update the store to ensure the value is present in the appropriate
    // data sink (records/queuedRecords), but only record an update if the value
    // changed.
    this._writer.putField(recordID, storageKey, nextValue);

    // TODO: Flow: `nextValue` is an array, array indexing should work
    if (Array.isArray(prevValue) && Array.isArray(nextValue) && prevValue.length === nextValue.length && prevValue.every(function (prev, ii) {
      return prev === nextValue[ii];
    })) {
      return;
    } else if (prevValue === nextValue) {
      return;
    }
    this.recordUpdate(recordID);
  };

  /**
   * Writes data for connection fields such as `news_feed` or `friends`. The
   * response data is expected to be array of edge objects.
   */


  RelayQueryWriter.prototype._writeConnection = function _writeConnection(field, state, recordID, connectionData) {
    // Each unique combination of filter calls is stored in its own
    // generated record (ex: `field.orderby(x)` results are separate from
    // `field.orderby(y)` results).
    var storageKey = field.getStorageKey();
    var connectionID = this._store.getLinkedRecordID(recordID, storageKey) || require('./generateClientID')();

    var connectionRecordState = this._store.getRecordState(connectionID);
    var hasEdges = !!(field.getFieldByStorageKey(EDGES) || connectionData != null && typeof connectionData === 'object' && connectionData[EDGES]);
    var path = require('./RelayQueryPath').getPath(state.path, field, connectionID);
    // always update the store to ensure the value is present in the appropriate
    // data sink (records/queuedRecords), but only record an update if the value
    // changed.
    this._writer.putRecord(connectionID, null, path);
    this._writer.putLinkedRecordID(recordID, storageKey, connectionID);
    // record the create/update only if something changed
    if (connectionRecordState !== EXISTENT) {
      this.recordUpdate(recordID);
      this.recordCreate(connectionID);
    }

    // Only create a range if `edges` field is present
    // Overwrite an existing range only if the new force index is greater
    if (hasEdges && (!this._writer.hasRange(connectionID) || this._forceIndex && this._forceIndex > this._store.getRangeForceIndex(connectionID))) {
      this._writer.putRange(connectionID, field.getCallsWithValues(), this._forceIndex);
      this.recordUpdate(connectionID);
    }

    var connectionState = {
      nodeID: null,
      path: path,
      recordID: connectionID,
      responseData: connectionData
    };
    this._traverseConnection(field, field, connectionState);
  };

  /**
   * Recurse through connection subfields and write their results. This is
   * necessary because handling an `edges` field also requires information about
   * the parent connection field (see `_writeEdges`).
   */


  RelayQueryWriter.prototype._traverseConnection = function _traverseConnection(connection, // the parent connection
  node, // the parent connection or an intermediary fragment
  state) {
    var _this3 = this;

    node.getChildren().forEach(function (child) {
      if (child instanceof require('./RelayQuery').Field) {
        if (child.getSchemaName() === EDGES) {
          _this3._writeEdges(connection, child, state);
        } else if (child.getSchemaName() !== PAGE_INFO) {
          // Page info is handled by the range
          // Otherwise, write metadata fields normally (ex: `count`)
          _this3.visit(child, state);
        }
      } else {
        // Fragment case, recurse keeping track of parent connection
        _this3._traverseConnection(connection, child, state);
      }
    });
  };

  /**
   * Update a connection with newly fetched edges.
   */


  RelayQueryWriter.prototype._writeEdges = function _writeEdges(connection, edges, state) {
    var _this4 = this;

    var connectionID = state.recordID;
    var connectionData = state.responseData;

    require('fbjs/lib/invariant')(typeof connectionData === 'object' && connectionData !== null, 'RelayQueryWriter: Cannot write edges for malformed connection `%s` on ' + 'record `%s`, expected the response to be an object.', connection.getDebugName(), connectionID);
    var edgesData = connectionData[EDGES];

    // Validate response data.
    if (edgesData == null) {
      require('fbjs/lib/warning')(false, 'RelayQueryWriter: Cannot write edges for connection `%s` on record ' + '`%s`, expected a response for field `edges`.', connection.getDebugName(), connectionID);
      return;
    }
    require('fbjs/lib/invariant')(Array.isArray(edgesData), 'RelayQueryWriter: Cannot write edges for connection `%s` on record ' + '`%s`, expected `edges` to be an array.', connection.getDebugName(), connectionID);

    var rangeCalls = connection.getCallsWithValues();
    require('fbjs/lib/invariant')(require('./RelayConnectionInterface').hasRangeCalls(rangeCalls), 'RelayQueryWriter: Cannot write edges for connection on record ' + '`%s` without `first`, `last`, or `find` argument.', connectionID);
    var rangeInfo = this._store.getRangeMetadata(connectionID, rangeCalls);
    require('fbjs/lib/invariant')(rangeInfo, 'RelayQueryWriter: Expected a range to exist for connection field `%s` ' + 'on record `%s`.', connection.getDebugName(), connectionID);
    var fetchedEdgeIDs = [];
    var filteredEdges = rangeInfo.filteredEdges;
    var isUpdate = false;
    var nextIndex = 0;
    // Traverse connection edges, reusing existing edges if they exist
    edgesData.forEach(function (edgeData) {
      // validate response data
      if (edgeData == null) {
        return;
      }
      require('fbjs/lib/invariant')(typeof edgeData === 'object' && edgeData, 'RelayQueryWriter: Cannot write edge for connection field `%s` on ' + 'record `%s`, expected an object.', connection.getDebugName(), connectionID);

      var nodeData = edgeData[NODE];
      if (nodeData == null) {
        return;
      }

      require('fbjs/lib/invariant')(typeof nodeData === 'object', 'RelayQueryWriter: Expected node to be an object for field `%s` on ' + 'record `%s`.', connection.getDebugName(), connectionID);

      // For consistency, edge IDs are calculated from the connection & node ID.
      // A node ID is only generated if the node does not have an id and
      // there is no existing edge.
      var prevEdge = filteredEdges[nextIndex++];
      var nodeID = nodeData && nodeData[ID] || prevEdge && _this4._store.getLinkedRecordID(prevEdge.edgeID, NODE) || require('./generateClientID')();
      // TODO: Flow: `nodeID` is `string`
      var edgeID = require('./generateClientEdgeID')(connectionID, nodeID);
      var path = require('./RelayQueryPath').getPath(state.path, edges, edgeID);
      _this4.createRecordIfMissing(edges, edgeID, path, null);
      fetchedEdgeIDs.push(edgeID);

      // Write data for the edge, using `nodeID` as the id for direct descendant
      // `node` fields. This is necessary for `node`s that do not have an `id`,
      // which would cause the generated ID here to not match the ID generated
      // in `_writeLink`.
      _this4.traverse(edges, {
        nodeID: nodeID,
        path: path,
        recordID: edgeID,
        responseData: edgeData
      });
      isUpdate = isUpdate || !prevEdge || edgeID !== prevEdge.edgeID;
    });

    var pageInfo = connectionData[PAGE_INFO] || require('./RelayConnectionInterface').getDefaultPageInfo();
    this._writer.putRangeEdges(connectionID, rangeCalls, pageInfo, fetchedEdgeIDs);

    // Only broadcast an update to the range if an edge was added/changed.
    // Node-level changes will broadcast at the node ID.
    if (isUpdate) {
      this.recordUpdate(connectionID);
    }
  };

  /**
   * Writes a plural linked field such as `actors`. The response data is
   * expected to be an array of item objects. These fields are similar to
   * connections, but do not support range calls such as `first` or `after`.
   */


  RelayQueryWriter.prototype._writePluralLink = function _writePluralLink(field, state, recordID, fieldData) {
    var _this5 = this;

    var storageKey = field.getStorageKey();
    require('fbjs/lib/invariant')(Array.isArray(fieldData), 'RelayQueryWriter: Expected array data for field `%s` on record `%s`.', field.getDebugName(), recordID);

    var prevLinkedIDs = this._store.getLinkedRecordIDs(recordID, storageKey);
    var nextLinkedIDs = [];
    var nextRecords = {};
    var isUpdate = false;
    var nextIndex = 0;
    fieldData.forEach(function (nextRecord) {
      // validate response data
      if (nextRecord == null) {
        return;
      }
      require('fbjs/lib/invariant')(typeof nextRecord === 'object' && nextRecord, 'RelayQueryWriter: Expected elements for plural field `%s` to be ' + 'objects.', storageKey);

      // Reuse existing generated IDs if the node does not have its own `id`.
      var prevLinkedID = prevLinkedIDs && prevLinkedIDs[nextIndex];
      var nextLinkedID = nextRecord[ID] || prevLinkedID || require('./generateClientID')();
      nextLinkedIDs.push(nextLinkedID);

      var path = require('./RelayQueryPath').getPath(state.path, field, nextLinkedID);
      _this5.createRecordIfMissing(field, nextLinkedID, path, nextRecord);
      nextRecords[nextLinkedID] = { record: nextRecord, path: path };
      isUpdate = isUpdate || nextLinkedID !== prevLinkedID;
      nextIndex++;
    });
    // Write the linked records before traverse to prevent generating extraneous
    // client ids.
    this._writer.putLinkedRecordIDs(recordID, storageKey, nextLinkedIDs);
    nextLinkedIDs.forEach(function (nextLinkedID) {
      var itemData = nextRecords[nextLinkedID];
      if (itemData) {
        _this5.traverse(field, {
          nodeID: null, // never propagate `nodeID` past the first linked field
          path: itemData.path,
          recordID: nextLinkedID,
          responseData: itemData.record
        });
      }
    });
    // Only broadcast a list-level change if a record was changed/added/removed
    if (isUpdate || !prevLinkedIDs || prevLinkedIDs.length !== nextLinkedIDs.length) {
      this.recordUpdate(recordID);
    }
  };

  /**
   * Writes a link from one record to another, for example linking the `viewer`
   * record to the `actor` record in the query `viewer { actor }`. The `field`
   * variable is the field being linked (`actor` in the example).
   */


  RelayQueryWriter.prototype._writeLink = function _writeLink(field, state, recordID, fieldData) {
    var nodeID = state.nodeID;

    var storageKey = field.getStorageKey();
    require('fbjs/lib/invariant')(typeof fieldData === 'object' && fieldData !== null, 'RelayQueryWriter: Expected data for non-scalar field `%s` on record ' + '`%s` to be an object.', field.getDebugName(), recordID);

    // Prefer the actual `id` if present, otherwise generate one (if an id
    // was already generated it is reused). `node`s within a connection are
    // a special case as the ID used here must match the one generated prior to
    // storing the parent `edge`.
    var prevLinkedID = this._store.getLinkedRecordID(recordID, storageKey);
    var nextLinkedID = field.getSchemaName() === NODE && nodeID || fieldData[ID] || prevLinkedID || require('./generateClientID')();

    var path = require('./RelayQueryPath').getPath(state.path, field, nextLinkedID);
    this.createRecordIfMissing(field, nextLinkedID, path, fieldData);
    // always update the store to ensure the value is present in the appropriate
    // data sink (record/queuedRecords), but only record an update if the value
    // changed.
    this._writer.putLinkedRecordID(recordID, storageKey, nextLinkedID);
    if (prevLinkedID !== nextLinkedID) {
      this.recordUpdate(recordID);
    }

    this.traverse(field, {
      nodeID: null,
      path: path,
      recordID: nextLinkedID,
      responseData: fieldData
    });
  };

  return RelayQueryWriter;
}(require('./RelayQueryVisitor'));

module.exports = RelayQueryWriter;