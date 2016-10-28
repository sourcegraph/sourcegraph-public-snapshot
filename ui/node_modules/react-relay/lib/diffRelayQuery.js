/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule diffRelayQuery
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var ID = require('./RelayNodeInterface').ID;

var ID_TYPE = require('./RelayNodeInterface').ID_TYPE;

var NODE_TYPE = require('./RelayNodeInterface').NODE_TYPE;

var TYPENAME = require('./RelayNodeInterface').TYPENAME;

var EDGES = require('./RelayConnectionInterface').EDGES;

var NODE = require('./RelayConnectionInterface').NODE;

var PAGE_INFO = require('./RelayConnectionInterface').PAGE_INFO;

var idField = require('./RelayQuery').Field.build({
  fieldName: ID,
  metadata: {
    isRequisite: true
  },
  type: 'String'
});
var typeField = require('./RelayQuery').Field.build({
  fieldName: TYPENAME,
  metadata: {
    isRequisite: true
  },
  type: 'String'
});
var nodeWithID = require('./RelayQuery').Field.build({
  fieldName: require('./RelayNodeInterface').NODE,
  children: [idField, typeField],
  metadata: {
    canHaveSubselections: true
  },
  type: NODE_TYPE
});

/**
 * @internal
 *
 * Computes the difference between the data requested in `root` and the data
 * available in `store`. It returns a minimal set of queries that will fulfill
 * the difference, or an empty array if the query can be resolved locally.
 */
function diffRelayQuery(root, store, queryTracker) {
  var path = require('./RelayQueryPath').create(root);
  var queries = [];

  var visitor = new RelayDiffQueryBuilder(store, queryTracker);
  var rootIdentifyingArg = root.getIdentifyingArg();
  var rootIdentifyingArgValue = rootIdentifyingArg && rootIdentifyingArg.value || null;
  var isPluralCall = Array.isArray(rootIdentifyingArgValue) && rootIdentifyingArgValue.length > 1;
  var metadata = void 0;
  if (rootIdentifyingArg != null) {
    metadata = {
      identifyingArgName: rootIdentifyingArg.name,
      identifyingArgType: rootIdentifyingArg.type != null ? rootIdentifyingArg.type : ID_TYPE,
      isAbstract: root.isAbstract(),
      isDeferred: false,
      isPlural: false
    };
  }
  var fieldName = root.getFieldName();
  var storageKey = root.getStorageKey();
  require('./forEachRootCallArg')(root, function (_ref) {
    var identifyingArgValue = _ref.identifyingArgValue;
    var identifyingArgKey = _ref.identifyingArgKey;

    var nodeRoot = void 0;
    if (isPluralCall) {
      require('fbjs/lib/invariant')(identifyingArgValue != null, 'diffRelayQuery(): Unexpected null or undefined value in root call ' + 'argument array for query, `%s(...).', fieldName);
      nodeRoot = require('./RelayQuery').Root.build(root.getName(), fieldName, [identifyingArgValue], root.getChildren(), metadata, root.getType());
    } else {
      // Reuse `root` if it only maps to one result.
      nodeRoot = root;
    }

    // The whole query must be fetched if the root dataID is unknown.
    var dataID = store.getDataID(storageKey, identifyingArgKey);
    if (dataID == null) {
      queries.push(nodeRoot);
      return;
    }

    // Diff the current dataID
    var scope = makeScope(dataID);
    var diffOutput = visitor.visit(nodeRoot, path, scope);
    var diffNode = diffOutput ? diffOutput.diffNode : null;
    if (diffNode) {
      require('fbjs/lib/invariant')(diffNode instanceof require('./RelayQuery').Root, 'diffRelayQuery(): Expected result to be a root query.');
      queries.push(diffNode);
    }
  });
  return queries.concat(visitor.getSplitQueries());
}

/**
 * @internal
 *
 * A transform for (node + store) -> (diff + tracked queries). It is analagous
 * to `RelayQueryTransform` with the main differences as follows:
 * - there is no `state` (which allowed for passing data up and down the tree).
 * - data is passed down via `scope`, which flows from a parent field down
 *   through intermediary fragments to the nearest child field.
 * - data is passed up via the return type `{diffNode, trackedNode}`, where:
 *   - `diffNode`: subset of the input that could not diffed out
 *   - `trackedNode`: subset of the input that must be tracked
 *
 * The provided `queryTracker`, if any, is updated whenever the traversal of a
 * node results in a `trackedNode` being created. New top-level queries are not
 * returned up the tree, and instead are available via `getSplitQueries()`.
 *
 * @note If no `queryTracker` is provided, all tracking-related functionality is
 * skipped.
 */

