/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule buildRQL
 * 
 */

'use strict';

var _extends3 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

// Cache results of executing fragment query builders.
var fragmentCache = new (require('fbjs/lib/Map'))();

// Cache results of executing component-specific route query builders.
var queryCache = new (require('fbjs/lib/Map'))();

function isDeprecatedCallWithArgCountGreaterThan(nodeBuilder, count) {
  var argLength = nodeBuilder.length;
  if (process.env.NODE_ENV !== 'production') {
    var mockImpl = nodeBuilder;
    while (mockImpl && mockImpl._getMockImplementation) {
      mockImpl = mockImpl._getMockImplementation();
    }
    if (mockImpl) {
      argLength = mockImpl.length;
    }
  }
  return argLength > count;
}

/**
 * @internal
 *
 * Builds a static node representation using a supplied query or fragment
 * builder. This is used for routes, containers, and mutations.
 *
 * If the supplied fragment builder produces an invalid node (e.g. the wrong
 * node type), these will return `undefined`. This is not to be confused with
 * a return value of `null`, which may result from the lack of a node.
 */
var buildRQL = {
  Fragment: function Fragment(fragmentBuilder, values) {
    var node = fragmentCache.get(fragmentBuilder);
    if (node) {
      return require('./QueryBuilder').getFragment(node);
    }
    var variables = toVariables(values);
    require('fbjs/lib/invariant')(!isDeprecatedCallWithArgCountGreaterThan(fragmentBuilder, 1), 'Relay.QL: Deprecated usage detected. If you are trying to define a ' + 'fragment, use `variables => Relay.QL`.');
    node = fragmentBuilder(variables);
    var fragment = node != null ? require('./QueryBuilder').getFragment(node) : null;
    if (!fragment) {
      return fragment;
    }
    fragmentCache.set(fragmentBuilder, fragment);
    return fragment;
  },
  Query: function Query(queryBuilder, Component, queryName, values) {
    var queryCacheEnabled = require('./RelayQueryCaching').getEnabled();
    var node = void 0;
    if (!queryCacheEnabled) {
      node = buildNode(queryBuilder, Component, queryName, values);
    } else {
      var componentCache = queryCache.get(queryBuilder);
      if (!componentCache) {
        componentCache = new (require('fbjs/lib/Map'))();
        queryCache.set(queryBuilder, componentCache);
      } else {
        node = componentCache.get(Component);
      }
      if (!node) {
        node = buildNode(queryBuilder, Component, queryName, values);
      }
      componentCache.set(Component, node);
    }
    if (node) {
      return require('./QueryBuilder').getQuery(node) || undefined;
    }
    return null;
  }
};

/**
 * @internal
 */
function buildNode(queryBuilder, Component, queryName, values) {
  var variables = toVariables(values);
  require('fbjs/lib/invariant')(!isDeprecatedCallWithArgCountGreaterThan(queryBuilder, 2), 'Relay.QL: Deprecated usage detected. If you are trying to define a ' + 'query, use `(Component, variables) => Relay.QL`.');
  var node = void 0;
  if (isDeprecatedCallWithArgCountGreaterThan(queryBuilder, 0)) {
    node = queryBuilder(Component, variables);
  } else {
    node = queryBuilder(Component, variables);
    var query = require('./QueryBuilder').getQuery(node);
    if (query) {
      (function () {
        var hasFragment = false;
        var hasScalarFieldsOnly = true;
        if (query.children) {
          query.children.forEach(function (child) {
            if (child) {
              hasFragment = hasFragment || child.kind === 'Fragment';
              hasScalarFieldsOnly = hasScalarFieldsOnly && child.kind === 'Field' && (!child.children || child.children.length === 0);
            }
          });
        }
        if (!hasFragment) {
          var children = query.children ? [].concat(query.children) : [];
          require('fbjs/lib/invariant')(hasScalarFieldsOnly, 'Relay.QL: Expected query `%s` to be empty. For example, use ' + '`node(id: $id)`, not `node(id: $id) { ... }`.', query.fieldName);
          var fragmentVariables = require('fbjs/lib/filterObject')(variables, function (_, name) {
            return Component.hasVariable(name);
          });
          children.push(Component.getFragment(queryName, fragmentVariables));
          node = (0, _extends3['default'])({}, query, {
            children: children
          });
        }
      })();
    }
  }
  return node;
}

function toVariables(variables) // ConcreteCallVariable should flow into mixed
{
  return require('fbjs/lib/mapObject')(variables, function (_, name) {
    return require('./QueryBuilder').createCallVariable(name);
  });
}

require('./RelayProfiler').instrumentMethods(buildRQL, {
  Fragment: 'buildRQL.Fragment',
  Query: 'buildRQL.Query'
});

module.exports = buildRQL;