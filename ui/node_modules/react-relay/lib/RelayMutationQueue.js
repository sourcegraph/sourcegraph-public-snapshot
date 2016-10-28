/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayMutationQueue
 * 
 */

'use strict';

var _defineProperty3 = _interopRequireDefault(require('babel-runtime/helpers/defineProperty'));

var _extends4 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var CLIENT_MUTATION_ID = require('./RelayConnectionInterface').CLIENT_MUTATION_ID;

var transactionIDCounter = 0;

/**
 * @internal
 *
 * Coordinates execution of concurrent mutations, including application and
 * rollback of optimistic payloads and enqueueing mutations with the same
 * collision key.
 */

var RelayMutationQueue = function () {
  function RelayMutationQueue(storeData) {
    (0, _classCallCheck3['default'])(this, RelayMutationQueue);

    this._collisionQueueMap = {};
    this._pendingTransactionMap = {};
    this._queue = [];
    this._storeData = storeData;
    this._willBatchRefreshQueuedData = false;
  }

  /**
   * High-level API for creating a RelayMutationTransaction from a
   * RelayMutation.
   */


  RelayMutationQueue.prototype.createTransaction = function createTransaction(mutation, callbacks) {
    return this.createTransactionWithPendingTransaction(null, function (id, mutationTransaction) {
      return new RelayPendingTransaction({
        id: id,
        mutation: mutation,
        mutationTransaction: mutationTransaction,
        onFailure: callbacks && callbacks.onFailure,
        onSuccess: callbacks && callbacks.onSuccess
      });
    });
  };

  /**
   * @internal
   *
   * This is a lower-level API used to create transactions based on:
   *
   * - An object that conforms to the PendingTransaction type; or
   * - A function that can build such an object.
   *
   * Used by the high-level `createTransaction` API, but also enables us to
   * run legacy and low-level mutations.
   */


  RelayMutationQueue.prototype.createTransactionWithPendingTransaction = function createTransactionWithPendingTransaction(pendingTransaction, transactionBuilder) {
    require('fbjs/lib/invariant')(pendingTransaction || transactionBuilder, 'RelayMutationQueue: `createTransactionWithPendingTransaction()` ' + 'expects a PendingTransaction or TransactionBuilder.');
    var id = getNextID();
    var mutationTransaction = new (require('./RelayMutationTransaction'))(this, id);
    var transaction = pendingTransaction || transactionBuilder(id, mutationTransaction);
    this._pendingTransactionMap[id] = transaction;
    this._queue.push(transaction);
    return mutationTransaction;
  };

  RelayMutationQueue.prototype.getTransaction = function getTransaction(id) {
    var transaction = this._pendingTransactionMap[id];
    if (transaction) {
      return transaction.mutationTransaction;
    }
    return null;
  };

  RelayMutationQueue.prototype.getError = function getError(id) {
    return this._get(id).error;
  };

  RelayMutationQueue.prototype.getStatus = function getStatus(id) {
    return this._get(id).status;
  };

  RelayMutationQueue.prototype.applyOptimistic = function applyOptimistic(id) {
    var transaction = this._get(id);
    transaction.status = require('./RelayMutationTransactionStatus').UNCOMMITTED;
    transaction.error = null;
    this._handleOptimisticUpdate(transaction);
  };

  RelayMutationQueue.prototype.commit = function commit(id) {
    var transaction = this._get(id);
    var collisionKey = transaction.getCollisionKey();
    var collisionQueue = collisionKey && this._collisionQueueMap[collisionKey];
    if (collisionQueue) {
      collisionQueue.push(transaction);
      transaction.status = require('./RelayMutationTransactionStatus').COMMIT_QUEUED;
      transaction.error = null;
      return;
    }
    if (collisionKey) {
      this._collisionQueueMap[collisionKey] = [transaction];
    }
    this._handleCommit(transaction);
  };

  RelayMutationQueue.prototype.rollback = function rollback(id) {
    var transaction = this._get(id);
    var collisionKey = transaction.getCollisionKey();
    if (collisionKey) {
      var collisionQueue = this._collisionQueueMap[collisionKey];
      if (collisionQueue) {
        var index = collisionQueue.indexOf(transaction);
        if (index !== -1) {
          collisionQueue.splice(index, 1);
        }
        if (collisionQueue.length === 0) {
          delete this._collisionQueueMap[collisionKey];
        }
      }
    }
    this._handleRollback(transaction);
  };

  RelayMutationQueue.prototype._get = function _get(id) {
    var transaction = this._pendingTransactionMap[id];
    require('fbjs/lib/invariant')(transaction, 'RelayMutationQueue: `%s` is not a valid pending transaction ID.', id);
    return transaction;
  };

  RelayMutationQueue.prototype._handleOptimisticUpdate = function _handleOptimisticUpdate(transaction) {
    var optimisticResponse = transaction.getOptimisticResponse();
    var optimisticQuery = transaction.getOptimisticQuery(this._storeData);
    if (optimisticResponse && optimisticQuery) {
      var configs = transaction.getOptimisticConfigs() || transaction.getConfigs();
      this._storeData.handleUpdatePayload(optimisticQuery, optimisticResponse, {
        configs: configs,
        isOptimisticUpdate: true
      });
    }
  };

  RelayMutationQueue.prototype._handleCommitFailure = function _handleCommitFailure(transaction, error) {
    var status = error ? require('./RelayMutationTransactionStatus').COMMIT_FAILED : require('./RelayMutationTransactionStatus').COLLISION_COMMIT_FAILED;
    transaction.status = status;
    transaction.error = error;

    var shouldRollback = true;
    var onFailure = transaction.onFailure;
    if (onFailure) {
      var preventAutoRollback = function preventAutoRollback() {
        shouldRollback = false;
      };
      require('fbjs/lib/ErrorUtils').applyWithGuard(onFailure, null, [transaction.mutationTransaction, preventAutoRollback], null, 'RelayMutationTransaction:onCommitFailure');
    }

    if (error) {
      this._failCollisionQueue(transaction);
    }

    // Might have already been rolled back via `onFailure`.
    if (shouldRollback && this._pendingTransactionMap.hasOwnProperty(transaction.id)) {
      this._handleRollback(transaction);
    }
    this._batchRefreshQueuedData();
  };

  RelayMutationQueue.prototype._handleCommitSuccess = function _handleCommitSuccess(transaction, response) {
    this._advanceCollisionQueue(transaction);
    this._clearPendingTransaction(transaction);

    this._refreshQueuedData();
    this._storeData.handleUpdatePayload(transaction.getQuery(this._storeData), response[transaction.getCallName()], {
      configs: transaction.getConfigs(),
      isOptimisticUpdate: false
    });

    var onSuccess = transaction.onSuccess;
    if (onSuccess) {
      require('fbjs/lib/ErrorUtils').applyWithGuard(onSuccess, null, [response], null, 'RelayMutationTransaction:onCommitSuccess');
    }
  };

  RelayMutationQueue.prototype._handleCommit = function _handleCommit(transaction) {
    var _this = this;

    transaction.status = require('./RelayMutationTransactionStatus').COMMITTING;
    transaction.error = null;

    var request = new (require('./RelayMutationRequest'))(transaction.getQuery(this._storeData), transaction.getFiles());
    this._storeData.getNetworkLayer().sendMutation(request);

    request.done(function (result) {
      return _this._handleCommitSuccess(transaction, result.response);
    }, function (error) {
      return _this._handleCommitFailure(transaction, error);
    });
  };

  RelayMutationQueue.prototype._handleRollback = function _handleRollback(transaction) {
    this._clearPendingTransaction(transaction);
    this._batchRefreshQueuedData();
  };

  RelayMutationQueue.prototype._clearPendingTransaction = function _clearPendingTransaction(transaction) {
    delete this._pendingTransactionMap[transaction.id];
    this._queue = this._queue.filter(function (tx) {
      return tx !== transaction;
    });
  };

  RelayMutationQueue.prototype._advanceCollisionQueue = function _advanceCollisionQueue(transaction) {
    var collisionKey = transaction.getCollisionKey();
    if (collisionKey) {
      var collisionQueue = this._collisionQueueMap[collisionKey];
      if (collisionQueue) {
        // Remove the transaction that called this function.
        collisionQueue.shift();

        if (collisionQueue.length) {
          this._handleCommit(collisionQueue[0]);
        } else {
          delete this._collisionQueueMap[collisionKey];
        }
      }
    }
  };

  RelayMutationQueue.prototype._failCollisionQueue = function _failCollisionQueue(failedTransaction) {
    var _this2 = this;

    var collisionKey = failedTransaction.getCollisionKey();
    if (collisionKey) {
      var collisionQueue = this._collisionQueueMap[collisionKey];
      if (collisionQueue) {
        // Remove the transaction that called this function.
        collisionQueue.forEach(function (queuedTransaction) {
          if (queuedTransaction !== failedTransaction) {
            _this2._handleCommitFailure(queuedTransaction, null);
          }
        });
        delete this._collisionQueueMap[collisionKey];
      }
    }
  };

  RelayMutationQueue.prototype._batchRefreshQueuedData = function _batchRefreshQueuedData() {
    var _this3 = this;

    if (!this._willBatchRefreshQueuedData) {
      this._willBatchRefreshQueuedData = true;
      require('fbjs/lib/resolveImmediate')(function () {
        _this3._willBatchRefreshQueuedData = false;
        _this3._refreshQueuedData();
      });
    }
  };

  RelayMutationQueue.prototype._refreshQueuedData = function _refreshQueuedData() {
    var _this4 = this;

    this._storeData.clearQueuedData();
    this._queue.forEach(function (transaction) {
      return _this4._handleOptimisticUpdate(transaction);
    });
  };

  return RelayMutationQueue;
}();

