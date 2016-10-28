/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayRefQueryDescriptor
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @internal
 *
 * Represents a node that will eventually become a "ref query".
 *
 * Includes the `nodePath` (ancestor nodes) that can be used to construct an
 * appropriate the JSONPath for the query.
 *
 * @see splitDeferredRelayQueries
 */

var RelayRefQueryDescriptor = function RelayRefQueryDescriptor(node, nodePath) {
  (0, _classCallCheck3['default'])(this, RelayRefQueryDescriptor);

  this.node = node;
  this.nodePath = nodePath;
};

module.exports = RelayRefQueryDescriptor;