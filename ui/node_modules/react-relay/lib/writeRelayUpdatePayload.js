/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule writeRelayUpdatePayload
 * 
 */

'use strict';

var _extends3 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _defineProperty3 = _interopRequireDefault(require('babel-runtime/helpers/defineProperty'));

var _stringify2 = _interopRequireDefault(require('babel-runtime/core-js/json/stringify'));

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

// TODO: Replace with enumeration for possible config types.
/* OperationConfig was originally typed such that each property had the type
 * mixed.  Mixed is safer than any, but that safety comes from Flow forcing you
 * to inspect a mixed value at runtime before using it.  However these mixeds
 * are ending up everywhere and are not being inspected */

var CLIENT_MUTATION_ID = require('./RelayConnectionInterface').CLIENT_MUTATION_ID;

var EDGES = require('./RelayConnectionInterface').EDGES;

var ANY_TYPE = require('./RelayNodeInterface').ANY_TYPE;

var ID = require('./RelayNodeInterface').ID;

var NODE = require('./RelayNodeInterface').NODE;

var APPEND = require('./GraphQLMutatorConstants').APPEND;

var IGNORE = require('./GraphQLMutatorConstants').IGNORE;

var PREPEND = require('./GraphQLMutatorConstants').PREPEND;

var REFETCH = require('./GraphQLMutatorConstants').REFETCH;

var REMOVE = require('./GraphQLMutatorConstants').REMOVE;

var EDGES_FIELD = require('./RelayQuery').Field.build({
  fieldName: EDGES,
  type: ANY_TYPE,
  metadata: {
    canHaveSubselections: true,
    isPlural: true
  }
});
var IGNORED_KEYS = (0, _defineProperty3['default'])({
  error: true
}, CLIENT_MUTATION_ID, true);
var STUB_CURSOR_ID = 'client:cursor';

/**
 * @internal
 *
 * Applies the results of an update operation (mutation/subscription) to the
 * store.
 */
function writeRelayUpdatePayload(writer, operation, payload, _ref) {
  var configs = _ref.configs;
  var isOptimisticUpdate = _ref.isOptimisticUpdate;

  configs.forEach(function (config) {
    switch (config.type) {
      case require('./RelayMutationType').NODE_DELETE:
        handleNodeDelete(writer, payload, config);
        break;
      case require('./RelayMutationType').RANGE_ADD:
        handleRangeAdd(writer, payload, operation, config, isOptimisticUpdate);
        break;
      case require('./RelayMutationType').RANGE_DELETE:
        handleRangeDelete(writer, payload, config);
        break;
      case require('./RelayMutationType').FIELDS_CHANGE:
      case require('./RelayMutationType').REQUIRED_CHILDREN:
        break;
      default:
        console.error('Expected a valid mutation handler type, got `%s`.', config.type);
    }
  });

  handleMerge(writer, payload, operation);
}

/**
 * Handles the payload for a node deletion mutation, reading the ID of the node
 * to delete from the payload based on the config and then deleting references
 * to the node.
 */
function handleNodeDelete(writer, payload, config) {
  var recordIDs = payload[config.deletedIDFieldName];
  if (!recordIDs) {
    // for some mutations, deletions don't always occur so if there's no field
    // in the payload, carry on
    return;
  }

  if (Array.isArray(recordIDs)) {
    recordIDs.forEach(function (id) {
      deleteRecord(writer, id);
    });
  } else {
    deleteRecord(writer, recordIDs);
  }
}

/**
 * Deletes the record from the store, also removing any references to the node
 * from any ranges that contain it (along with the containing edges).
 */
function deleteRecord(writer, recordID) {
  var store = writer.getRecordStore();
  var recordWriter = writer.getRecordWriter();
  // skip if already deleted
  var status = store.getRecordState(recordID);
  if (status === require('./RelayRecordState').NONEXISTENT) {
    return;
  }

  // Delete the node from any ranges it may be a part of
  var connectionIDs = store.getConnectionIDsForRecord(recordID);
  if (connectionIDs) {
    connectionIDs.forEach(function (connectionID) {
      var edgeID = require('./generateClientEdgeID')(connectionID, recordID);
      recordWriter.applyRangeUpdate(connectionID, edgeID, REMOVE);
      writer.recordUpdate(edgeID);
      writer.recordUpdate(connectionID);
      // edges are never nodes, so this will not infinitely recurse
      deleteRecord(writer, edgeID);
    });
  }

  // delete the node
  recordWriter.deleteRecord(recordID);
  writer.recordUpdate(recordID);
}

