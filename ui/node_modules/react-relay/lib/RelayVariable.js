/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayVariable
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var RelayVariable = function () {
  function RelayVariable(name) {
    (0, _classCallCheck3['default'])(this, RelayVariable);

    this.name = name;
  }

  RelayVariable.prototype.equals = function equals(other) {
    return other instanceof RelayVariable && other.getName() === this.name;
  };

  RelayVariable.prototype.getName = function getName() {
    return this.name;
  };

  return RelayVariable;
}();

module.exports = RelayVariable;