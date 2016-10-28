/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayMutationQuery
 * 
 */

'use strict';

var REFETCH = require('./GraphQLMutatorConstants').REFETCH;

// This should probably use disjoint unions.

var CLIENT_MUTATION_ID = require('./RelayConnectionInterface').CLIENT_MUTATION_ID;

var ANY_TYPE = require('./RelayNodeInterface').ANY_TYPE;

var ID = require('./RelayNodeInterface').ID;

var TYPENAME = require('./RelayNodeInterface').TYPENAME;

/**
 * @internal
 *
 * Constructs query fragments that are sent with mutations, which should ensure
 * that any records changed as a result of mutations are brought up-to-date.
 *
 * The fragments are a minimal subset created by intersecting the "fat query"
 * (fields that a mutation declares may have changed) with the "tracked query"
 * (fields representing data previously queried and written into the store).
 */


var RelayMutationQuery = {
  /**
   * Accepts a mapping from field names to data IDs. The field names must exist
   * as top-level fields in the fat query. These top-level fields are used to
   * re-fetch any data that has changed for records identified by the data IDs.
   *
   * The supplied mapping may contain multiple field names. In addition, each
   * field name may map to an array of data IDs if the field is plural.
   */

  buildFragmentForFields: function buildFragmentForFields(_ref) {
    var fatQuery = _ref.fatQuery;
    var fieldIDs = _ref.fieldIDs;
    var tracker = _ref.tracker;

    var mutatedFields = [];
    require('fbjs/lib/forEachObject')(fieldIDs, function (dataIDOrIDs, fieldName) {
      var fatField = getFieldFromFatQuery(fatQuery, fieldName);
      var dataIDs = [].concat(dataIDOrIDs);
      var trackedChildren = [];
      dataIDs.forEach(function (dataID) {
        trackedChildren.push.apply(trackedChildren, tracker.getTrackedChildrenForID(dataID));
      });
      var trackedField = fatField.clone(trackedChildren);
      var mutationField = null;
      if (trackedField) {
        mutationField = require('./intersectRelayQuery')(trackedField, fatField);
        if (mutationField) {
          mutatedFields.push(mutationField);
        }
      }
      /* eslint-disable no-console */
      if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
        console.groupCollapsed('Building fragment for `' + fieldName + '`');
        console.log(require('./RelayNodeInterface').ID + ': ', dataIDOrIDs);

        var RelayMutationDebugPrinter = require('./RelayMutationDebugPrinter');
        RelayMutationDebugPrinter.printMutation(trackedField && buildMutationFragment(fatQuery, [trackedField]), 'Tracked Fragment');
        RelayMutationDebugPrinter.printMutation(buildMutationFragment(fatQuery, [fatField]), 'Fat Fragment');
        RelayMutationDebugPrinter.printMutation(mutationField && buildMutationFragment(fatQuery, [mutationField]), 'Intersected Fragment');
        console.groupEnd();
      }
      /* eslint-enable no-console */
    });
    return buildMutationFragment(fatQuery, mutatedFields);
  },


  /**
   * Creates a fragment used to update any data as a result of a mutation that
   * deletes an edge from a connection. The primary difference between this and
   * `createForFields` is whether or not the connection edges are re-fetched.
   *
   * `connectionName`
   *   Name of the connection field from which the edge is being deleted.
   *
   * `parentID`
   *   ID of the parent record containing the connection which may have metadata
   *   that needs to be re-fetched.
   *
   * `parentName`
   *   Name of the top-level field in the fat query that corresponds to the
   *   parent record.
   */
  buildFragmentForEdgeDeletion: function buildFragmentForEdgeDeletion(_ref2) {
    var fatQuery = _ref2.fatQuery;
    var connectionName = _ref2.connectionName;
    var parentID = _ref2.parentID;
    var parentName = _ref2.parentName;
    var tracker = _ref2.tracker;

    var fatParent = getFieldFromFatQuery(fatQuery, parentName);

    // The connection may not be explicit in the fat query, but if it is, we
    // try to validate it.
    getConnectionAndValidate(fatParent, parentName, connectionName);

    var mutatedFields = [];
    var trackedParent = fatParent.clone(tracker.getTrackedChildrenForID(parentID));
    if (trackedParent) {
      var filterUnterminatedRange = function filterUnterminatedRange(node) {
        return node.getSchemaName() === connectionName;
      };
      var mutatedField = require('./intersectRelayQuery')(trackedParent, fatParent, filterUnterminatedRange);
      if (mutatedField) {
        // If we skipped validation above, we get a second chance here.
        getConnectionAndValidate(mutatedField, parentName, connectionName);

        mutatedFields.push(mutatedField);
      }
    }
    return buildMutationFragment(fatQuery, mutatedFields);
  },


  /**
   * Creates a fragment used to fetch data necessary to insert a new edge into
   * an existing connection.
   *
   * `connectionName`
   *   Name of the connection field into which the edge is being inserted.
   *
   * `parentID`
   *   ID of the parent record containing the connection which may have metadata
   *   that needs to be re-fetched.
   *
   * `edgeName`
   *   Name of the top-level field in the fat query that corresponds to the
   *   newly inserted edge.
   *
   * `parentName`
   *   Name of the top-level field in the fat query that corresponds to the
   *   parent record. If not supplied, metadata on the parent record and any
   *   connections without entries in `rangeBehaviors` will not be updated.
   */
  buildFragmentForEdgeInsertion: function buildFragmentForEdgeInsertion(_ref3) {
    var fatQuery = _ref3.fatQuery;
    var connectionName = _ref3.connectionName;
    var parentID = _ref3.parentID;
    var edgeName = _ref3.edgeName;
    var parentName = _ref3.parentName;
    var rangeBehaviors = _ref3.rangeBehaviors;
    var tracker = _ref3.tracker;

    var mutatedFields = [];
    var keysWithoutRangeBehavior = {};
    var trackedChildren = tracker.getTrackedChildrenForID(parentID);
    var trackedConnections = [];
    trackedChildren.forEach(function (trackedChild) {
      trackedConnections.push.apply(trackedConnections, findDescendantFields(trackedChild, connectionName));
    });

    if (trackedConnections.length) {
      (function () {
        // If the first instance of the connection passes validation, all will.
        validateConnection(parentName, connectionName, trackedConnections[0]);

        var mutatedEdgeFields = [];
        trackedConnections.forEach(function (trackedConnection) {
          var trackedEdges = findDescendantFields(trackedConnection, 'edges');
          if (!trackedEdges.length) {
            return;
          }

          var callsWithValues = trackedConnection.getRangeBehaviorCalls();
          var rangeBehavior = require('./getRangeBehavior')(rangeBehaviors, callsWithValues);
          /* eslint-disable no-console */
          if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
            var serializeRelayQueryCall = require('./serializeRelayQueryCall');
            var serializedCalls = callsWithValues.map(serializeRelayQueryCall).sort().join('');
            console.log(serializedCalls + ': ' + (rangeBehavior || ''));
          }
          /* eslint-enable no-console */
          if (rangeBehavior && rangeBehavior !== REFETCH) {
            // Include edges from all connections that exist in `rangeBehaviors`.
            // This may add duplicates, but they will eventually be flattened.
            trackedEdges.forEach(function (trackedEdge) {
              mutatedEdgeFields.push.apply(mutatedEdgeFields, trackedEdge.getChildren());
            });
          } else {
            // If the connection is not in `rangeBehaviors` or we have explicitly
            // set the behavior to `refetch`, re-fetch it.
            require('fbjs/lib/warning')(rangeBehavior === REFETCH, 'RelayMutation: The connection `%s` on the mutation field `%s` ' + 'that corresponds to the ID `%s` did not match any of the ' + '`rangeBehaviors` specified in your RANGE_ADD config. This means ' + 'that the entire connection will be refetched. Configure a range ' + 'behavior for this mutation in order to fetch only the new edge ' + 'and to enable optimistic mutations or use `refetch` to squelch ' + 'this warning.', trackedConnection.getStorageKey(), parentName, parentID);
            keysWithoutRangeBehavior[trackedConnection.getShallowHash()] = true;
          }
        });
        if (mutatedEdgeFields.length) {
          mutatedFields.push(buildEdgeField(parentID, edgeName, mutatedEdgeFields));
        }
      })();
    }

    if (parentName != null) {
      var fatParent = getFieldFromFatQuery(fatQuery, parentName);

      // The connection may not be explicit in the fat query, but if it is, we
      // try to validate it.
      getConnectionAndValidate(fatParent, parentName, connectionName);

      var trackedParent = fatParent.clone(trackedChildren);
      if (trackedParent) {
        var filterUnterminatedRange = function filterUnterminatedRange(node) {
          return node.getSchemaName() === connectionName && !keysWithoutRangeBehavior.hasOwnProperty(node.getShallowHash());
        };
        var mutatedParent = require('./intersectRelayQuery')(trackedParent, fatParent, filterUnterminatedRange);
        if (mutatedParent) {
          mutatedFields.push(mutatedParent);
        }
      }
    }

    return buildMutationFragment(fatQuery, mutatedFields);
  },


  /**
   * Creates a fragment used to fetch the given optimistic response.
   */
  buildFragmentForOptimisticUpdate: function buildFragmentForOptimisticUpdate(_ref4) {
    var response = _ref4.response;
    var fatQuery = _ref4.fatQuery;

    // Silences RelayQueryNode being incompatible with sub-class RelayQueryField
    // A detailed error description is available in #7635477
    var mutatedFields = require('./RelayOptimisticMutationUtils').inferRelayFieldsFromData(response);
    return buildMutationFragment(fatQuery, mutatedFields);
  },


  /**
   * Creates a RelayQuery.Mutation used to fetch the given optimistic response.
   */
  buildQueryForOptimisticUpdate: function buildQueryForOptimisticUpdate(_ref5) {
    var fatQuery = _ref5.fatQuery;
    var mutation = _ref5.mutation;
    var response = _ref5.response;

    var children = [require('fbjs/lib/nullthrows')(RelayMutationQuery.buildFragmentForOptimisticUpdate({
      response: response,
      fatQuery: fatQuery
    }))];
    return require('./RelayQuery').Mutation.build('OptimisticQuery', fatQuery.getType(), mutation.calls[0].name, null, children, mutation.metadata);
  },


  /**
   * Creates a RelayQuery.Mutation for the given config. See type
   * `MutationConfig` and the `buildFragmentForEdgeInsertion`,
   * `buildFragmentForEdgeDeletion` and `buildFragmentForFields` methods above
   * for possible configs.
   */
  buildQuery: function buildQuery(_ref6) {
    var configs = _ref6.configs;
    var fatQuery = _ref6.fatQuery;
    var input = _ref6.input;
    var mutationName = _ref6.mutationName;
    var mutation = _ref6.mutation;
    var tracker = _ref6.tracker;

    var children = [require('./RelayQuery').Field.build({
      fieldName: CLIENT_MUTATION_ID,
      type: 'String',
      metadata: { isRequisite: true }
    })];
    /* eslint-disable no-console */
    if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
      console.groupCollapsed('Mutation Configs');
    }
    /* eslint-enable no-console */
    configs.forEach(function (config) {
      switch (config.type) {
        case require('./RelayMutationType').REQUIRED_CHILDREN:
          var newChildren = config.children.map(function (child) {
            return require('./RelayQuery').Fragment.create(child, require('./RelayMetaRoute').get('$buildQuery'), {});
          });
          children = children.concat(newChildren);
          /* eslint-disable no-console */
          if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
            (function () {
              var RelayMutationDebugPrinter = require('./RelayMutationDebugPrinter');
              console.groupCollapsed('REQUIRED_CHILDREN');
              newChildren.forEach(function (child, index) {
                console.groupCollapsed(index);
                RelayMutationDebugPrinter.printMutation(child);
                console.groupEnd();
              });
              console.groupEnd();
            })();
          }
          /* eslint-enable no-console */
          break;

        case require('./RelayMutationType').RANGE_ADD:
          /* eslint-disable no-console */
          if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
            console.groupCollapsed('RANGE_ADD');
          }
          /* eslint-enable no-console */
          children.push(RelayMutationQuery.buildFragmentForEdgeInsertion({
            connectionName: config.connectionName,
            edgeName: config.edgeName,
            fatQuery: fatQuery,
            parentID: config.parentID,
            parentName: config.parentName,
            rangeBehaviors: sanitizeRangeBehaviors(config.rangeBehaviors),
            tracker: tracker
          }));
          /* eslint-disable no-console */
          if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
            console.groupEnd();
          }
          /* eslint-enable no-console */
          break;

        case require('./RelayMutationType').RANGE_DELETE:
        case require('./RelayMutationType').NODE_DELETE:
          var edgeDeletion = RelayMutationQuery.buildFragmentForEdgeDeletion({
            connectionName: config.connectionName,
            fatQuery: fatQuery,
            parentID: config.parentID,
            parentName: config.parentName,
            tracker: tracker
          });
          children.push(edgeDeletion);
          var deletedIDFieldName = Array.isArray(config.deletedIDFieldName) ? config.deletedIDFieldName.concat(ID) : [config.deletedIDFieldName];
          var nodeDeletion = buildFragmentForDeletedConnectionNodeID(deletedIDFieldName, fatQuery);
          children.push(nodeDeletion);
          /* eslint-disable no-console */
          if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
            var configType = config === require('./RelayMutationType').RANGE_DELETE ? 'RANGE_DELETE' : 'NODE_DELETE';
            console.groupCollapsed(configType);

            var _RelayMutationDebugPrinter = require('./RelayMutationDebugPrinter');
            _RelayMutationDebugPrinter.printMutation(edgeDeletion, 'Edge Fragment');
            _RelayMutationDebugPrinter.printMutation(nodeDeletion, 'Node Fragment');

            console.groupEnd();
          }
          /* eslint-enable no-console */
          break;

        case require('./RelayMutationType').FIELDS_CHANGE:
          /* eslint-disable no-console */
          if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
            console.groupCollapsed('FIELDS_CHANGE');
          }
          /* eslint-enable no-console */
          children.push(RelayMutationQuery.buildFragmentForFields({
            fatQuery: fatQuery,
            fieldIDs: config.fieldIDs,
            tracker: tracker
          }));
          /* eslint-disable no-console */
          if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
            console.groupEnd();
          }
          /* eslint-enable no-console */
          break;

        default:
          require('fbjs/lib/invariant')(false, 'RelayMutationQuery: Unrecognized config key `%s` for `%s`.', config.type, mutationName);
      }
    });
    /* eslint-disable no-console */
    if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
      console.groupEnd();
    }
    /* eslint-enable no-console */
    return require('./RelayQuery').Mutation.build(mutationName, fatQuery.getType(), mutation.calls[0].name, input, children.filter(function (child) {
      return child != null;
    }), mutation.metadata);
  }
};