var RelayDiffQueryBuilder = function () {
  function RelayDiffQueryBuilder(store, queryTracker) {
    (0, _classCallCheck3['default'])(this, RelayDiffQueryBuilder);

    this._store = store;
    this._splitQueries = [];
    this._queryTracker = queryTracker;
  }

  RelayDiffQueryBuilder.prototype.splitQuery = function splitQuery(root) {
    this._splitQueries.push(root);
  };

  RelayDiffQueryBuilder.prototype.getSplitQueries = function getSplitQueries() {
    return this._splitQueries;
  };

  RelayDiffQueryBuilder.prototype.visit = function visit(node, path, scope) {
    if (node instanceof require('./RelayQuery').Field) {
      return this.visitField(node, path, scope);
    } else if (node instanceof require('./RelayQuery').Fragment) {
      return this.visitFragment(node, path, scope);
    } else if (node instanceof require('./RelayQuery').Root) {
      return this.visitRoot(node, path, scope);
    }
  };

  RelayDiffQueryBuilder.prototype.visitRoot = function visitRoot(node, path, scope) {
    return this.traverse(node, path, scope);
  };

  RelayDiffQueryBuilder.prototype.visitFragment = function visitFragment(node, path, scope) {
    return this.traverse(node, path, scope);
  };

  /**
   * Diffs the field conditionally based on the `scope` from the nearest
   * ancestor field.
   */


  RelayDiffQueryBuilder.prototype.visitField = function visitField(node, path, _ref2) {
    var connectionField = _ref2.connectionField;
    var dataID = _ref2.dataID;
    var edgeID = _ref2.edgeID;
    var rangeInfo = _ref2.rangeInfo;

    // special case when inside a connection traversal
    if (connectionField && rangeInfo) {
      if (edgeID) {
        // When traversing a specific connection edge only look at `edges`
        if (node.getSchemaName() === EDGES) {
          return this.diffConnectionEdge(connectionField, node, // edge field
          require('./RelayQueryPath').getPath(path, node, edgeID), edgeID, rangeInfo);
        } else {
          return null;
        }
      } else {
        // When traversing connection metadata fields, edges/page_info are
        // only kept if there are range extension calls. Other fields fall
        // through to regular diffing.
        if (node.getSchemaName() === EDGES || node.getSchemaName() === PAGE_INFO) {
          return rangeInfo.diffCalls.length > 0 ? {
            diffNode: node,
            trackedNode: null
          } : null;
        }
      }
    }

    // default field diffing algorithm
    if (!node.canHaveSubselections()) {
      return this.diffScalar(node, dataID);
    } else if (node.isGenerated()) {
      return {
        diffNode: node,
        trackedNode: null
      };
    } else if (node.isConnection()) {
      return this.diffConnection(node, path, dataID);
    } else if (node.isPlural()) {
      return this.diffPluralLink(node, path, dataID);
    } else {
      return this.diffLink(node, path, dataID);
    }
  };

  /**
   * Visit all the children of the given `node` and merge their results.
   */


  RelayDiffQueryBuilder.prototype.traverse = function traverse(node, path, scope) {
    var _this = this;

    var diffNode = void 0;
    var diffChildren = void 0;
    var trackedNode = void 0;
    var trackedChildren = void 0;
    var hasDiffField = false;
    var hasTrackedField = false;

    node.getChildren().forEach(function (child) {
      if (child instanceof require('./RelayQuery').Field) {
        var diffOutput = _this.visitField(child, path, scope);
        var diffChild = diffOutput ? diffOutput.diffNode : null;
        var trackedChild = diffOutput && _this._queryTracker ? diffOutput.trackedNode : null;

        // Diff uses child nodes and keeps requisite fields
        if (diffChild) {
          diffChildren = diffChildren || [];
          diffChildren.push(diffChild);
          hasDiffField = hasDiffField || !diffChild.isGenerated();
        } else if (child.isRequisite() && !scope.rangeInfo) {
          // The presence of `rangeInfo` indicates that we are traversing
          // connection metadata fields, in which case `visitField` will ensure
          // that `edges` and `page_info` are kept when necessary. The requisite
          // check alone could cause these fields to be added back when not
          // needed.
          //
          // Example: `friends.first(3) {count, edges {...}, page_info {...} }
          // If all `edges` were fetched but `count` is unfetched, the diff
          // should be `friends.first(3) {count}` and not include `page_info`.
          diffChildren = diffChildren || [];
          diffChildren.push(child);
        }
        if (_this._queryTracker) {
          // Tracker uses tracked children and keeps requisite fields
          if (trackedChild) {
            trackedChildren = trackedChildren || [];
            trackedChildren.push(trackedChild);
            hasTrackedField = hasTrackedField || !trackedChild.isGenerated();
          } else if (child.isRequisite()) {
            trackedChildren = trackedChildren || [];
            trackedChildren.push(child);
          }
        }
      } else if (child instanceof require('./RelayQuery').Fragment) {
        var isCompatibleType = require('./isCompatibleRelayFragmentType')(child, _this._store.getType(scope.dataID));
        if (isCompatibleType) {
          if (child.isTrackingEnabled()) {
            var hash = child.getCompositeHash();
            if (_this._store.hasFragmentData(scope.dataID, hash)) {
              return {
                diffNode: null,
                trackedNode: null
              };
            }
          }

          var _diffOutput = _this.traverse(child, path, scope);
          var _diffChild = _diffOutput ? _diffOutput.diffNode : null;
          var _trackedChild = _diffOutput ? _diffOutput.trackedNode : null;

          if (_diffChild) {
            diffChildren = diffChildren || [];
            diffChildren.push(_diffChild);
            hasDiffField = true;
          }
          if (_trackedChild) {
            trackedChildren = trackedChildren || [];
            trackedChildren.push(_trackedChild);
            hasTrackedField = true;
          }
        } else {
          // Non-matching fragment types are similar to requisite fields:
          // they don't need to be diffed against and should only be included
          // if something *else* is missing from the node.
          diffChildren = diffChildren || [];
          diffChildren.push(child);
        }
      }
    });

    // Only return diff/tracked node if there are non-generated fields
    if (diffChildren && hasDiffField) {
      diffNode = node.clone(diffChildren);
    }
    if (trackedChildren && hasTrackedField) {
      trackedNode = node.clone(trackedChildren);
    }
    // Record tracked nodes. Fragments can be skipped because these will
    // always be composed into, and therefore tracked by, their nearest
    // non-fragment parent.
    if (this._queryTracker && trackedNode && !(trackedNode instanceof require('./RelayQuery').Fragment)) {
      this._queryTracker.trackNodeForID(trackedNode, scope.dataID);
    }

    return {
      diffNode: diffNode,
      trackedNode: trackedNode
    };
  };

  /**
   * Diff a scalar field such as `name` or `id`.
   */


  RelayDiffQueryBuilder.prototype.diffScalar = function diffScalar(field, dataID) {
    if (this._store.getField(dataID, field.getStorageKey()) === undefined) {
      return {
        diffNode: field,
        trackedNode: null
      };
    }
    return null;
  };

  /**
   * Diff a field-of-fields such as `profile_picture {...}`. Returns early if
   * the field has not been fetched, otherwise the result of traversal.
   */


  RelayDiffQueryBuilder.prototype.diffLink = function diffLink(field, path, dataID) {
    var nextDataID = this._store.getLinkedRecordID(dataID, field.getStorageKey());
    if (nextDataID === undefined) {
      return {
        diffNode: field,
        trackedNode: null
      };
    }
    if (nextDataID === null) {
      return {
        diffNode: null,
        trackedNode: this._queryTracker ? field : null
      };
    }

    return this.traverse(field, require('./RelayQueryPath').getPath(path, field, nextDataID), makeScope(nextDataID));
  };

  /**
   * Diffs a non-connection plural field against each of the fetched items.
   * Note that scalar plural fields are handled by `_diffScalar`.
   */


  RelayDiffQueryBuilder.prototype.diffPluralLink = function diffPluralLink(field, path, dataID) {
    var _this2 = this;

    var linkedIDs = this._store.getLinkedRecordIDs(dataID, field.getStorageKey());
    if (linkedIDs === undefined) {
      // not fetched
      return {
        diffNode: field,
        trackedNode: null
      };
    } else if (linkedIDs === null || linkedIDs.length === 0) {
      // Don't fetch if array is null or empty, but still track the fragment
      return {
        diffNode: null,
        trackedNode: this._queryTracker ? field : null
      };
    } else if (field.getInferredRootCallName() === NODE) {
      var _ret = function () {
        // The items in this array are fetchable and may have been filled in
        // from other sources, so check them all. For example, `Story{actors}`
        // is an array (but not a range), and the Actors in that array likely
        // had data fetched for them elsewhere (like `viewer(){actor}`).
        var hasSplitQueries = false;
        linkedIDs.forEach(function (itemID) {
          var itemState = _this2.traverse(field, require('./RelayQueryPath').getPath(path, field, itemID), makeScope(itemID));
          if (itemState) {
            // If any child was tracked then `field` will also be tracked
            hasSplitQueries = hasSplitQueries || !!itemState.trackedNode || !!itemState.diffNode;
            // split diff nodes into root queries
            if (itemState.diffNode) {
              _this2.splitQuery(buildRoot(itemID, itemState.diffNode.getChildren(), require('./RelayQueryPath').getName(path), field.getType()));
            }
          }
        });
        // if sub-queries are split then this *entire* field will be tracked,
        // therefore we don't need to merge the `trackedNode` from each item
        if (hasSplitQueries) {
          return {
            v: {
              diffNode: null,
              trackedNode: _this2._queryTracker ? field : null
            }
          };
        }
      }();

      if (typeof _ret === "object") return _ret.v;
    } else {
      var _ret2 = function () {
        // The items in this array are not fetchable by ID, so nothing else could
        // have fetched additional data for individual items. If any item in this
        // list is missing data, refetch the whole field.

        var atLeastOneItemHasMissingData = false;
        var atLeastOneItemHasTrackedData = false;

        linkedIDs.some(function (itemID) {
          var itemState = _this2.traverse(field, require('./RelayQueryPath').getPath(path, field, itemID), makeScope(itemID));
          if (itemState && itemState.diffNode) {
            atLeastOneItemHasMissingData = true;
          }
          if (itemState && itemState.trackedNode) {
            atLeastOneItemHasTrackedData = true;
          }
          // Exit early if possible
          return atLeastOneItemHasMissingData && atLeastOneItemHasTrackedData;
        });

        if (atLeastOneItemHasMissingData || atLeastOneItemHasTrackedData) {
          return {
            v: {
              diffNode: atLeastOneItemHasMissingData ? field : null,
              trackedNode: atLeastOneItemHasTrackedData ? field : null
            }
          };
        }
      }();

      if (typeof _ret2 === "object") return _ret2.v;
    }
    return null;
  };

  /**
   * Diff a connection field such as `news_feed.first(3)`. Returns early if
   * the range has not been fetched or the entire range has already been
   * fetched. Otherwise the diff output is a clone of `field` with updated
   * after/first and before/last calls.
   */


  RelayDiffQueryBuilder.prototype.diffConnection = function diffConnection(field, path, dataID) {
    var _this3 = this;

    var store = this._store;
    var connectionID = store.getLinkedRecordID(dataID, field.getStorageKey());
    var rangeInfo = store.getRangeMetadata(connectionID, field.getCallsWithValues());
    // Keep the field if the connection is unfetched
    if (connectionID === undefined) {
      return {
        diffNode: field,
        trackedNode: null
      };
    }
    // Don't fetch if connection is null, but continue to track the fragment if
    // appropriate.
    if (connectionID === null) {
      return this._queryTracker ? {
        diffNode: null,
        trackedNode: field
      } : null;
    }
    // If metadata fields but not edges are fetched, diff as a normal field.
    // In practice, `rangeInfo` is `undefined` if unfetched, `null` if the
    // connection was deleted (in which case `connectionID` is null too).
    if (rangeInfo == null) {
      return this.traverse(field, require('./RelayQueryPath').getPath(path, field, connectionID), makeScope(connectionID));
    }
    var diffCalls = rangeInfo.diffCalls;
    var filteredEdges = rangeInfo.filteredEdges;

    // check existing edges for missing fields

    var hasSplitQueries = false;
    filteredEdges.forEach(function (edge) {
      var scope = {
        connectionField: field,
        dataID: connectionID,
        edgeID: edge.edgeID,
        rangeInfo: rangeInfo
      };
      var diffOutput = _this3.traverse(field, require('./RelayQueryPath').getPath(path, field, edge.edgeID), scope);
      // If any edges were missing data (resulting in a split query),
      // then the entire original connection field must be tracked.
      if (diffOutput) {
        hasSplitQueries = hasSplitQueries || !!diffOutput.trackedNode;
      }
    });

    // Scope has null `edgeID` to skip looking at `edges` fields.
    var scope = {
      connectionField: field,
      dataID: connectionID,
      edgeID: null,
      rangeInfo: rangeInfo
    };
    // diff non-`edges` fields such as `count`
    var diffOutput = this.traverse(field, require('./RelayQueryPath').getPath(path, field, connectionID), scope);
    var diffNode = diffOutput ? diffOutput.diffNode : null;
    var trackedNode = diffOutput ? diffOutput.trackedNode : null;
    if (diffCalls.length && diffNode instanceof require('./RelayQuery').Field) {
      diffNode = diffNode.cloneFieldWithCalls(diffNode.getChildren(), diffCalls);
    }
    // if a sub-query was split, then we must track the entire field, which will
    // be a superset of the `trackedNode` from traversing any metadata fields.
    // Example:
    // dataID: `4`
    // node: `friends.first(3)`
    // diffNode: null
    // splitQueries: `node(friend1) {...}`, `node(friend2) {...}`
    //
    // In this case the two fetched `node` queries do not reflect the fact that
    // `friends.first(3)` were fetched for item `4`, so `friends.first(3)` has
    // to be tracked as-is.
    if (hasSplitQueries) {
      trackedNode = field;
    }

    return {
      diffNode: diffNode,
      trackedNode: this._queryTracker ? trackedNode : null
    };
  };

  /**
   * Diff an `edges` field for the edge rooted at `edgeID`, splitting a new
   * root query to fetch any missing data (via a `node(id)` root if the
   * field is refetchable or a `...{connection.find(id){}}` query if the
   * field is not refetchable).
   */


  RelayDiffQueryBuilder.prototype.diffConnectionEdge = function diffConnectionEdge(connectionField, edgeField, path, edgeID, rangeInfo) {

    var hasSplitQueries = false;
    var diffOutput = this.traverse(edgeField, require('./RelayQueryPath').getPath(path, edgeField, edgeID), makeScope(edgeID));
    var diffNode = diffOutput ? diffOutput.diffNode : null;
    var trackedNode = diffOutput ? diffOutput.trackedNode : null;
    var nodeID = this._store.getLinkedRecordID(edgeID, NODE);

    if (diffNode) {
      if (!nodeID || require('./RelayRecord').isClientID(nodeID)) {
        require('fbjs/lib/warning')(connectionField.isConnectionWithoutNodeID(), 'RelayDiffQueryBuilder: Field `node` on connection `%s` cannot be ' + 'retrieved if it does not have an `id` field. If you expect fields ' + 'to be retrieved on this field, add an `id` field in the schema. ' + 'If you choose to ignore this warning, you can silence it by ' + 'adding `@relay(isConnectionWithoutNodeID: true)` to the ' + 'connection field.', connectionField.getStorageKey());
      } else {
        var _splitNodeAndEdgesFie = splitNodeAndEdgesFields(diffNode);

        var diffEdgesField = _splitNodeAndEdgesFie.edges;
        var diffNodeField = _splitNodeAndEdgesFie.node;

        // split missing `node` fields into a `node(id)` root query

        if (diffNodeField) {
          hasSplitQueries = true;
          var nodeField = edgeField.getFieldByStorageKey('node');
          require('fbjs/lib/invariant')(nodeField, 'RelayDiffQueryBuilder: Expected connection `%s` to have a ' + '`node` field.', connectionField.getSchemaName());
          this.splitQuery(buildRoot(nodeID, diffNodeField.getChildren(), require('./RelayQueryPath').getName(path), nodeField.getType()));
        }

        // split missing `edges` fields into a `connection.find(id)` query
        // if `find` is supported, otherwise warn
        if (diffEdgesField) {
          if (connectionField.isFindable()) {
            diffEdgesField = diffEdgesField.clone(diffEdgesField.getChildren().concat(nodeWithID));
            var connectionFind = connectionField.cloneFieldWithCalls([diffEdgesField], rangeInfo.filterCalls.concat({ name: 'find', value: nodeID }));
            if (connectionFind) {
              hasSplitQueries = true;
              // current path has `parent`, `connection`, `edges`; pop to parent
              var connectionParent = require('./RelayQueryPath').getParent(require('./RelayQueryPath').getParent(path));
              var connectionQuery = require('./RelayQueryPath').getQuery(this._store, connectionParent, connectionFind);
              this.splitQuery(connectionQuery);
            }
          } else {
            require('fbjs/lib/warning')(false, 'RelayDiffQueryBuilder: connection `edges{*}` fields can only ' + 'be refetched if the connection supports the `find` call. ' + 'Cannot refetch data for field `%s`.', connectionField.getStorageKey());
          }
        }
      }
    }

    // Connection edges will never return diff nodes; instead missing fields
    // are fetched by new root queries. Tracked nodes are returned if either
    // a child field was tracked or missing fields were split into a new query.
    // The returned `trackedNode` is never tracked directly: instead it serves
    // as an indicator to `diffConnection` that the entire connection field must
    // be tracked.
    return this._queryTracker ? {
      diffNode: null,
      trackedNode: hasSplitQueries ? edgeField : trackedNode
    } : null;
  };

  return RelayDiffQueryBuilder;
}();

