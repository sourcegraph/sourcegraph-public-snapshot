/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayMutationRequest
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @internal
 *
 * Instances of these are made available via `RelayNetworkLayer.sendMutation`.
 */

var RelayMutationRequest = function (_Deferred) {
  (0, _inherits3['default'])(RelayMutationRequest, _Deferred);

  function RelayMutationRequest(mutation, files) {
    (0, _classCallCheck3['default'])(this, RelayMutationRequest);

    var _this = (0, _possibleConstructorReturn3['default'])(this, _Deferred.call(this));

    _this._mutation = mutation;
    _this._printedQuery = null;
    _this._files = files;
    return _this;
  }

  /**
   * @public
   *
   * Gets a string name used to refer to this request for printing debug output.
   */


  RelayMutationRequest.prototype.getDebugName = function getDebugName() {
    return this._mutation.getName();
  };

  /**
   * @public
   *
   * Gets an optional map from name to File objects.
   */


  RelayMutationRequest.prototype.getFiles = function getFiles() {
    return this._files;
  };

  /**
   * @public
   *
   * Gets the variables used by the mutation. These variables should be
   * serialized and sent in the GraphQL request.
   */


  RelayMutationRequest.prototype.getVariables = function getVariables() {
    return this._getPrintedQuery().variables;
  };

  /**
   * @public
   *
   * Gets a string representation of the GraphQL mutation.
   */


  RelayMutationRequest.prototype.getQueryString = function getQueryString() {
    return this._getPrintedQuery().text;
  };

  /**
   * @public
   * @unstable
   */


  RelayMutationRequest.prototype.getMutation = function getMutation() {
    return this._mutation;
  };

  /**
   * @private
   *
   * Returns the memoized printed query.
   */


  RelayMutationRequest.prototype._getPrintedQuery = function _getPrintedQuery() {
    if (!this._printedQuery) {
      this._printedQuery = require('./printRelayQuery')(this._mutation);
    }
    return this._printedQuery;
  };

  return RelayMutationRequest;
}(require('fbjs/lib/Deferred'));

module.exports = RelayMutationRequest;