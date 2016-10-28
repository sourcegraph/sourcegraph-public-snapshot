/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayQL
 * 
 */

'use strict';

var _assign2 = _interopRequireDefault(require('babel-runtime/core-js/object/assign'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @public
 *
 * This is a tag function used with template strings to provide the facade of a
 * runtime GraphQL parser. Example usage:
 *
 *   Relay.QL`fragment on User { name }`
 *
 * In actuality, a Babel transform parses these tag templates and replaces it
 * with an internal representation of the query structure.
 */
function RelayQL(strings) {
  require('fbjs/lib/invariant')(false, 'RelayQL: Unexpected invocation at runtime. Either the Babel transform ' + 'was not set up, or it failed to identify this call site. Make sure it ' + 'is being used verbatim as `Relay.QL`.');
}

function assertValidFragment(substitution) {
  require('fbjs/lib/invariant')(substitution instanceof require('./RelayFragmentReference') || require('./QueryBuilder').getFragment(substitution) || require('./QueryBuilder').getFragmentReference(substitution), 'RelayQL: Invalid fragment composition, use ' + '`${Child.getFragment(\'name\')}`.');
}

/**
 * Private helper methods used by the transformed code.
 */
(0, _assign2['default'])(RelayQL, {
  __frag: function __frag(substitution) {
    if (typeof substitution === 'function') {
      // Route conditional fragment, e.g. `${route => matchRoute(route, ...)}`.
      return new (require('./RelayRouteFragment'))(substitution);
    }
    if (substitution != null) {
      if (Array.isArray(substitution)) {
        substitution.forEach(assertValidFragment);
      } else {
        assertValidFragment(substitution);
      }
    }
    return substitution;
  },
  __var: function __var(expression) {
    var variable = require('./QueryBuilder').getCallVariable(expression);
    if (variable) {
      require('fbjs/lib/invariant')(false, 'RelayQL: Invalid argument `%s` supplied via template substitution. ' + 'Instead, use an inline variable (e.g. `comments(count: $count)`).', variable.callVariableName);
    }
    return require('./QueryBuilder').createCallValue(expression);
  },
  __id: function __id() {
    return require('./generateConcreteFragmentID')();
  },
  __createFragment: function __createFragment(fragment, variableMapping) {
    return new (require('./RelayFragmentReference'))(function () {
      return fragment;
    }, null, variableMapping);
  }
});

module.exports = RelayQL;