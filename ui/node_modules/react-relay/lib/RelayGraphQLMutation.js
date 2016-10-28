/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayGraphQLMutation
 * 
 */

'use strict';

var _defineProperty3 = _interopRequireDefault(require('babel-runtime/helpers/defineProperty'));

var _extends5 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var CLIENT_MUTATION_ID = require('./RelayConnectionInterface').CLIENT_MUTATION_ID;

var COUNTER_PREFIX = 'RelayGraphQLMutation';
var collisionIDCounter = 0;

/**
 * @internal
 *
 * Low-level API for modeling a GraphQL mutation.
 *
 * This is the lowest level of abstraction at which product code may deal with
 * mutations in Relay, and it corresponds to the mutation operation ("a write
 * followed by a fetch") described in the GraphQL Specification. You specify
 * the mutation, the inputs, and the query.
 *
 * (There is an even lower-level representation, `RelayMutationRequest`,
 * underlying this which is an entirely internal implementation detail that
 * product code need not be aware of.)
 *
 * Low-level mutations cannot (yet) be applied optimistically or rolled back.
 * They don't provide any bells and whistles such as fat queries or tracked
 * queries. If you want those, you can opt in to the higher-level
 * `RelayMutation` API.
 *
 * @see http://facebook.github.io/graphql/.
 *
 */

var RelayGraphQLMutation = function () {

  /**
   * Simplest method for creating a RelayGraphQLMutation instance from a static
   * `mutation`, some `variables` and an `environment`.
   */

  RelayGraphQLMutation.create = function create(mutation, variables, environment) {
    return new RelayGraphQLMutation(mutation, variables, null, environment);
  };

  /**
   * Specialized method for creating RelayGraphQLMutation instances that takes a
   * `files` object in addition to the base `mutation`, `variables` and
   * `environment` parameters.
   */


  RelayGraphQLMutation.createWithFiles = function createWithFiles(mutation, variables, files, environment) {
    return new RelayGraphQLMutation(mutation, variables, files, environment);
  };

  /**
   * General constructor for creating RelayGraphQLMutation instances with
   * optional `files`, `callbacks` and `collisionKey` arguments.
   *
   * Callers must provide an appropriate `mutation`:
   *
   *    Relay.QL`
   *      mutation StoryLikeQuery {
   *        likeStory(input: $input) {
   *          clientMutationId
   *          story {
   *            likeCount
   *            likers {
   *              actor {
   *                name
   *              }
   *            }
   *          }
   *        }
   *      }
   *    `;
   *
   * And set of `variables`:
   *
   *    {
   *      input: {
   *        feedbackId: 'aFeedbackId',
   *      },
   *    }
   *
   * As per the GraphQL Relay Specification:
   *
   * - The mutation should take a single argument named "input".
   * - That input argument should contain a (string) "clientMutationId" property
   *   for the purposes of reconciling requests and responses (automatically
   *   added by the RelayGraphQLMutation API).
   * - The query should request "clientMutationId" as a subselection.
   *
   * @see http://facebook.github.io/relay/docs/graphql-mutations.html
   * @see http://facebook.github.io/relay/graphql/mutations.htm
   *
   * If not supplied, a unique collision key is derived (meaning that the
   * created mutation will be independent and not collide with any other).
   */


  function RelayGraphQLMutation(query, variables, files, environment, callbacks, collisionKey) {
    (0, _classCallCheck3['default'])(this, RelayGraphQLMutation);

    this._query = query;
    this._variables = variables;
    this._files = files || null;
    this._environment = environment;
    this._callbacks = callbacks || null;
    this._collisionKey = collisionKey || COUNTER_PREFIX + ':collisionKey:' + getNextCollisionID();
    this._transaction = null;
  }

  /**
   * Call this to optimistically apply an update to the store.
   *
   * The optional `config` parameter can be used to configure a `RANGE_ADD` type
   * mutation, similar to `RelayMutation` API.
   *
   * Optionally, follow up with a call to `commit()` to send the mutation
   * to the server.
   *
   * Note: An optimistic update may only be applied once.
   */


  RelayGraphQLMutation.prototype.applyOptimistic = function applyOptimistic(optimisticQuery, optimisticResponse, configs) {
    require('fbjs/lib/invariant')(!this._transaction, 'RelayGraphQLMutation: `applyOptimistic()` was called on an instance ' + 'that already has a transaction in progress.');
    this._transaction = this._createTransaction(optimisticQuery, optimisticResponse);
    return this._transaction.applyOptimistic(configs);
  };

  /**
   * Call this to send the mutation to the server.
   *
   * The optional `config` parameter can be used to configure a `RANGE_ADD` type
   * mutation, similar to the `RelayMutation` API.
   *
   * Optionally, precede with a call to `applyOptimistic()` to apply an update
   * optimistically to the store.
   *
   * Note: This method may only be called once per instance.
   */


  RelayGraphQLMutation.prototype.commit = function commit(configs) {
    if (!this._transaction) {
      this._transaction = this._createTransaction();
    }
    return this._transaction.commit(configs);
  };

  RelayGraphQLMutation.prototype._createTransaction = function _createTransaction(optimisticQuery, optimisticResponse) {
    return new PendingGraphQLTransaction(this._environment, this._query, this._variables, this._files, optimisticQuery, optimisticResponse, this._collisionKey, this._callbacks);
  };

  return RelayGraphQLMutation;
}();

