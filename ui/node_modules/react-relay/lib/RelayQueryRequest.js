/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayQueryRequest
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
 * Instances of these are made available via `RelayNetworkLayer.sendQueries`.
 */

var RelayQueryRequest = function (_Deferred) {
  (0, _inherits3['default'])(RelayQueryRequest, _Deferred);

  function RelayQueryRequest(query) {
    (0, _classCallCheck3['default'])(this, RelayQueryRequest);

    var _this = (0, _possibleConstructorReturn3['default'])(this, _Deferred.call(this));

    _this._printedQuery = null;
    _this._query = query;
    return _this;
  }

  /**
   * @public
   *
   * Gets a string name used to refer to this request for printing debug output.
   */


  RelayQueryRequest.prototype.getDebugName = function getDebugName() {
    return this._query.getName();
  };

  /**
   * @public
   *
   * Gets a unique identifier for this query. These identifiers are useful for
   * assigning response payloads to their corresponding queries when sent in a
   * single GraphQL request.
   */


  RelayQueryRequest.prototype.getID = function getID() {
    return this._query.getID();
  };

  /**
   * @public
   *
   * Gets the variables used by the query. These variables should be serialized
   * and sent in the GraphQL request.
   */


  RelayQueryRequest.prototype.getVariables = function getVariables() {
    var printedQuery = this._printedQuery;
    if (!printedQuery) {
      printedQuery = require('./printRelayQuery')(this._query);
      this._printedQuery = printedQuery;
    }
    return printedQuery.variables;
  };

  /**
   * @public
   *
   * Gets a string representation of the GraphQL query.
   */


  RelayQueryRequest.prototype.getQueryString = function getQueryString() {
    var printedQuery = this._printedQuery;
    if (!printedQuery) {
      printedQuery = require('./printRelayQuery')(this._query);
      this._printedQuery = printedQuery;
    }
    return printedQuery.text;
  };

  /**
   * @public
   * @unstable
   */


  RelayQueryRequest.prototype.getQuery = function getQuery() {
    return this._query;
  };

  return RelayQueryRequest;
}(require('fbjs/lib/Deferred'));

module.exports = RelayQueryRequest;