function getFieldFromFatQuery(fatQuery, fieldName) {
  var field = fatQuery.getFieldByStorageKey(fieldName);
  require('fbjs/lib/invariant')(field, 'RelayMutationQuery: Invalid field name on fat query, `%s`.', fieldName);
  return field;
}

function buildMutationFragment(fatQuery, fields) {
  var fragment = require('./RelayQuery').Fragment.build('MutationQuery', fatQuery.getType(), fields);

  require('fbjs/lib/invariant')(fragment instanceof require('./RelayQuery').Fragment, 'RelayMutationQuery: Expected a fragment.');
  return fragment;
}

function buildFragmentForDeletedConnectionNodeID(fieldNames, fatQuery) {
  require('fbjs/lib/invariant')(fieldNames.length > 0, 'RelayMutationQuery: Invalid deleted node id name.');
  var field = require('./RelayQuery').Field.build({
    fieldName: fieldNames[fieldNames.length - 1],
    type: 'String'
  });
  for (var ii = fieldNames.length - 2; ii >= 0; ii--) {
    field = require('./RelayQuery').Field.build({
      fieldName: fieldNames[ii],
      type: ANY_TYPE,
      children: [field],
      metadata: {
        canHaveSubselections: true
      }
    });
  }
  return buildMutationFragment(fatQuery, [field]);
}

