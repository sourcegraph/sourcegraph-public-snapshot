/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayMutationDebugPrinter
 * 
 */

'use strict';

/**
 * @internal
 *
 * Helper functions to print mutation queries for debugging purposes.
 */
var RelayMutationDebugPrinter = {
  printOptimisticMutation: function printOptimisticMutation(query, response) {
    /* eslint-disable no-console */
    if (!console.groupCollapsed || !console.groupEnd) {
      return;
    }
    RelayMutationDebugPrinter.printMutation(query, 'Optimistic');

    console.groupCollapsed('Optimistic Response');
    console.log(response);
    console.groupEnd();
    /* eslint-enable no-console */
  },
  printMutation: function printMutation(query, name) {
    /* eslint-disable no-console */
    if (!console.groupCollapsed || !console.groupEnd) {
      return;
    }
    var printedQuery = query ? require('./printRelayQuery')(query) : null;
    name = name || 'Mutation';

    console.groupCollapsed(name + ' Variables');
    console.log(printedQuery ? printedQuery.variables : {});
    console.groupEnd();

    console.groupCollapsed(name + ' Query');
    console.log(printedQuery ? printedQuery.text : '');
    console.groupEnd();
    /* eslint-enable no-console */
  }
};

module.exports = RelayMutationDebugPrinter;