/**
 * Handles merging the results of the mutation/subscription into the store,
 * updating each top-level field in the data according the fetched
 * fields/fragments.
 */
function handleMerge(writer, payload, operation) {
  var store = writer.getRecordStore();

  // because optimistic payloads may not contain all fields, we loop over
  // the data that is present and then have to recurse the query to find
  // the matching fields.
  //
  // TODO #7167718: more efficient mutation/subscription writes
  for (var fieldName in payload) {
    if (!Object.prototype.hasOwnProperty.call(payload, fieldName)) {
      continue;
    }
    var payloadData = payload[fieldName]; // #9357395
    if (typeof payloadData !== 'object' || payloadData == null) {
      continue;
    }
    // if the field is an argument-less root call, determine the corresponding
    // root record ID
    var rootID = store.getDataID(fieldName);
    // check for valid data (has an ID or is an array) and write the field
    if (ID in payloadData || rootID || Array.isArray(payloadData)) {
      mergeField(writer, fieldName, payloadData, operation);
    }
  }
}

/**
 * Merges the results of a single top-level field into the store.
 */
function mergeField(writer, fieldName, payload, operation) {
  // don't write mutation/subscription metadata fields
  if (fieldName in IGNORED_KEYS) {
    return;
  }
  if (Array.isArray(payload)) {
    payload.forEach(function (item) {
      if (typeof item === 'object' && item != null && !Array.isArray(item)) {
        if (getString(item, ID)) {
          mergeField(writer, fieldName, item, operation);
        }
      }
    });
    return;
  }
  // reassign to preserve type information in below closure
  var payloadData = payload;

  var store = writer.getRecordStore();
  var recordID = getString(payloadData, ID);
  var path = void 0;

  if (recordID != null) {
    path = require('./RelayQueryPath').createForID(recordID, 'writeRelayUpdatePayload');
  } else {
    recordID = store.getDataID(fieldName);
    if (!recordID) {
      require('fbjs/lib/invariant')(false, 'writeRelayUpdatePayload(): Expected a record ID in the response ' + 'payload supplied to update the store for field `%s`, ' + 'payload keys [%s], operation name `%s`.', fieldName, (0, _keys2['default'])(payload).join(', '), operation.getName());
    }

    // Root fields that do not accept arguments
    path = require('./RelayQueryPath').create(require('./RelayQuery').Root.build('writeRelayUpdatePayload', fieldName, null, null, {
      identifyingArgName: null,
      identifyingArgType: null,
      isAbstract: true,
      isDeferred: false,
      isPlural: false
    }, ANY_TYPE));
  }
  // write the results for only the current field, for every instance of that
  // field in any subfield/fragment in the query.
  var handleNode = function handleNode(node) {
    node.getChildren().forEach(function (child) {
      if (child instanceof require('./RelayQuery').Fragment) {
        handleNode(child);
      } else if (child instanceof require('./RelayQuery').Field && child.getSerializationKey() === fieldName) {
        // for flow: types are lost in closures
        if (path && recordID) {
          // ensure the record exists and then update it
          writer.createRecordIfMissing(child, recordID, path, payloadData);
          writer.writePayload(child, recordID, payloadData, path);
        }
      }
    });
  };
  handleNode(operation);
}

/**
 * Handles the payload for a range addition. The configuration specifies:
 * - which field in the payload contains data for the new edge
 * - the list of fetched ranges to which the edge should be added
 * - whether to append/prepend to each of those ranges
 */
