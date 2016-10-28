/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule getRelayQueries
 * 
 */

'use strict';

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var queryCache = new (require('fbjs/lib/Map'))();

/**
 * @internal
 *
 * `getRelayQueries` retrieves all queries for a component given a route.
 */
function getRelayQueries(Component, route) {
  var queryCachingEnabled = require('./RelayQueryCaching').getEnabled();
  if (!queryCachingEnabled) {
    return buildQuerySet(Component, route);
  }
  var cache = queryCache.get(Component);
  if (!cache) {
    cache = {};
    queryCache.set(Component, cache);
  }
  var cacheKey = route.name + ':' + require('./stableStringify')(route.params);
  if (cache.hasOwnProperty(cacheKey)) {
    return cache[cacheKey];
  }
  var querySet = buildQuerySet(Component, route);
  cache[cacheKey] = querySet;
  return querySet;
}

/**
 * @internal
 */
function buildQuerySet(Component, route) {
  var querySet = {};
  Component.getFragmentNames().forEach(function (fragmentName) {
    querySet[fragmentName] = null;
  });
  (0, _keys2['default'])(route.queries).forEach(function (queryName) {
    if (!Component.hasFragment(queryName)) {
      require('fbjs/lib/warning')(false, 'Relay.QL: query `%s.queries.%s` is invalid, expected fragment ' + '`%s.fragments.%s` to be defined.', route.name, queryName, Component.displayName, queryName);
      return;
    }
    var queryBuilder = route.queries[queryName];
    if (queryBuilder) {
      var concreteQuery = require('./buildRQL').Query(queryBuilder, Component, queryName, route.params);
      require('fbjs/lib/invariant')(concreteQuery !== undefined, 'Relay.QL: query `%s.queries.%s` is invalid, a typical query is ' + 'defined using: () => Relay.QL`query { ... }`.', route.name, queryName);
      if (concreteQuery) {
        var rootQuery = require('./RelayQuery').Root.create(concreteQuery, require('./RelayMetaRoute').get(route.name), route.params);
        var identifyingArg = rootQuery.getIdentifyingArg();
        if (!identifyingArg || identifyingArg.value !== undefined) {
          querySet[queryName] = rootQuery;
          return;
        }
      }
    }
    querySet[queryName] = null;
  });
  return querySet;
}

module.exports = require('./RelayProfiler').instrument('Relay.getQueries', getRelayQueries);