/**
 * @private
 */


var RelayPendingTransaction = function () {
  function RelayPendingTransaction(transactionData) {
    (0, _classCallCheck3['default'])(this, RelayPendingTransaction);

    this.error = null;
    this.id = transactionData.id;
    this.mutation = transactionData.mutation;
    this.mutationTransaction = transactionData.mutationTransaction;
    this.onFailure = transactionData.onFailure;
    this.onSuccess = transactionData.onSuccess;
    this.status = require('./RelayMutationTransactionStatus').CREATED;
  }

  // Lazily computed and memoized private properties


  RelayPendingTransaction.prototype.getCallName = function getCallName() {
    if (!this._callName) {
      this._callName = this.getMutationNode().calls[0].name;
    }
    return this._callName;
  };

  RelayPendingTransaction.prototype.getCollisionKey = function getCollisionKey() {
    if (this._collisionKey === undefined) {
      this._collisionKey = this.mutation.getCollisionKey() || null;
    }
    return this._collisionKey;
  };

  RelayPendingTransaction.prototype.getConfigs = function getConfigs() {
    if (!this._configs) {
      this._configs = this.mutation.getConfigs();
    }
    return this._configs;
  };

  RelayPendingTransaction.prototype.getFatQuery = function getFatQuery() {
    if (!this._fatQuery) {
      var fragment = require('./fromGraphQL').Fragment(this.mutation.getFatQuery());
      require('fbjs/lib/invariant')(fragment instanceof require('./RelayQuery').Fragment, 'RelayMutationQueue: Expected `getFatQuery` to return a GraphQL ' + 'Fragment');
      this._fatQuery = require('fbjs/lib/nullthrows')(require('./flattenRelayQuery')(fragment, {
        // TODO #10341736
        // This used to be `preserveEmptyNodes: fragment.isPattern()`. We
        // discovered that products were not marking their fat queries as
        // patterns (by adding `@relay(pattern: true)`) which was causing
        // `preserveEmptyNodes` to be false. This meant that empty fields,
        // would be stripped instead of being used to produce an intersection
        // with the tracked query. Products were able to do this because the
        // Babel Relay plugin doesn't produce validation errors for empty
        // fields. It should, and we will make it do so, but for now we're
        // going to set this to `true` always, and make the plugin warn when
        // it encounters an empty field that supports subselections in a
        // non-pattern fragment. Revert this when done.
        preserveEmptyNodes: true,
        shouldRemoveFragments: true
      }));
    }
    return this._fatQuery;
  };

  RelayPendingTransaction.prototype.getFiles = function getFiles() {
    if (this._files === undefined) {
      this._files = this.mutation.getFiles() || null;
    }
    return this._files;
  };

  RelayPendingTransaction.prototype.getInputVariable = function getInputVariable() {
    if (!this._inputVariable) {
      var inputVariable = (0, _extends4['default'])({}, this.mutation.getVariables(), (0, _defineProperty3['default'])({}, CLIENT_MUTATION_ID, this.id));
      this._inputVariable = inputVariable;
    }
    return this._inputVariable;
  };

  RelayPendingTransaction.prototype.getMutationNode = function getMutationNode() {
    if (!this._mutationNode) {
      var mutationNode = require('./QueryBuilder').getMutation(this.mutation.getMutation());
      require('fbjs/lib/invariant')(mutationNode, 'RelayMutation: Expected `getMutation` to return a mutation created ' + 'with Relay.QL`mutation { ... }`.');
      this._mutationNode = mutationNode;
    }
    return this._mutationNode;
  };

  RelayPendingTransaction.prototype.getOptimisticConfigs = function getOptimisticConfigs() {
    if (this._optimisticConfigs === undefined) {
      this._optimisticConfigs = this.mutation.getOptimisticConfigs() || null;
    }
    return this._optimisticConfigs;
  };

  RelayPendingTransaction.prototype.getOptimisticQuery = function getOptimisticQuery(storeData) {
    if (this._optimisticQuery === undefined) {
      /* eslint-disable no-console */
      if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
        console.groupCollapsed('Optimistic query for `' + this.getCallName() + '`');
      }
      /* eslint-enable no-console */
      var optimisticResponse = this._getRawOptimisticResponse();
      if (optimisticResponse) {
        var optimisticConfigs = this.getOptimisticConfigs();
        var tracker = getTracker(storeData);
        if (optimisticConfigs) {
          this._optimisticQuery = require('./RelayMutationQuery').buildQuery({
            configs: optimisticConfigs,
            fatQuery: this.getFatQuery(),
            input: this.getInputVariable(),
            mutationName: this.mutation.constructor.name,
            mutation: this.getMutationNode(),
            tracker: tracker
          });
        } else {
          this._optimisticQuery = require('./RelayMutationQuery').buildQueryForOptimisticUpdate({
            response: optimisticResponse,
            fatQuery: this.getFatQuery(),
            mutation: this.getMutationNode()
          });
        }
      } else {
        this._optimisticQuery = null;
      }
      /* eslint-disable no-console */
      if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
        require('./RelayMutationDebugPrinter').printOptimisticMutation(this._optimisticQuery, optimisticResponse);

        console.groupEnd();
      }
      /* eslint-enable no-console */
    }
    return this._optimisticQuery;
  };

  RelayPendingTransaction.prototype._getRawOptimisticResponse = function _getRawOptimisticResponse() {
    if (this._rawOptimisticResponse === undefined) {
      var optimisticResponse = this.mutation.getOptimisticResponse() || null;
      if (optimisticResponse) {
        optimisticResponse[CLIENT_MUTATION_ID] = this.id;
      }
      this._rawOptimisticResponse = optimisticResponse;
    }
    return this._rawOptimisticResponse;
  };

  RelayPendingTransaction.prototype.getOptimisticResponse = function getOptimisticResponse() {
    if (this._optimisticResponse === undefined) {
      var optimisticResponse = this._getRawOptimisticResponse();
      if (optimisticResponse) {
        this._optimisticResponse = require('./RelayOptimisticMutationUtils').inferRelayPayloadFromData(optimisticResponse);
      } else {
        this._optimisticResponse = optimisticResponse;
      }
    }
    return this._optimisticResponse;
  };

  RelayPendingTransaction.prototype.getQuery = function getQuery(storeData) {
    if (!this._query) {
      /* eslint-disable no-console */
      if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
        console.groupCollapsed('Mutation query for `' + this.getCallName() + '`');
      }
      /* eslint-enable no-console */
      var tracker = getTracker(storeData);
      this._query = require('./RelayMutationQuery').buildQuery({
        configs: this.getConfigs(),
        fatQuery: this.getFatQuery(),
        input: this.getInputVariable(),
        mutationName: this.getMutationNode().name,
        mutation: this.getMutationNode(),
        tracker: tracker
      });
      /* eslint-disable no-console */
      if (process.env.NODE_ENV !== 'production' && console.groupCollapsed && console.groupEnd) {
        require('./RelayMutationDebugPrinter').printMutation(this._query);
        console.groupEnd();
      }
      /* eslint-enable no-console */
    }
    return this._query;
  };

  return RelayPendingTransaction;
}();

function getTracker(storeData) {
  var tracker = storeData.getQueryTracker();
  require('fbjs/lib/invariant')(tracker, 'RelayMutationQueue: a RelayQueryTracker was not configured but an ' + 'attempt to process a RelayMutation instance was made. Either use ' + 'RelayGraphQLMutation (which does not require a tracker), or use ' + '`RelayStoreData.injectQueryTracker()` to configure a tracker. Be ' + 'aware that trackers are provided by default, so if you are seeing ' + 'this error it means that somebody has explicitly intended to opt ' + 'out of query tracking.');
  return tracker;
}

function getNextID() {
  return require('fbjs/lib/base62')(transactionIDCounter++);
}

module.exports = RelayMutationQueue;