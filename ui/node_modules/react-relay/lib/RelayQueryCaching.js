/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayQueryCaching
 * 
 */

'use strict';

var queryCachingEnabled = true;

/**
 * Methods for configuring caching of Relay queries.
 */
var RelayQueryCaching = {
  /**
   * `disable` turns off caching of queries for `getRelayQueries` and
   * `buildRQL`.
   */

  disable: function disable() {
    queryCachingEnabled = false;
  },


  /**
   * @internal
   */
  getEnabled: function getEnabled() {
    return queryCachingEnabled;
  }
};

module.exports = RelayQueryCaching;