/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayGraphModeInterface
 * 
 */

'use strict';

var RelayGraphModeInterface = {
  CACHE_KEY: '__key',
  DEFERRED_FRAGMENTS: '__deferred__',
  FRAGMENTS: '__fragments__',
  REF_KEY: '__ref',

  // Operation types
  PUT_EDGES: 'putEdges',
  PUT_NODES: 'putNodes',
  PUT_ROOT: 'putRoot'
};

module.exports = RelayGraphModeInterface;