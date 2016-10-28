/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayQueryPath
 * 
 */

'use strict';

var EDGES = require('./RelayConnectionInterface').EDGES;

var ID = require('./RelayNodeInterface').ID;

var ID_TYPE = require('./RelayNodeInterface').ID_TYPE;

var NODE = require('./RelayNodeInterface').NODE;

var NODE_TYPE = require('./RelayNodeInterface').NODE_TYPE;

var TYPENAME = require('./RelayNodeInterface').TYPENAME;

var idField = require('./RelayQuery').Field.build({
  fieldName: ID,
  type: 'String'
});
var typeField = require('./RelayQuery').Field.build({
  fieldName: TYPENAME,
  type: 'String'
});

/**
 * @internal
 *
 * Represents the path (root plus fields) within a query that fetched a
 * particular node. Each step of the path may represent a root query (for
 * refetchable nodes) or the field path from the nearest refetchable node.
 */
var RelayQueryPath = {
  createForID: function createForID(dataID, name, routeName) {
    require('fbjs/lib/invariant')(!require('./RelayRecord').isClientID(dataID), 'RelayQueryPath.createForID: Expected dataID to be a server id, got ' + '`%s`.', dataID);
    return {
      dataID: dataID,
      name: name,
      routeName: routeName || '$RelayQuery',
      type: 'node'
    };
  },
  create: function create(root) {
    if (root.getFieldName() === NODE) {
      var identifyingArg = root.getIdentifyingArg();
      if (identifyingArg && typeof identifyingArg.value === 'string') {
        return {
          dataID: identifyingArg.value,
          name: root.getName(),
          routeName: root.getRoute().name,
          type: 'node'
        };
      }
    }
    return {
      root: root,
      type: 'root'
    };
  },
  getPath: function getPath(parent, node, dataID) {
    if (dataID == null || require('./RelayRecord').isClientID(dataID)) {
      return {
        node: node,
        parent: parent,
        type: 'client'
      };
    } else if (parent.type === 'node' && parent.dataID === dataID) {
      return parent;
    } else {
      return {
        dataID: dataID,
        name: RelayQueryPath.getName(parent),
        routeName: RelayQueryPath.getRouteName(parent),
        type: 'node'
      };
    }
  },
  isRootPath: function isRootPath(path) {
    return path.type === 'node' || path.type === 'root';
  },
  getParent: function getParent(path) {
    require('fbjs/lib/invariant')(path.type === 'client', 'RelayQueryPath: Cannot get the parent of a root path.');
    return path.parent;
  },
  getName: function getName(path) {
    while (path.type === 'client') {
      path = path.parent;
    }
    if (path.type === 'root') {
      return path.root.getName();
    } else if (path.type === 'node') {
      return path.name;
    } else {
      require('fbjs/lib/invariant')(false, 'RelayQueryPath.getName(): Invalid path `%s`.', path);
    }
  },
  getRouteName: function getRouteName(path) {
    while (path.type === 'client') {
      path = path.parent;
    }
    if (path.type === 'root') {
      return path.root.getRoute().name;
    } else if (path.type === 'node') {
      return path.routeName;
    } else {
      require('fbjs/lib/invariant')(false, 'RelayQueryPath.getRouteName(): Invalid path `%s`.', path);
    }
  },
  getQuery: function getQuery(store, path, appendNode) {
    var child = appendNode;
    var prevField = void 0;
    while (path.type === 'client') {
      var _node = path.node;
      if (_node instanceof require('./RelayQuery').Field) {
        var schemaName = _node.getSchemaName();
        require('fbjs/lib/warning')(!prevField || prevField !== EDGES || !_node.isConnection(), 'RelayQueryPath.getQuery(): Cannot generate accurate query for ' + 'path with connection `%s`. Consider adding an `id` field to each ' + '`node` to make them refetchable.', schemaName);
        prevField = schemaName;
      }
      var idFieldName = _node instanceof require('./RelayQuery').Field ? _node.getInferredPrimaryKey() : ID;
      if (idFieldName) {
        child = _node.clone([child, _node.getFieldByStorageKey(idFieldName), _node.getFieldByStorageKey(TYPENAME)]);
      } else {
        child = _node.clone([child]);
      }
      path = path.parent;
    }
    var root = path.type === 'root' ? path.root : createRootQueryFromNodePath(path);
    var children = [child, root.getFieldByStorageKey(ID), root.getFieldByStorageKey(TYPENAME)];
    var rootChildren = getRootFragmentForQuery(store, root, children);
    var pathQuery = root.cloneWithRoute(rootChildren, appendNode.getRoute());
    // for flow
    require('fbjs/lib/invariant')(pathQuery instanceof require('./RelayQuery').Root, 'RelayQueryPath: Expected the root of path `%s` to be a query.', RelayQueryPath.getName(path));
    return pathQuery;
  }
};

function createRootQueryFromNodePath(nodePath) {
  return require('./RelayQuery').Root.build(nodePath.name, NODE, nodePath.dataID, [idField, typeField], {
    identifyingArgName: ID,
    identifyingArgType: ID_TYPE,
    isAbstract: true,
    isDeferred: false,
    isPlural: false
  }, NODE_TYPE, nodePath.routeName);
}

function getRootFragmentForQuery(store, root, children) {
  var nextChildren = [];
  // $FlowIssue: Flow isn't recognizing that `filter(x => !!x)` returns a list
  // of non-null values.
  children.forEach(function (child) {
    if (child) {
      nextChildren.push(child);
    }
  });
  if (!root.isAbstract()) {
    // No need to wrap child nodes of a known concrete type.
    return nextChildren;
  }
  var identifyingArgKeys = [];
  require('./forEachRootCallArg')(root, function (_ref) {
    var identifyingArgKey = _ref.identifyingArgKey;

    identifyingArgKeys.push(identifyingArgKey);
  });
  var identifyingArgKey = identifyingArgKeys[0];
  var rootID = store.getDataID(root.getStorageKey(), identifyingArgKey);
  var rootType = rootID && store.getType(rootID);

  if (rootType != null) {
    return [require('./RelayQuery').Fragment.build(root.getName(), rootType, nextChildren)];
  } else {
    var rootState = rootID != null ? store.getRecordState(rootID) : require('./RelayRecordState').UNKNOWN;
    require('fbjs/lib/warning')(false, 'RelayQueryPath: No typename found for %s record `%s`. Generating a ' + 'possibly invalid query.', rootState.toLowerCase(), rootID);
    return nextChildren;
  }
}

module.exports = RelayQueryPath;