function handleRangeAdd(writer, payload, operation, config, isOptimisticUpdate) {
  var clientMutationID = getString(payload, CLIENT_MUTATION_ID);
  require('fbjs/lib/invariant')(clientMutationID, 'writeRelayUpdatePayload(): Expected operation `%s` to have a `%s`.', operation.getName(), CLIENT_MUTATION_ID);
  var store = writer.getRecordStore();

  // Extracts the new edge from the payload
  var edge = getObject(payload, config.edgeName);
  var edgeNode = edge && getObject(edge, NODE);
  if (!edge || !edgeNode) {
    return;
  }

  // Extract the id of the node with the connection that we are adding to.
  var connectionParentID = config.parentID;
  if (!connectionParentID) {
    var edgeSource = getObject(edge, 'source');
    if (edgeSource) {
      connectionParentID = getString(edgeSource, ID);
    }
  }
  require('fbjs/lib/invariant')(connectionParentID, 'writeRelayUpdatePayload(): Cannot insert edge without a configured ' + '`parentID` or a `%s.source.id` field.', config.edgeName);

  var nodeID = getString(edgeNode, ID) || require('./generateClientID')();
  var cursor = edge.cursor || STUB_CURSOR_ID;
  var edgeData = (0, _extends3['default'])({}, edge, {
    cursor: cursor,
    node: (0, _extends3['default'])({}, edgeNode, {
      id: nodeID
    })
  });

  // add the node to every connection for this field
  var connectionIDs = store.getConnectionIDsForField(connectionParentID, config.connectionName);
  if (connectionIDs) {
    connectionIDs.forEach(function (connectionID) {
      return addRangeNode(writer, operation, config, connectionID, nodeID, edgeData);
    });
  }

  if (isOptimisticUpdate) {
    // optimistic updates need to record the generated client ID for
    // a to-be-created node
    require('./RelayMutationTracker').putClientIDForMutation(nodeID, clientMutationID);
  } else {
    // non-optimistic updates check for the existence of a generated client
    // ID (from the above `if` clause) and link the client ID to the actual
    // server ID.
    var clientNodeID = require('./RelayMutationTracker').getClientIDForMutation(clientMutationID);
    if (clientNodeID) {
      require('./RelayMutationTracker').updateClientServerIDMap(clientNodeID, nodeID);
      require('./RelayMutationTracker').deleteClientIDForMutation(clientMutationID);
    }
  }
}

/**
 * Writes the node data for the given field to the store and prepends/appends
 * the node to the given connection.
 */
function addRangeNode(writer, operation, config, connectionID, nodeID, edgeData) {
  var store = writer.getRecordStore();
  var recordWriter = writer.getRecordWriter();
  var filterCalls = store.getRangeFilterCalls(connectionID);
  var rangeBehavior = filterCalls ? require('./getRangeBehavior')(config.rangeBehaviors, filterCalls) : null;

  // no range behavior specified for this combination of filter calls
  if (!rangeBehavior) {
    require('fbjs/lib/warning')(rangeBehavior, 'Using `null` as a rangeBehavior value is deprecated. Use `ignore` to avoid ' + 'refetching a range.');
    return;
  }

  if (rangeBehavior === IGNORE) {
    return;
  }

  var edgeID = require('./generateClientEdgeID')(connectionID, nodeID);
  var path = store.getPathToRecord(connectionID);
  require('fbjs/lib/invariant')(path, 'writeRelayUpdatePayload(): Expected a path for connection record, `%s`.', connectionID);
  path = require('./RelayQueryPath').getPath(path, EDGES_FIELD, edgeID);

  // create the edge record
  writer.createRecordIfMissing(EDGES_FIELD, edgeID, path, edgeData);

  // write data for all `edges` fields
  // TODO #7167718: more efficient mutation/subscription writes
  var hasEdgeField = false;
  var handleNode = function handleNode(node) {
    node.getChildren().forEach(function (child) {
      if (child instanceof require('./RelayQuery').Fragment) {
        handleNode(child);
      } else if (child instanceof require('./RelayQuery').Field && child.getSchemaName() === config.edgeName) {
        hasEdgeField = true;
        if (path) {
          writer.writePayload(child, edgeID, edgeData, path);
        }
      }
    });
  };
  handleNode(operation);

  require('fbjs/lib/invariant')(hasEdgeField, 'writeRelayUpdatePayload(): Expected mutation query to include the ' + 'relevant edge field, `%s`.', config.edgeName);

  // append/prepend the item to the range.
  if (rangeBehavior in require('./GraphQLMutatorConstants').RANGE_OPERATIONS) {
    recordWriter.applyRangeUpdate(connectionID, edgeID, rangeBehavior);
    writer.recordUpdate(connectionID);
  } else {
    console.error('writeRelayUpdatePayload(): invalid range operation `%s`, valid ' + 'options are `%s`, `%s`, `%s`, or `%s`.', rangeBehavior, APPEND, PREPEND, IGNORE, REFETCH);
  }
}

/**
 * Handles the payload for a range edge deletion, which removes the edge from
 * a specified range but does not delete the node for that edge. The config
 * specifies the path within the payload that contains the connection ID.
 */