/**
 * Helper to construct a plain scope for the given `dataID`.
 */


function makeScope(dataID) {
  return {
    connectionField: null,
    dataID: dataID,
    edgeID: null,
    rangeInfo: null
  };
}

/**
 * Returns a clone of the input with `edges` and `node` sub-fields split into
 * separate `edges` and `node` roots. Example:
 *
 * Input:
 * edges {
 *   edge_field,
 *   node {
 *     a,
 *     b
 *   },
 *   ${
 *     Fragment {
 *       edge_field_2,
 *       node {
 *         c
 *       }
 *     }
 *   }
 * }
 *
 * Output:
 * node:
 *   edges {
 *     a,      // flattened
 *     b,      // flattend
 *     ${
 *       Fragment {
 *         c  // flattened
 *       }
 *     }
 *   }
 * edges:
 *   edges {
 *     edge_field,
 *     ${
 *       Fragment {
 *         edge_field_2
 *       }
 *     }
 *   }
 */
function splitNodeAndEdgesFields(edgeOrFragment) {
  var children = edgeOrFragment.getChildren();
  var edgeChildren = [];
  var nodeChild = null;
  var nodeChildren = [];
  var hasEdgeChild = false;
  for (var ii = 0; ii < children.length; ii++) {
    var child = children[ii];
    if (child instanceof require('./RelayQuery').Field) {
      if (child.getSchemaName() === NODE) {
        var subFields = child.getChildren();
        nodeChildren = nodeChildren.concat(subFields);
        // can skip if `node` only has an `id` field
        if (!nodeChild) {
          if (subFields.length === 1) {
            var subField = subFields[0];
            if (!(subField instanceof require('./RelayQuery').Field) || subField.getSchemaName() !== ID) {
              nodeChild = child;
            }
          } else {
            nodeChild = child;
          }
        }
      } else {
        edgeChildren.push(child);
        hasEdgeChild = hasEdgeChild || !child.isRequisite();
      }
    } else if (child instanceof require('./RelayQuery').Fragment) {
      var _splitNodeAndEdgesFie2 = splitNodeAndEdgesFields(child);

      var _edges = _splitNodeAndEdgesFie2.edges;
      var _node = _splitNodeAndEdgesFie2.node;

      if (_edges) {
        edgeChildren.push(_edges);
        hasEdgeChild = true;
      }
      if (_node) {
        nodeChildren.push(_node);
        nodeChild = _node;
      }
    }
  }

  return {
    edges: hasEdgeChild ? edgeOrFragment.clone(edgeChildren) : null,
    node: nodeChild && require('./RelayQuery').Fragment.build('diffRelayQuery', nodeChild.getType(), nodeChildren, {
      isAbstract: nodeChild.isAbstract()
    })
  };
}

function buildRoot(rootID, nodes, name, type) {
  var children = [idField, typeField];
  var fields = [];
  nodes.forEach(function (node) {
    if (node instanceof require('./RelayQuery').Field) {
      fields.push(node);
    } else {
      children.push(node);
    }
  });
  children.push(require('./RelayQuery').Fragment.build('diffRelayQuery', type, fields));

  return require('./RelayQuery').Root.build(name, NODE, rootID, children, {
    identifyingArgName: ID,
    identifyingArgType: ID_TYPE,
    isAbstract: true,
    isDeferred: false,
    isPlural: false
  }, NODE_TYPE);
}

module.exports = require('./RelayProfiler').instrument('diffRelayQuery', diffRelayQuery);