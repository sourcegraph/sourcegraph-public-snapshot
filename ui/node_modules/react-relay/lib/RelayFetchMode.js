/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayFetchMode
 * 
 */

'use strict';

var _freeze2 = _interopRequireDefault(require('babel-runtime/core-js/object/freeze'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var RelayFetchMode = (0, _freeze2['default'])({
  CLIENT: 'CLIENT',
  PRELOAD: 'PRELOAD',
  REFETCH: 'REFETCH'
});

module.exports = RelayFetchMode;