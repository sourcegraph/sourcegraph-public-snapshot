/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayQueryConfig
 * 
 */

'use strict';

var _extends3 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _freeze2 = _interopRequireDefault(require('babel-runtime/core-js/object/freeze'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * Configures the root queries and initial variables that define the context in
 * which the top-level component's fragments are requested. This is meant to be
 * subclassed, of which instances are supplied to `RelayRootContainer`.
 */

var RelayQueryConfig = function () {

  // TODO: Deprecate `routeName`, #8478719.

  function RelayQueryConfig(initialVariables) {
    (0, _classCallCheck3['default'])(this, RelayQueryConfig);

    require('fbjs/lib/invariant')(this.constructor !== RelayQueryConfig, 'RelayQueryConfig: Abstract class cannot be instantiated.');

    Object.defineProperty(this, 'name', {
      enumerable: true,
      value: this.constructor.routeName
    });
    Object.defineProperty(this, 'params', {
      enumerable: true,
      value: this.prepareVariables((0, _extends3['default'])({}, initialVariables)) || {}
    });
    Object.defineProperty(this, 'queries', {
      enumerable: true,
      value: (0, _extends3['default'])({}, this.constructor.queries)
    });

    if (process.env.NODE_ENV !== 'production') {
      (0, _freeze2['default'])(this.params);
      (0, _freeze2['default'])(this.queries);
    }
  }

  /**
   * Provides an opportunity to perform additional logic on the variables.
   * Child class should override this function to perform custom logic.
   */


  RelayQueryConfig.prototype.prepareVariables = function prepareVariables(prevVariables) {
    return prevVariables;
  };

  return RelayQueryConfig;
}();

module.exports = RelayQueryConfig;