function buildEdgeField(parentID, edgeName, edgeFields) {
  var fields = [require('./RelayQuery').Field.build({
    fieldName: 'cursor',
    type: 'String'
  }), require('./RelayQuery').Field.build({
    fieldName: TYPENAME,
    type: 'String'
  })];
  if (require('./RelayConnectionInterface').EDGES_HAVE_SOURCE_FIELD && !require('./RelayRecord').isClientID(parentID)) {
    fields.push(require('./RelayQuery').Field.build({
      children: [require('./RelayQuery').Field.build({
        fieldName: ID,
        type: 'String'
      }), require('./RelayQuery').Field.build({
        fieldName: TYPENAME,
        type: 'String'
      })],
      fieldName: 'source',
      metadata: { canHaveSubselections: true },
      type: ANY_TYPE
    }));
  }
  fields.push.apply(fields, edgeFields);
  var edgeField = require('./flattenRelayQuery')(require('./RelayQuery').Field.build({
    children: fields,
    fieldName: edgeName,
    metadata: { canHaveSubselections: true },
    type: ANY_TYPE
  }));
  require('fbjs/lib/invariant')(edgeField instanceof require('./RelayQuery').Field, 'RelayMutationQuery: Expected a field.');
  return edgeField;
}

function sanitizeRangeBehaviors(rangeBehaviors) {
  // Prior to 0.4.1 you would have to specify the args in your range behaviors
  // in the same order they appeared in your query. From 0.4.1 onward, args in a
  // range behavior key must be in alphabetical order.

  // No need to sanitize if defined as a function
  if (typeof rangeBehaviors === 'function') {
    return rangeBehaviors;
  }

  var unsortedKeys = void 0;
  require('fbjs/lib/forEachObject')(rangeBehaviors, function (value, key) {
    if (key !== '') {
      var keyParts = key
      // Remove the last parenthesis
      .slice(0, -1)
      // Slice on unescaped parentheses followed immediately by a `.`
      .split(/\)\./);
      var sortedKey = keyParts.sort().join(').') + (keyParts.length ? ')' : '');
      if (sortedKey !== key) {
        unsortedKeys = unsortedKeys || [];
        unsortedKeys.push(key);
      }
    }
  });
  if (unsortedKeys) {
    require('fbjs/lib/invariant')(false, 'RelayMutation: To define a range behavior key without sorting ' + 'the arguments alphabetically is disallowed as of Relay 0.5.1. Please ' + 'sort the argument names of the range behavior key%s `%s`%s.', unsortedKeys.length === 1 ? '' : 's', unsortedKeys.length === 1 ? unsortedKeys[0] : unsortedKeys.length === 2 ? unsortedKeys[0] + '` and `' + unsortedKeys[1] : unsortedKeys.slice(0, -1).join('`, `'), unsortedKeys.length > 2 ? ', and `' + unsortedKeys.slice(-1) + '`' : '');
  }
  return rangeBehaviors;
}

