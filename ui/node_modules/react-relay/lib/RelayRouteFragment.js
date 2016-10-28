/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayRouteFragment
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * Represents a query fragment that is conditional upon the active route as a
 * function that returns either a literal fragment or a fragment reference.
 *
 * Example GraphQL:
 *
 * ```
 * Node {
 *   ${(route) => matchRoute(route, ...)}
 * }
 * ```
 */

var RelayRouteFragment = function () {
  function RelayRouteFragment(builder) {
    (0, _classCallCheck3['default'])(this, RelayRouteFragment);

    this._builder = builder;
  }

  /**
   * Returns the query fragment that matches the given route, if any.
   */


  RelayRouteFragment.prototype.getFragmentForRoute = function getFragmentForRoute(route) {
    return this._builder(route);
  };

  return RelayRouteFragment;
}();

module.exports = RelayRouteFragment;