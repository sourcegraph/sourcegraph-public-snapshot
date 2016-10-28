/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayPendingQueryTracker
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

var _promise2 = _interopRequireDefault(require('fbjs/lib/Promise'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @internal
 *
 * Tracks pending (in-flight) queries.
 */

var RelayPendingQueryTracker = function () {
  // Asynchronous mapping from preload query IDs to results.

  function RelayPendingQueryTracker(storeData) {
    (0, _classCallCheck3['default'])(this, RelayPendingQueryTracker);

    this._pendingFetchMap = {};
    this._preloadQueryMap = new (require('fbjs/lib/PromiseMap'))();
    this._storeData = storeData;
  }

  /**
   * Used by `GraphQLQueryRunner` to enqueue new queries.
   */


  RelayPendingQueryTracker.prototype.add = function add(params) {
    return new PendingFetch(params, {
      pendingFetchMap: this._pendingFetchMap,
      preloadQueryMap: this._preloadQueryMap,
      storeData: this._storeData
    });
  };

  RelayPendingQueryTracker.prototype.hasPendingQueries = function hasPendingQueries() {
    return hasItems(this._pendingFetchMap);
  };

  /**
   * Clears all pending query tracking. Does not cancel the queries themselves.
   */


  RelayPendingQueryTracker.prototype.resetPending = function resetPending() {
    this._pendingFetchMap = {};
  };

  RelayPendingQueryTracker.prototype.resolvePreloadQuery = function resolvePreloadQuery(queryID, result) {
    this._preloadQueryMap.resolveKey(queryID, result);
  };

  RelayPendingQueryTracker.prototype.rejectPreloadQuery = function rejectPreloadQuery(queryID, error) {
    this._preloadQueryMap.rejectKey(queryID, error);
  };

  return RelayPendingQueryTracker;
}();

/**
 * @private
 */


var PendingFetch = function () {
  function PendingFetch(_ref, _ref2) {
    var fetchMode = _ref.fetchMode;
    var forceIndex = _ref.forceIndex;
    var query = _ref.query;
    var pendingFetchMap = _ref2.pendingFetchMap;
    var preloadQueryMap = _ref2.preloadQueryMap;
    var storeData = _ref2.storeData;
    (0, _classCallCheck3['default'])(this, PendingFetch);

    var queryID = query.getID();
    this._forceIndex = forceIndex;
    this._pendingFetchMap = pendingFetchMap;
    this._preloadQueryMap = preloadQueryMap;
    this._query = query;
    this._resolvedDeferred = new (require('fbjs/lib/Deferred'))();
    this._resolvedQuery = false;
    this._storeData = storeData;

    this._fetchQueryPromise = fetchMode === require('./RelayFetchMode').PRELOAD ? this._preloadQueryMap.get(queryID) : storeData.getNetworkLayer().fetchRelayQuery(query);

    this._fetchedQuery = false;
    this._error = null;

    this._pendingFetchMap[queryID] = {
      fetch: this,
      query: query
    };
    this._fetchQueryPromise.done(this._handleQuerySuccess.bind(this), this._handleQueryFailure.bind(this));
  }

  PendingFetch.prototype.isResolvable = function isResolvable() {
    return this._resolvedQuery;
  };

  PendingFetch.prototype.getQuery = function getQuery() {
    return this._query;
  };

  PendingFetch.prototype.getResolvedPromise = function getResolvedPromise() {
    return this._resolvedDeferred.getPromise();
  };

  PendingFetch.prototype._handleQuerySuccess = function _handleQuerySuccess(result) {
    var _this = this;

    this._fetchedQuery = true;

    this._storeData.getTaskQueue().enqueue(function () {
      var response = result.response;
      require('fbjs/lib/invariant')(response && typeof response === 'object', 'RelayPendingQueryTracker: Expected response to be an object, got ' + '`%s`.', response ? typeof response : response);
      _this._storeData.handleQueryPayload(_this._query, response, _this._forceIndex);
    }).done(this._markQueryAsResolved.bind(this), this._markAsRejected.bind(this));
  };

  PendingFetch.prototype._handleQueryFailure = function _handleQueryFailure(error) {
    this._markAsRejected(error);
  };

  PendingFetch.prototype._markQueryAsResolved = function _markQueryAsResolved() {
    var queryID = this.getQuery().getID();
    delete this._pendingFetchMap[queryID];

    this._resolvedQuery = true;
    this._updateResolvedDeferred();
  };

  PendingFetch.prototype._markAsRejected = function _markAsRejected(error) {
    var queryID = this.getQuery().getID();
    delete this._pendingFetchMap[queryID];

    console.warn(error.message);

    this._error = error;
    this._updateResolvedDeferred();
  };

  PendingFetch.prototype._updateResolvedDeferred = function _updateResolvedDeferred() {
    if (this._isSettled() && !this._resolvedDeferred.isSettled()) {
      if (this._error) {
        this._resolvedDeferred.reject(this._error);
      } else {
        this._resolvedDeferred.resolve(undefined);
      }
    }
  };

  PendingFetch.prototype._isSettled = function _isSettled() {
    return !!this._error || this._resolvedQuery;
  };

  return PendingFetch;
}();

function hasItems(map) {
  return !!(0, _keys2['default'])(map).length;
}

module.exports = RelayPendingQueryTracker;