function getNextCollisionID() {
  return collisionIDCounter++;
}

/**
 * @internal
 *
 * Data structure conforming to the `PendingTransaction` interface specified by
 * `RelayMutationQueue`.
 */

var PendingGraphQLTransaction = function () {

  // Other properties:
  // These properties required to conform to the PendingTransaction interface:
  function PendingGraphQLTransaction(environment, query, variables, files, optimisticQuery, optimisticResponse, collisionKey, callbacks) {
    (0, _classCallCheck3['default'])(this, PendingGraphQLTransaction);

    this._configs = [];
    this._query = query;
    this._variables = variables;
    this._optimisticQuery = optimisticQuery || null;
    this._optimisticResponse = optimisticResponse || null;
    this._collisionKey = collisionKey;
    this.onFailure = callbacks && callbacks.onFailure;
    this.onSuccess = callbacks && callbacks.onSuccess;
    this.status = require('./RelayMutationTransactionStatus').CREATED;
    this.error = null;
    this._mutation = null;
    this._optimisticConfigs = null;
    this._optimisticMutation = null;

    this.mutationTransaction = environment.getStoreData().getMutationQueue().createTransactionWithPendingTransaction(this);

    this.id = this.mutationTransaction.getID();
  }

  // Methods from the PendingTransaction interface.

  PendingGraphQLTransaction.prototype.getCallName = function getCallName() {
    require('fbjs/lib/invariant')(this._mutation, 'RelayGraphQLMutation: `getCallName()` called but no mutation exists ' + '(`getQuery()` must be called first to construct the mutation).');
    return this._mutation.getCall().name;
  };

  PendingGraphQLTransaction.prototype.getCollisionKey = function getCollisionKey() {
    return this._collisionKey;
  };

  PendingGraphQLTransaction.prototype.getConfigs = function getConfigs() {
    return this._configs;
  };

  PendingGraphQLTransaction.prototype.getFiles = function getFiles() {
    return this._files;
  };

  PendingGraphQLTransaction.prototype.getOptimisticConfigs = function getOptimisticConfigs() {
    return this._optimisticConfigs;
  };

  PendingGraphQLTransaction.prototype.getOptimisticQuery = function getOptimisticQuery(storeData) {
    if (!this._optimisticMutation && this._optimisticQuery) {
      var concreteMutation = require('./QueryBuilder').getMutation(this._optimisticQuery);
      var mutation = require('./RelayQuery').Mutation.create(concreteMutation, require('./RelayMetaRoute').get('$RelayGraphQLMutation'), this._getVariables());
      this._optimisticMutation = mutation; // Cast RelayQuery.{Node -> Mutation}.
    }
    return this._optimisticMutation;
  };

  PendingGraphQLTransaction.prototype.getOptimisticResponse = function getOptimisticResponse() {
    return (0, _extends5['default'])({}, this._optimisticResponse, (0, _defineProperty3['default'])({}, CLIENT_MUTATION_ID, this.id));
  };

  PendingGraphQLTransaction.prototype.getQuery = function getQuery(storeData) {
    if (!this._mutation) {
      var concreteMutation = require('./QueryBuilder').getMutation(this._query);
      var mutation = require('./RelayQuery').Mutation.create(concreteMutation, require('./RelayMetaRoute').get('$RelayGraphQLMutation'), this._getVariables());
      this._mutation = mutation; // Cast RelayQuery.{Node -> Mutation}.
    }
    return this._mutation;
  };

  // Additional methods outside the PendingTransaction interface.

  PendingGraphQLTransaction.prototype.commit = function commit(configs) {
    if (configs) {
      this._configs = configs;
    }
    return this.mutationTransaction.commit();
  };

  PendingGraphQLTransaction.prototype.applyOptimistic = function applyOptimistic(configs) {
    if (configs) {
      this._optimisticConfigs = configs;
    }
    return this.mutationTransaction.applyOptimistic();
  };

  PendingGraphQLTransaction.prototype._getVariables = function _getVariables() {
    var input = this._variables.input;
    if (!input) {
      require('fbjs/lib/invariant')(false, 'RelayGraphQLMutation: Required `input` variable is missing ' + '(supplied variables were: [%s]).', (0, _keys2['default'])(this._variables).join(', '));
    }
    return (0, _extends5['default'])({}, this._variables, {
      input: (0, _extends5['default'])({}, input, (0, _defineProperty3['default'])({}, CLIENT_MUTATION_ID, this.id))
    });
  };

  return PendingGraphQLTransaction;
}();

module.exports = RelayGraphQLMutation;