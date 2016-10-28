/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayCacheProcessor
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * An asynchronous traversal that knows how to read roots and nodes from a
 * `CacheManager`. Root reads yield the `dataID` of the root, if found.
 * Node reads yield the `Record` associated with a supplied `dataID`, if found.
 *
 * Visitors: Ensure that only one read is ever in flight for a given root/node.
 *           Maintain a list of states to process after each read completes.
 * Queuers:  Perform the work of kicking off a root/node read.
 * Handlers: Subclasses of `RelayCacheProcessor` can implement this method to
 *           actually perform work after a root/node read completes.
 */

var RelayCacheProcessor = function (_RelayQueryVisitor) {
  (0, _inherits3['default'])(RelayCacheProcessor, _RelayQueryVisitor);

  function RelayCacheProcessor(cacheManager, callbacks) {
    (0, _classCallCheck3['default'])(this, RelayCacheProcessor);

    var _this = (0, _possibleConstructorReturn3['default'])(this, _RelayQueryVisitor.call(this));

    _this._cacheManager = cacheManager;
    _this._callbacks = callbacks;
    _this._pendingNextStates = {};
    _this._pendingRoots = {};
    _this._state = 'PENDING';
    return _this;
  }

  RelayCacheProcessor.prototype.abort = function abort() {
    require('fbjs/lib/warning')(this._state === 'LOADING', 'RelayCacheProcessor: Can only abort an in-progress read operation.');
    this._state = 'COMPLETED';
  };

  RelayCacheProcessor.prototype.handleFailure = function handleFailure(error) {
    require('fbjs/lib/invariant')(this._state !== 'COMPLETED', 'RelayStoreReader: Query set already failed/completed.');
    this._state = 'COMPLETED';
    this._callbacks.onFailure && this._callbacks.onFailure(error);
  };

  RelayCacheProcessor.prototype.handleNodeVisited = function handleNodeVisited(node, dataID, record, nextState) {
    return;
  };

  RelayCacheProcessor.prototype.handleIdentifiedRootVisited = function handleIdentifiedRootVisited(query, dataID, identifyingArgKey, nextState) {
    return;
  };

  RelayCacheProcessor.prototype.process = function process(processorFn) {
    require('fbjs/lib/invariant')(this._state === 'PENDING', 'RelayCacheProcessor: A `read` is in progress.');
    this._state = 'LOADING';
    processorFn();
    if (this._isDone()) {
      this._handleSuccess();
    }
  };

  RelayCacheProcessor.prototype.queueIdentifiedRoot = function queueIdentifiedRoot(query, identifyingArgKey, nextState) {
    var _this2 = this;

    var storageKey = query.getStorageKey();
    this._cacheManager.readRootCall(storageKey, identifyingArgKey || '', function (error, dataID) {
      if (_this2._state === 'COMPLETED') {
        return;
      }
      if (error) {
        _this2.handleFailure(error);
        return;
      }
      _this2.handleIdentifiedRootVisited(query, dataID, identifyingArgKey, nextState);
      var rootKey = _this2._getRootKey(storageKey, identifyingArgKey);
      var pendingRoots = _this2._pendingRoots[rootKey];
      delete _this2._pendingRoots[rootKey];
      for (var ii = 0; ii < pendingRoots.length; ii++) {
        if (_this2._state === 'COMPLETED') {
          return;
        }
        _this2.traverse(pendingRoots[ii], nextState);
      }
      if (_this2._isDone()) {
        _this2._handleSuccess();
      }
    });
  };

  RelayCacheProcessor.prototype.queueNode = function queueNode(node, dataID, nextState) {
    var _this3 = this;

    this._cacheManager.readNode(dataID, function (error, record) {
      if (_this3._state === 'COMPLETED') {
        return;
      }
      if (error) {
        _this3.handleFailure(error);
        return;
      }
      _this3.handleNodeVisited(node, dataID, record, nextState);
      var pendingNextStates = _this3._pendingNextStates[dataID];
      delete _this3._pendingNextStates[dataID];
      for (var ii = 0; ii < pendingNextStates.length; ii++) {
        if (_this3._state === 'COMPLETED') {
          return;
        }
        _this3.traverse(node, pendingNextStates[ii]);
      }
      if (_this3._isDone()) {
        _this3._handleSuccess();
      }
    });
  };

  RelayCacheProcessor.prototype.visitIdentifiedRoot = function visitIdentifiedRoot(query, identifyingArgKey, nextState) {
    var storageKey = query.getStorageKey();
    var rootKey = this._getRootKey(storageKey, identifyingArgKey);
    if (this._pendingRoots.hasOwnProperty(rootKey)) {
      this._pendingRoots[rootKey].push(query);
    } else {
      this._pendingRoots[rootKey] = [query];
      this.queueIdentifiedRoot(query, identifyingArgKey, nextState);
    }
  };

  RelayCacheProcessor.prototype.visitNode = function visitNode(node, dataID, nextState) {
    if (this._pendingNextStates.hasOwnProperty(dataID)) {
      this._pendingNextStates[dataID].push(nextState);
    } else {
      this._pendingNextStates[dataID] = [nextState];
      this.queueNode(node, dataID, nextState);
    }
  };

  RelayCacheProcessor.prototype.visitRoot = function visitRoot(query, nextState) {
    var _this4 = this;

    require('./forEachRootCallArg')(query, function (_ref) {
      var identifyingArgKey = _ref.identifyingArgKey;

      if (_this4._state === 'COMPLETED') {
        return;
      }
      _this4.visitIdentifiedRoot(query, identifyingArgKey, nextState);
    });
  };

  RelayCacheProcessor.prototype._getRootKey = function _getRootKey(storageKey, identifyingArgKey) {
    return storageKey + '*' + (identifyingArgKey || '');
  };

  RelayCacheProcessor.prototype._handleSuccess = function _handleSuccess() {
    require('fbjs/lib/invariant')(this._state !== 'COMPLETED', 'RelayStoreReader: Query set already failed/completed.');
    this._state = 'COMPLETED';
    this._callbacks.onSuccess && this._callbacks.onSuccess();
  };

  RelayCacheProcessor.prototype._isDone = function _isDone() {
    return require('fbjs/lib/isEmpty')(this._pendingRoots) && require('fbjs/lib/isEmpty')(this._pendingNextStates) && this._state === 'LOADING';
  };

  return RelayCacheProcessor;
}(require('./RelayQueryVisitor'));

module.exports = RelayCacheProcessor;