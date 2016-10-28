/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule createRelayQuery
 * 
 */

'use strict';

function createRelayQuery(node, variables) {
  require('fbjs/lib/invariant')(typeof variables === 'object' && variables != null && !Array.isArray(variables), 'Relay.Query: Expected `variables` to be an object.');
  return require('./RelayQuery').Root.create(node, require('./RelayMetaRoute').get('$createRelayQuery'), variables);
}

module.exports = createRelayQuery;