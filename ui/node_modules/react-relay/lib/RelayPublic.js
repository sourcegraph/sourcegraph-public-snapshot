/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayPublic
 * 
 */

'use strict';

if (typeof global.__REACT_DEVTOOLS_GLOBAL_HOOK__ !== 'undefined') {
  global.__REACT_DEVTOOLS_GLOBAL_HOOK__._relayInternals = require('./RelayInternals');
}

/**
 * Relay contains the set of public methods used to initialize and orchestrate
 * a React application that uses GraphQL to declare data dependencies.
 */
var RelayPublic = {
  Environment: require('./RelayEnvironment'),
  Mutation: require('./RelayMutation'),
  PropTypes: require('./RelayPropTypes'),
  QL: require('./RelayQL'),
  QueryConfig: require('./RelayQueryConfig'),
  ReadyStateRenderer: require('./RelayReadyStateRenderer'),
  Renderer: require('./RelayRenderer'),
  RootContainer: require('./RelayRootContainer'),
  Route: require('./RelayRoute'),
  Store: require('./RelayStore'),

  createContainer: require('./RelayContainer').create,
  createQuery: require('./createRelayQuery'),
  getQueries: require('./getRelayQueries'),
  disableQueryCaching: require('./RelayQueryCaching').disable,
  injectNetworkLayer: require('./RelayStore').injectNetworkLayer.bind(require('./RelayStore')),
  injectTaskScheduler: require('./RelayStore').injectTaskScheduler.bind(require('./RelayStore')),
  isContainer: require('./isRelayContainer')
};

module.exports = RelayPublic;