function handleRangeDelete(writer, payload, config) {
  var store = writer.getRecordStore();

  var recordIDs = null;

  if (Array.isArray(config.deletedIDFieldName)) {
    recordIDs = getIDsFromPath(store, config.deletedIDFieldName, payload);
  } else {
    recordIDs = payload[config.deletedIDFieldName];

    // Coerce numbers to strings for backwards compatibility.
    if (typeof recordIDs === 'number') {
      require('fbjs/lib/warning')(false, 'writeRelayUpdatePayload(): Expected `%s` to be a string, got the ' + 'number `%s`.', config.deletedIDFieldName, recordIDs);
      recordIDs = '' + recordIDs;
    }

    require('fbjs/lib/invariant')(recordIDs == null || !Array.isArray(recordIDs) || typeof recordIDs !== 'string', 'writeRelayUpdatePayload(): Expected `%s` to be an array/string, got `%s`.', config.deletedIDFieldName, (0, _stringify2['default'])(recordIDs));

    if (!Array.isArray(recordIDs)) {
      recordIDs = [recordIDs];
    }
  }

  require('fbjs/lib/invariant')(recordIDs != null, 'writeRelayUpdatePayload(): Missing ID(s) for deleted record at field `%s`.', config.deletedIDFieldName);

  // Extract the id of the node with the connection that we are deleting from.
  var connectionName = config.pathToConnection.pop();
  var connectionParentIDs = getIDsFromPath(store, config.pathToConnection, payload);
  // Restore pathToConnection to its original state
  config.pathToConnection.push(connectionName);
  if (!connectionParentIDs) {
    return;
  }
  var connectionParentID = connectionParentIDs[0];

  var connectionIDs = store.getConnectionIDsForField(connectionParentID, connectionName);
  if (connectionIDs) {
    connectionIDs.forEach(function (connectionID) {
      if (recordIDs) {
        recordIDs.forEach(function (recordID) {
          deleteRangeEdge(writer, connectionID, recordID);
        });
      }
    });
  }
}

/**
 * Removes an edge from a connection without modifying the node data.
 */
function deleteRangeEdge(writer, connectionID, nodeID) {
  var recordWriter = writer.getRecordWriter();
  var edgeID = require('./generateClientEdgeID')(connectionID, nodeID);
  recordWriter.applyRangeUpdate(connectionID, edgeID, REMOVE);

  deleteRecord(writer, edgeID);
  writer.recordUpdate(connectionID);
}

/**
 * Given a payload of data and a path of fields, extracts the `id` of the node(s)
 * specified by the path.
 *
 * Examples:
 * path: ['root', 'field']
 * data: {root: {field: {id: 'xyz'}}}
 *
 * path: ['root', 'field']
 * data: {root: {field: [{id: 'abc'}, {id: 'def'}]}}
 *
 * Returns:
 * ['xyz']
 *
 * ['abc', 'def']
 */
function getIDsFromPath(store, path, payload) {
  // We have a special case for the path for root nodes without ids like
  // ['viewer']. We try to match it up with something in the root call mapping
  // first.
  if (path.length === 1) {
    var rootCallID = store.getDataID(path[0]);
    if (rootCallID) {
      return [rootCallID];
    }
  }

  var payloadItems = payload;
  path.forEach(function (step, idx) {
    if (!payloadItems || Array.isArray(payloadItems)) {
      return;
    }
    if (idx === path.length - 1) {
      payloadItems = getObjectOrArray(payloadItems, step);
    } else {
      payloadItems = getObject(payloadItems, step);
    }
  });

  if (payloadItems) {
    if (!Array.isArray(payloadItems)) {
      payloadItems = [payloadItems];
    }
    return payloadItems.map(function (item) {
      var id = getString(item, ID);
      require('fbjs/lib/invariant')(id != null, 'writeRelayUpdatePayload(): Expected `%s.id` to be a string.', path.join('.'));
      return id;
    });
  }
  return null;
}

function getString(payload, field) {
  var value = payload[field];
  // Coerce numbers to strings for backwards compatibility.
  if (typeof value === 'number') {
    require('fbjs/lib/warning')(false, 'writeRelayUpdatePayload(): Expected `%s` to be a string, got the ' + 'number `%s`.', field, value);
    value = '' + value;
  }
  require('fbjs/lib/invariant')(value == null || typeof value === 'string', 'writeRelayUpdatePayload(): Expected `%s` to be a string, got `%s`.', field, (0, _stringify2['default'])(value));
  return value;
}

function getObject(payload, field) {
  var value = payload[field];
  require('fbjs/lib/invariant')(value == null || typeof value === 'object' && !Array.isArray(value), 'writeRelayUpdatePayload(): Expected `%s` to be an object, got `%s`.', field, (0, _stringify2['default'])(value));
  return value;
}

function getObjectOrArray(payload, field) {
  var value = payload[field];
  require('fbjs/lib/invariant')(value == null || typeof value === 'object', 'writeRelayUpdatePayload(): Expected `%s` to be an object/array, got `%s`.', field, (0, _stringify2['default'])(value));
  return value;
}

module.exports = require('./RelayProfiler').instrument('writeRelayUpdatePayload', writeRelayUpdatePayload);