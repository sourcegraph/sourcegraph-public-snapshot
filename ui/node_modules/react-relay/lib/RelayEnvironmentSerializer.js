/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayEnvironmentSerializer
 * 
 */

'use strict';

var _stringify2 = _interopRequireDefault(require('babel-runtime/core-js/json/stringify'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var RelayEnvironmentSerializer = {
  serialize: function serialize(relayEnvironment) {
    return (0, _stringify2['default'])(relayEnvironment.getStoreData());
  },
  deserialize: function deserialize(str) {
    return new (require('./RelayEnvironment'))(require('./RelayStoreData').fromJSON(JSON.parse(str)));
  }
};

module.exports = RelayEnvironmentSerializer;