/**
 * Confirms that the `connection` field extracted from the fat query at
 * `parentName` -> `connectionName` is actually a connection.
 */
function validateConnection(parentName, connectionName, connection) {
  require('fbjs/lib/invariant')(connection.isConnection(), 'RelayMutationQuery: Expected field `%s`%s to be a connection.', connectionName, parentName ? ' on `' + parentName + '`' : '');
}

/**
 * Convenience wrapper around validateConnection that gracefully attempts to
 * extract the connection identified by `connectionName` from the `parentField`.
 * If the connection isn't present (because it wasn't in the fat query or
 * because it didn't survive query intersection), validation is skipped.
 */
function getConnectionAndValidate(parentField, parentName, connectionName) {
  var connections = findDescendantFields(parentField, connectionName);
  if (connections.length) {
    // If the first instance of the connection passes validation, all will.
    validateConnection(parentName, connectionName, connections[0]);
  }
}

/**
 * Finds all direct and indirect child fields of `node` with the given
 * field name.
 */
function findDescendantFields(rootNode, fieldName) {
  var fields = [];
  function traverse(node) {
    if (node instanceof require('./RelayQuery').Field) {
      if (node.getSchemaName() === fieldName) {
        fields.push(node);
        return;
      }
    }
    if (node === rootNode || node instanceof require('./RelayQuery').Fragment) {
      // Search fragments and the root node for matching fields, but skip
      // descendant non-matching fields.
      node.getChildren().forEach(function (child) {
        return traverse(child);
      });
    }
  }
  traverse(rootNode);
  return fields;
}

module.exports = RelayMutationQuery;