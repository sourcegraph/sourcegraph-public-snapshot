/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayQuery
 * 
 */

'use strict';

/* eslint-disable consistent-this */

var _extends3 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _freeze2 = _interopRequireDefault(require('babel-runtime/core-js/object/freeze'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

// conditional field calls/values

// TODO: replace once #6525923 is resolved
var IF = 'if';
var UNLESS = 'unless';
var TRUE = 'true';
var FALSE = 'false';
var SKIP = 'skip';
var INCLUDE = 'include';

var _nextQueryID = 0;

var DEFAULT_FRAGMENT_METADATA = {
  isDeferred: false,
  isContainerFragment: false,
  isTypeConditional: false
};
var EMPTY_DIRECTIVES = [];
var EMPTY_CALLS = [];

if (process.env.NODE_ENV !== 'production') {
  (0, _freeze2['default'])(EMPTY_CALLS);
  (0, _freeze2['default'])(EMPTY_DIRECTIVES);
}

/**
 * @internal
 *
 * Queries in Relay are represented as trees. Possible nodes include the root,
 * fields, and fragments. Fields can have children, or they can be leaf nodes.
 * Root and fragment nodes must always have children.
 *
 * `RelayQueryNode` provides access to information such as the field name,
 * generated alias, sub-fields, and call values.
 *
 * Nodes are immutable; query modification is supported via `clone`:
 *
 * ```
 * var next = prev.clone(prev.getChildren().filter(f => ...));
 * ```
 *
 * Note: Mediating access to actual query nodes is necessary so that we can
 * replace the current mutable GraphQL nodes with an immutable query
 * representation. This class *must not* mutate the underlying `concreteNode`.
 * Instead, use an instance variable (see `clone()`).
 *
 * TODO (#6937314): RelayQueryNode support for toJSON/fromJSON
 */

var RelayQueryNode = function () {
  // for flow

  RelayQueryNode.create = function create(concreteNode, route, variables) {
    var node = createNode(concreteNode, route, variables);
    require('fbjs/lib/invariant')(node instanceof RelayQueryNode, 'RelayQueryNode.create(): ' + 'Expected a GraphQL fragment, mutation, or query.');
    return node;
  };

  /**
   * @private
   *
   * Base class for all node types, must not be directly instantiated.
   */


  function RelayQueryNode(concreteNode, route, variables) {
    (0, _classCallCheck3['default'])(this, RelayQueryNode);

    require('fbjs/lib/invariant')(this.constructor.name !== 'RelayQueryNode', 'RelayQueryNode: Abstract class cannot be instantiated.');
    this.__concreteNode__ = concreteNode;
    this.__route__ = route;
    this.__variables__ = variables;

    // lazily computed properties
    this.__calls__ = null;
    this.__children__ = null;
    this.__fieldMap__ = null;
    this.__hasDeferredDescendant__ = null;
    this.__hasValidatedConnectionCalls__ = null;
    this.__serializationKey__ = null;
    this.__storageKey__ = null;
  }

  RelayQueryNode.prototype.canHaveSubselections = function canHaveSubselections() {
    return true;
  };

  RelayQueryNode.prototype.isGenerated = function isGenerated() {
    return false;
  };

  RelayQueryNode.prototype.isRefQueryDependency = function isRefQueryDependency() {
    return false;
  };

  RelayQueryNode.prototype.clone = function clone(children) {
    if (!this.canHaveSubselections()) {
      // Compact new children *after* this check, for consistency.
      require('fbjs/lib/invariant')(children.length === 0, 'RelayQueryNode: Cannot add children to field `%s` because it does ' + 'not support sub-selections (sub-fields).', this instanceof RelayQueryField ? this.getSchemaName() : null);
      return this;
    }

    var prevChildren = this.getChildren();
    var nextChildren = cloneChildren(prevChildren, children);

    if (!nextChildren.length) {
      return null;
    } else if (nextChildren === prevChildren) {
      return this;
    }

    var clone = RelayQueryNode.create(this.__concreteNode__, this.__route__, this.__variables__);
    clone.__children__ = nextChildren;
    clone.__calls__ = this.__calls__;
    clone.__serializationKey__ = this.__serializationKey__;
    clone.__storageKey__ = this.__storageKey__;

    return clone;
  };

  RelayQueryNode.prototype.getChildren = function getChildren() {
    var _this = this;

    var children = this.__children__;
    if (!children) {
      (function () {
        var nextChildren = [];
        var concreteChildren = _this.__concreteNode__.children;
        if (concreteChildren) {
          concreteChildren.forEach(function (concreteChild) {
            if (concreteChild == null) {
              return;
            }
            var nodeOrNodes = createNode(concreteChild, _this.__route__, _this.__variables__);
            if (Array.isArray(nodeOrNodes)) {
              nodeOrNodes.forEach(function (node) {
                if (node && node.isIncluded()) {
                  nextChildren.push(node);
                }
              });
            } else if (nodeOrNodes && nodeOrNodes.isIncluded()) {
              nextChildren.push(nodeOrNodes);
            }
          });
        }
        _this.__children__ = nextChildren;
        children = nextChildren;
      })();
    }
    return children;
  };

  RelayQueryNode.prototype.isIncluded = function isIncluded() {
    // Bail out early since most nodes won't have directives
    if (!this.__concreteNode__.directives) {
      return true;
    }
    return this.getDirectives().every(function (directive) {
      if (directive.name === SKIP) {
        return !directive.args.some(function (arg) {
          return arg.name === IF && !!arg.value;
        });
      } else if (directive.name === INCLUDE) {
        return !directive.args.some(function (arg) {
          return arg.name === IF && !arg.value;
        });
      }
      return true;
    });
  };

  RelayQueryNode.prototype.getDirectives = function getDirectives() {
    var _this2 = this;

    var concreteDirectives = this.__concreteNode__.directives;
    if (concreteDirectives) {
      return this.__concreteNode__.directives.map(function (directive) {
        return {
          args: require('./callsFromGraphQL')(directive.args, _this2.__variables__),
          name: directive.name
        };
      });
    }
    return EMPTY_DIRECTIVES;
  };

  RelayQueryNode.prototype.getField = function getField(field) {
    return this.getFieldByStorageKey(field.getStorageKey());
  };

  RelayQueryNode.prototype.getFieldByStorageKey = function getFieldByStorageKey(storageKey) {
    var fieldMap = this.__fieldMap__;
    if (!fieldMap) {
      fieldMap = {};
      var children = this.getChildren();
      for (var ii = 0; ii < children.length; ii++) {
        var child = children[ii];
        if (child instanceof RelayQueryField) {
          fieldMap[child.getStorageKey()] = child;
        }
      }
      this.__fieldMap__ = fieldMap;
    }
    return fieldMap[storageKey];
  };

  RelayQueryNode.prototype.getType = function getType() {
    return this.__concreteNode__.type;
  };

  RelayQueryNode.prototype.getRoute = function getRoute() {
    return this.__route__;
  };

  RelayQueryNode.prototype.getVariables = function getVariables() {
    return this.__variables__;
  };

  RelayQueryNode.prototype.hasDeferredDescendant = function hasDeferredDescendant() {
    var hasDeferredDescendant = this.__hasDeferredDescendant__;
    if (hasDeferredDescendant == null) {
      hasDeferredDescendant = this.canHaveSubselections() && this.getChildren().some(function (child) {
        return child.hasDeferredDescendant();
      });
      this.__hasDeferredDescendant__ = hasDeferredDescendant;
    }
    return hasDeferredDescendant;
  };

  RelayQueryNode.prototype.isAbstract = function isAbstract() {
    throw new Error('RelayQueryNode: Abstract function cannot be called.');
  };

  RelayQueryNode.prototype.isRequisite = function isRequisite() {
    return false;
  };

  /**
   * Determine if `this` and `that` are deeply equal.
   */


  RelayQueryNode.prototype.equals = function equals(that) {
    var thisChildren = this.getChildren();
    var thatChildren = that.getChildren();

    return thisChildren === thatChildren || thisChildren.length === thatChildren.length && thisChildren.every(function (c, ii) {
      return c.equals(thatChildren[ii]);
    });
  };

  /**
   * Performs a fast comparison of whether `this` and `that` represent identical
   * query nodes. Returns true only if the concrete nodes, routes, and variables
   * are all the same.
   *
   * Note that it is possible that this method can return false in cases where
   * `equals` would return true. This can happen when the concrete nodes are
   * different but structurally identical, or when the route/variables are
   * different but do not have an effect on the structure of the query.
   */


  RelayQueryNode.prototype.isEquivalent = function isEquivalent(that) {
    return this.__concreteNode__ === that.__concreteNode__ && this.__route__ === that.__route__ && require('fbjs/lib/shallowEqual')(this.__variables__, that.__variables__);
  };

  RelayQueryNode.prototype.createNode = function createNode(concreteNode) {
    return RelayQueryNode.create(concreteNode, this.__route__, this.__variables__);
  };

  RelayQueryNode.prototype.getConcreteQueryNode = function getConcreteQueryNode() {
    return this.__concreteNode__;
  };

  return RelayQueryNode;
}();

/**
 * @internal
 *
 * Wraps access to query root nodes.
 */


var RelayQueryRoot = function (_RelayQueryNode) {
  (0, _inherits3['default'])(RelayQueryRoot, _RelayQueryNode);


  /**
   * Helper to construct a new root query with the given attributes and 'empty'
   * route/variables.
   */

  RelayQueryRoot.build = function build(name, fieldName, value, children, metadata, type, routeName) {
    var nextChildren = children ? children.filter(function (child) {
      return !!child;
    }) : [];
    var batchCallVariable = require('./QueryBuilder').getBatchCallVariable(value);
    var identifyingArgValue = void 0;
    if (batchCallVariable) {
      identifyingArgValue = batchCallVariable;
    } else if (Array.isArray(value)) {
      identifyingArgValue = value.map(require('./QueryBuilder').createCallValue);
    } else if (value) {
      identifyingArgValue = require('./QueryBuilder').createCallValue(value);
    }
    var concreteRoot = require('./QueryBuilder').createQuery({
      fieldName: fieldName,
      identifyingArgValue: identifyingArgValue,
      metadata: metadata,
      name: name,
      type: type
    });
    var root = new RelayQueryRoot(concreteRoot, require('./RelayMetaRoute').get(routeName || '$RelayQuery'), {});
    root.__children__ = nextChildren;
    return root;
  };

  RelayQueryRoot.create = function create(concreteNode, route, variables) {
    var query = require('./QueryBuilder').getQuery(concreteNode);
    require('fbjs/lib/invariant')(query, 'RelayQueryRoot.create(): Expected a GraphQL `query { ... }`, got: %s', concreteNode);
    return new RelayQueryRoot(query, route, variables);
  };

  function RelayQueryRoot(concreteNode, route, variables) {
    (0, _classCallCheck3['default'])(this, RelayQueryRoot);

    var _this3 = (0, _possibleConstructorReturn3['default'])(this, _RelayQueryNode.call(this, concreteNode, route, variables));

    _this3.__batchCall__ = undefined;
    _this3.__id__ = undefined;
    _this3.__identifyingArg__ = undefined;
    _this3.__storageKey__ = undefined;

    // Ensure IDs are generated in the order that queries are created
    _this3.getID();
    return _this3;
  }

  RelayQueryRoot.prototype.canHaveSubselections = function canHaveSubselections() {
    return true;
  };

  RelayQueryRoot.prototype.getName = function getName() {
    var name = this.__concreteNode__.name;
    if (!name) {
      name = this.getID();
      this.__concreteNode__.name = name;
    }
    return name;
  };

  RelayQueryRoot.prototype.getID = function getID() {
    var id = this.__id__;
    if (id == null) {
      id = 'q' + _nextQueryID++;
      this.__id__ = id;
    }
    return id;
  };

  RelayQueryRoot.prototype.getBatchCall = function getBatchCall() {
    var batchCall = this.__batchCall__;
    if (batchCall === undefined) {
      var concreteCalls = this.__concreteNode__.calls;
      if (concreteCalls) {
        var callArg = concreteCalls[0] && concreteCalls[0].value;
        if (callArg != null && !Array.isArray(callArg) && callArg.kind === 'BatchCallVariable') {
          batchCall = {
            refParamName: 'ref_' + callArg.sourceQueryID,
            sourceQueryID: callArg.sourceQueryID,
            sourceQueryPath: callArg.jsonPath
          };
        }
      }
      batchCall = batchCall || null;
      this.__batchCall__ = batchCall;
    }
    return batchCall;
  };

  RelayQueryRoot.prototype.getCallsWithValues = function getCallsWithValues() {
    var calls = this.__calls__;
    if (!calls) {
      var concreteCalls = this.__concreteNode__.calls;
      if (concreteCalls) {
        calls = require('./callsFromGraphQL')(concreteCalls, this.__variables__);
      } else {
        calls = EMPTY_CALLS;
      }
      this.__calls__ = calls;
    }
    return calls;
  };

  RelayQueryRoot.prototype.getFieldName = function getFieldName() {
    return this.__concreteNode__.fieldName;
  };

  RelayQueryRoot.prototype.getIdentifyingArg = function getIdentifyingArg() {
    var _this4 = this;

    var identifyingArg = this.__identifyingArg__;
    if (!identifyingArg) {
      (function () {
        var metadata = _this4.__concreteNode__.metadata;
        var identifyingArgName = metadata.identifyingArgName;
        if (identifyingArgName != null) {
          identifyingArg = _this4.getCallsWithValues().find(function (c) {
            return c.name === identifyingArgName;
          });
          if (identifyingArg && metadata.identifyingArgType != null) {
            identifyingArg.type = metadata.identifyingArgType;
          }
          _this4.__identifyingArg__ = identifyingArg;
        }
      })();
    }
    return identifyingArg;
  };

  RelayQueryRoot.prototype.getStorageKey = function getStorageKey() {
    var _this5 = this;

    var storageKey = this.__storageKey__;
    if (!storageKey) {
      (function () {
        var args = _this5.getCallsWithValues();
        var identifyingArg = _this5.getIdentifyingArg();
        if (identifyingArg) {
          args = args.filter(function (arg) {
            return arg !== identifyingArg;
          });
        }
        var field = RelayQueryField.build({
          fieldName: _this5.getFieldName(),
          calls: args,
          type: _this5.getType()
        });
        storageKey = field.getStorageKey();
        _this5.__storageKey__ = storageKey;
      })();
    }
    return storageKey;
  };

  RelayQueryRoot.prototype.hasDeferredDescendant = function hasDeferredDescendant() {
    return this.isDeferred() || _RelayQueryNode.prototype.hasDeferredDescendant.call(this);
  };

  RelayQueryRoot.prototype.isAbstract = function isAbstract() {
    return !!this.__concreteNode__.metadata.isAbstract;
  };

  RelayQueryRoot.prototype.isDeferred = function isDeferred() {
    return !!this.__concreteNode__.isDeferred;
  };

  RelayQueryRoot.prototype.isPlural = function isPlural() {
    return !!this.__concreteNode__.metadata.isPlural;
  };

  RelayQueryRoot.prototype.cloneWithRoute = function cloneWithRoute(children, route) {
    if (this.__route__ === route) {
      return this.clone(children);
    }
    var clone = RelayQueryNode.create((0, _extends3['default'])({}, this.__concreteNode__, {
      name: route.name
    }), route, this.__variables__);
    clone.__children__ = children;
    return clone;
  };

  RelayQueryRoot.prototype.equals = function equals(that) {
    if (this === that) {
      return true;
    }
    if (!(that instanceof RelayQueryRoot)) {
      return false;
    }
    if (!require('fbjs/lib/areEqual')(this.getBatchCall(), that.getBatchCall())) {
      return false;
    }
    if (this.getFieldName() !== that.getFieldName() || !areCallValuesEqual(this.getCallsWithValues(), that.getCallsWithValues())) {
      return false;
    }
    return _RelayQueryNode.prototype.equals.call(this, that);
  };

  return RelayQueryRoot;
}(RelayQueryNode);

/**
 * @internal
 *
 * Abstract base class for mutations and subscriptions.
 */


var RelayQueryOperation = function (_RelayQueryNode2) {
  (0, _inherits3['default'])(RelayQueryOperation, _RelayQueryNode2);

  function RelayQueryOperation(concreteNode, route, variables) {
    (0, _classCallCheck3['default'])(this, RelayQueryOperation);

    var _this6 = (0, _possibleConstructorReturn3['default'])(this, _RelayQueryNode2.call(this, concreteNode, route, variables));

    require('fbjs/lib/invariant')(_this6.constructor.name !== 'RelayQueryOperation', 'RelayQueryOperation: Abstract class cannot be instantiated.');
    return _this6;
  }

  RelayQueryOperation.prototype.canHaveSubselections = function canHaveSubselections() {
    return true;
  };

  RelayQueryOperation.prototype.getName = function getName() {
    return this.__concreteNode__.name;
  };

  RelayQueryOperation.prototype.getResponseType = function getResponseType() {
    return this.__concreteNode__.responseType;
  };

  RelayQueryOperation.prototype.getType = function getType() {
    return this.getResponseType();
  };

  RelayQueryOperation.prototype.getInputType = function getInputType() {
    var inputType = this.__concreteNode__.metadata.inputType;
    require('fbjs/lib/invariant')(inputType, 'RelayQuery: Expected operation `%s` to be annotated with the type of ' + 'its argument. Either the babel transform was configured incorrectly, ' + 'or the schema failed to define an argument for this mutation.', this.getCall().name);
    return inputType;
  };

  RelayQueryOperation.prototype.getCall = function getCall() {
    var calls = this.__calls__;
    if (!calls) {
      var concreteCalls = this.__concreteNode__.calls;
      if (concreteCalls) {
        calls = require('./callsFromGraphQL')(concreteCalls, this.__variables__);
      } else {
        calls = EMPTY_CALLS;
      }
      this.__calls__ = calls;
    }
    return calls[0];
  };

  RelayQueryOperation.prototype.getCallVariableName = function getCallVariableName() {
    if (!this.__callVariableName__) {
      var concreteCalls = this.__concreteNode__.calls;
      var callVariable = concreteCalls && require('./QueryBuilder').getCallVariable(concreteCalls[0].value);
      require('fbjs/lib/invariant')(callVariable, 'RelayQuery: Expected mutation to have a single argument.');
      this.__callVariableName__ = callVariable.callVariableName;
    }
    return this.__callVariableName__;
  };

  /**
   * Mutations and subscriptions must have a concrete type due to the need for
   * requisite top-level fields.
   */


  RelayQueryOperation.prototype.isAbstract = function isAbstract() {
    return false;
  };

  return RelayQueryOperation;
}(RelayQueryNode);

/**
 * @internal
 *
 * Represents a GraphQL mutation.
 */


var RelayQueryMutation = function (_RelayQueryOperation) {
  (0, _inherits3['default'])(RelayQueryMutation, _RelayQueryOperation);

  function RelayQueryMutation() {
    (0, _classCallCheck3['default'])(this, RelayQueryMutation);
    return (0, _possibleConstructorReturn3['default'])(this, _RelayQueryOperation.apply(this, arguments));
  }

  /**
   * Helper to construct a new mutation with the given attributes and 'empty'
   * route/variables.
   */

  RelayQueryMutation.build = function build(name, responseType, callName, callValue, children, metadata, routeName) {
    var nextChildren = children ? children.filter(function (child) {
      return !!child;
    }) : [];
    var concreteMutation = require('./QueryBuilder').createMutation({
      calls: [require('./QueryBuilder').createCall(callName, require('./QueryBuilder').createCallVariable('input'))],
      metadata: metadata,
      name: name,
      responseType: responseType
    });
    var mutation = new RelayQueryMutation(concreteMutation, require('./RelayMetaRoute').get(routeName || '$RelayQuery'), { input: callValue || '' });
    mutation.__children__ = nextChildren;
    return mutation;
  };

  RelayQueryMutation.prototype.equals = function equals(that) {
    if (this === that) {
      return true;
    }
    if (!(that instanceof RelayQueryMutation)) {
      return false;
    }
    if (!require('fbjs/lib/areEqual')(this.getResponseType(), that.getResponseType())) {
      return false;
    }
    if (!require('fbjs/lib/areEqual')(this.getCall(), that.getCall())) {
      return false;
    }
    return _RelayQueryOperation.prototype.equals.call(this, that);
  };

  return RelayQueryMutation;
}(RelayQueryOperation);

/**
 * @internal
 *
 * Represents a GraphQL subscription.
 */


var RelayQuerySubscription = function (_RelayQueryOperation2) {
  (0, _inherits3['default'])(RelayQuerySubscription, _RelayQueryOperation2);

  function RelayQuerySubscription() {
    (0, _classCallCheck3['default'])(this, RelayQuerySubscription);
    return (0, _possibleConstructorReturn3['default'])(this, _RelayQueryOperation2.apply(this, arguments));
  }

  RelayQuerySubscription.create = function create(concreteNode, route, variables) {
    var subscription = require('./QueryBuilder').getSubscription(concreteNode);
    require('fbjs/lib/invariant')(subscription, 'RelayQuerySubscription.create(): ' + 'Expected a GraphQL `subscription { ... }`, got: %s', concreteNode);
    return new RelayQuerySubscription(concreteNode, route, variables);
  };

  RelayQuerySubscription.prototype.getPublishedPayloadType = function getPublishedPayloadType() {
    return this.getResponseType();
  };

  RelayQuerySubscription.prototype.equals = function equals(that) {
    if (this === that) {
      return true;
    }
    if (!(that instanceof RelayQuerySubscription)) {
      return false;
    }
    if (!require('fbjs/lib/areEqual')(this.getPublishedPayloadType(), that.getPublishedPayloadType())) {
      return false;
    }
    if (!require('fbjs/lib/areEqual')(this.getCall(), that.getCall())) {
      return false;
    }
    return _RelayQueryOperation2.prototype.equals.call(this, that);
  };

  return RelayQuerySubscription;
}(RelayQueryOperation);

/**
 * @internal
 *
 * Wraps access to query fragments.
 */


var RelayQueryFragment = function (_RelayQueryNode3) {
  (0, _inherits3['default'])(RelayQueryFragment, _RelayQueryNode3);


  /**
   * Helper to construct a new fragment with the given attributes and 'empty'
   * route/variables.
   */

  RelayQueryFragment.build = function build(name, type,
  /* $FlowIssue: #11220887
     `Array<Subclass-of-RelayQueryNode>` should be compatible here. */
  children, metadata, routeName) {
    var nextChildren = children ? children.filter(function (child) {
      return !!child;
    }) : [];
    var concreteFragment = require('./QueryBuilder').createFragment({
      name: name,
      type: type,
      metadata: metadata
    });
    var fragment = new RelayQueryFragment(concreteFragment, require('./RelayMetaRoute').get(routeName || '$RelayQuery'), {}, {
      isDeferred: !!(metadata && metadata.isDeferred),
      isContainerFragment: !!(metadata && metadata.isContainerFragment),
      isTypeConditional: !!(metadata && metadata.isTypeConditional)
    });
    fragment.__children__ = nextChildren;
    return fragment;
  };

  RelayQueryFragment.create = function create(concreteNode, route, variables, metadata) {
    var fragment = require('./QueryBuilder').getFragment(concreteNode);
    require('fbjs/lib/invariant')(fragment, 'RelayQueryFragment.create(): ' + 'Expected a GraphQL `fragment { ... }`, got: %s', concreteNode);
    return createMemoizedFragment(fragment, route, variables, metadata || DEFAULT_FRAGMENT_METADATA);
  };

  function RelayQueryFragment(concreteNode, route, variables, metadata) {
    (0, _classCallCheck3['default'])(this, RelayQueryFragment);

    var _this9 = (0, _possibleConstructorReturn3['default'])(this, _RelayQueryNode3.call(this, concreteNode, route, variables));

    _this9.__compositeHash__ = null;
    _this9.__metadata__ = metadata || DEFAULT_FRAGMENT_METADATA;
    _this9.__sourceCompositeHash__ = null;
    return _this9;
  }

  RelayQueryFragment.prototype.canHaveSubselections = function canHaveSubselections() {
    return true;
  };

  RelayQueryFragment.prototype.getDebugName = function getDebugName() {
    return this.__concreteNode__.name;
  };

  /**
   * The "concrete fragment id" uniquely identifies a Relay.QL`fragment ...`
   * within the source code of an application and will remain the same across
   * runs of a particular version of an application.
   */


  RelayQueryFragment.prototype.getConcreteFragmentID = function getConcreteFragmentID() {
    return this.__concreteNode__.id;
  };

  /**
   * The "composite hash" is similar to the concrete instance hash, but it also
   * differentiates between varying variable values or route names.
   *
   * The composite hash is used to:
   * - Avoid printing the same fragment twice, in order to reduce upload size.
   * - Remember which deferred fragment/data pairs have been fetched.
   */


  RelayQueryFragment.prototype.getCompositeHash = function getCompositeHash() {
    var compositeHash = this.__compositeHash__;
    if (!compositeHash) {
      // TODO: Simplify this hash function, #9599170.
      compositeHash = require('./generateRQLFieldAlias')(this.getConcreteFragmentID() + '.' + this.__route__.name + '.' + require('./stableStringify')(this.__variables__));
      this.__compositeHash__ = compositeHash;
    }
    return compositeHash;
  };

  RelayQueryFragment.prototype.getSourceCompositeHash = function getSourceCompositeHash() {
    return this.__sourceCompositeHash__;
  };

  RelayQueryFragment.prototype.isAbstract = function isAbstract() {
    return !!this.__concreteNode__.metadata.isAbstract;
  };

  RelayQueryFragment.prototype.isDeferred = function isDeferred() {
    return this.__metadata__.isDeferred;
  };

  RelayQueryFragment.prototype.isPattern = function isPattern() {
    return !!this.__concreteNode__.metadata.pattern;
  };

  RelayQueryFragment.prototype.isPlural = function isPlural() {
    var metadata = this.__concreteNode__.metadata;
    return !!( // FB Printer
    metadata.isPlural || metadata.plural);
  };

  // OSS Printer from `@relay`

  RelayQueryFragment.prototype.isTrackingEnabled = function isTrackingEnabled() {
    var metadata = this.__concreteNode__.metadata;
    return !!metadata.isTrackingEnabled;
  };

  RelayQueryFragment.prototype.cloneAsPlainFragment = function cloneAsPlainFragment() {
    return createMemoizedFragment(this.__concreteNode__, this.__route__, this.__variables__, DEFAULT_FRAGMENT_METADATA);
  };

  RelayQueryFragment.prototype.isContainerFragment = function isContainerFragment() {
    return this.__metadata__.isContainerFragment;
  };

  RelayQueryFragment.prototype.isTypeConditional = function isTypeConditional() {
    return this.__metadata__.isTypeConditional;
  };

  RelayQueryFragment.prototype.hasDeferredDescendant = function hasDeferredDescendant() {
    return this.isDeferred() || _RelayQueryNode3.prototype.hasDeferredDescendant.call(this);
  };

  RelayQueryFragment.prototype.clone = function clone(children) {
    var clone = _RelayQueryNode3.prototype.clone.call(this, children);
    if (clone !== this && clone instanceof RelayQueryFragment) {
      clone.__concreteNode__ = (0, _extends3['default'])({}, clone.__concreteNode__, {
        id: require('./generateConcreteFragmentID')()
      });
      clone.__metadata__ = (0, _extends3['default'])({}, this.__metadata__);

      // The container checks on the status of a deferred fragment using its
      // composite hash. We need to cache this hash in this cloned fragment
      // so it can be updated in the store with the correct hash when fetched.
      clone.__sourceCompositeHash__ = this.getCompositeHash();
    }
    return clone;
  };

  RelayQueryFragment.prototype.equals = function equals(that) {
    if (this === that) {
      return true;
    }
    if (!(that instanceof RelayQueryFragment)) {
      return false;
    }
    if (this.getType() !== that.getType()) {
      return false;
    }
    return _RelayQueryNode3.prototype.equals.call(this, that);
  };

  return RelayQueryFragment;
}(RelayQueryNode);

/**
 * @internal
 *
 * Wraps access to query fields.
 */


var RelayQueryField = function (_RelayQueryNode4) {
  (0, _inherits3['default'])(RelayQueryField, _RelayQueryNode4);

  RelayQueryField.create = function create(concreteNode, route, variables) {
    var field = require('./QueryBuilder').getField(concreteNode);
    require('fbjs/lib/invariant')(field, 'RelayQueryField.create(): Expected a GraphQL field, got: %s', concreteNode);
    return new RelayQueryField(field, route, variables);
  };

  /**
   * Helper to construct a new field with the given attributes and 'empty'
   * route/variables.
   */


  RelayQueryField.build = function build(_ref) {
    var alias = _ref.alias;
    var directives = _ref.directives;
    var calls = _ref.calls;
    var children = _ref.children;
    var fieldName = _ref.fieldName;
    var metadata = _ref.metadata;
    var routeName = _ref.routeName;
    var type = _ref.type;

    var nextChildren = children ? children.filter(function (child) {
      return !!child;
    }) : [];
    var concreteField = require('./QueryBuilder').createField({
      alias: alias,
      calls: calls ? require('./callsToGraphQL')(calls) : null,
      directives: directives ? require('./directivesToGraphQL')(directives) : null,
      fieldName: fieldName,
      metadata: metadata,
      type: type
    });
    var field = new RelayQueryField(concreteField, require('./RelayMetaRoute').get(routeName || '$RelayQuery'), {});
    field.__children__ = nextChildren;
    return field;
  };

  function RelayQueryField(concreteNode, route, variables) {
    (0, _classCallCheck3['default'])(this, RelayQueryField);

    var _this10 = (0, _possibleConstructorReturn3['default'])(this, _RelayQueryNode4.call(this, concreteNode, route, variables));

    _this10.__debugName__ = undefined;
    _this10.__isRefQueryDependency__ = false;
    _this10.__rangeBehaviorCalls__ = undefined;
    _this10.__shallowHash__ = undefined;
    return _this10;
  }

  RelayQueryField.prototype.canHaveSubselections = function canHaveSubselections() {
    return !!this.__concreteNode__.metadata.canHaveSubselections;
  };

  RelayQueryField.prototype.isAbstract = function isAbstract() {
    return !!this.__concreteNode__.metadata.isAbstract;
  };

  RelayQueryField.prototype.isFindable = function isFindable() {
    return !!this.__concreteNode__.metadata.isFindable;
  };

  RelayQueryField.prototype.isGenerated = function isGenerated() {
    return !!this.__concreteNode__.metadata.isGenerated;
  };

  RelayQueryField.prototype.isConnection = function isConnection() {
    return !!this.__concreteNode__.metadata.isConnection;
  };

  RelayQueryField.prototype.isConnectionWithoutNodeID = function isConnectionWithoutNodeID() {
    return !!this.__concreteNode__.metadata.isConnectionWithoutNodeID;
  };

  RelayQueryField.prototype.isPlural = function isPlural() {
    return !!this.__concreteNode__.metadata.isPlural;
  };

  RelayQueryField.prototype.isRefQueryDependency = function isRefQueryDependency() {
    return this.__isRefQueryDependency__;
  };

  RelayQueryField.prototype.isRequisite = function isRequisite() {
    return !!this.__concreteNode__.metadata.isRequisite;
  };

  RelayQueryField.prototype.getDebugName = function getDebugName() {
    var _this11 = this;

    var debugName = this.__debugName__;
    if (!debugName) {
      (function () {
        debugName = _this11.getSchemaName();
        var printedCoreArgs = void 0;
        _this11.getCallsWithValues().forEach(function (arg) {
          if (_this11._isCoreArg(arg)) {
            printedCoreArgs = printedCoreArgs || [];
            printedCoreArgs.push(require('./serializeRelayQueryCall')(arg));
          }
        });
        if (printedCoreArgs) {
          debugName += printedCoreArgs.sort().join('');
        }
        _this11.__debugName__ = debugName;
      })();
    }
    return debugName;
  };

  /**
   * The canonical name for the referenced field in the schema.
   */


  RelayQueryField.prototype.getSchemaName = function getSchemaName() {
    return this.__concreteNode__.fieldName;
  };

  /**
  * An Array of Calls to be used with rangeBehavior config functions.
  *
  * Non-core arguments (like connection and identifying arguments) are dropped.
  *   `field(first: 10, foo: "bar", baz: "bat")` => `'baz(bat).foo(bar)'`
  *   `username(name: "steve")`                  => `''`
  */


  RelayQueryField.prototype.getRangeBehaviorCalls = function getRangeBehaviorCalls() {
    var _this12 = this;

    require('fbjs/lib/invariant')(this.isConnection(), 'RelayQueryField: Range behavior keys are associated exclusively with ' + 'connection fields. `getRangeBehaviorCalls()` was called on the ' + 'non-connection field `%s`.', this.getSchemaName());

    var rangeBehaviorCalls = this.__rangeBehaviorCalls__;
    if (!rangeBehaviorCalls) {
      rangeBehaviorCalls = this.getCallsWithValues().filter(function (arg) {
        return _this12._isCoreArg(arg);
      });
      this.__rangeBehaviorCalls__ = rangeBehaviorCalls;
    }
    return rangeBehaviorCalls;
  };

  /**
   * The name for the field when serializing the query or interpreting query
   * responses from the server. The serialization key is derived from
   * all calls/values and hashed for compactness.
   *
   * Given the GraphQL
   *   `field(first: 10, foo: "bar", baz: "bat")`, or
   *   `field(baz: "bat", foo: "bar", first: 10)`
   *
   * ...the following serialization key will be produced:
   *   `generateRQLFieldAlias('field.bar(bat).first(10).foo(bar)')`
   */


  RelayQueryField.prototype.getSerializationKey = function getSerializationKey() {
    var serializationKey = this.__serializationKey__;
    if (!serializationKey) {
      var _key = this.getSchemaName();
      var calls = this.getCallsWithValues();
      if (calls.length) {
        var alias = this.__concreteNode__.alias;
        if (alias != null) {
          _key += '.' + alias;
        }
        _key += calls.map(require('./serializeRelayQueryCall')).sort().join('');
      }
      serializationKey = require('./generateRQLFieldAlias')(_key);
      this.__serializationKey__ = serializationKey;
    }
    return serializationKey;
  };

  /**
   * Returns a hash of the field name and all argument values.
   */


  RelayQueryField.prototype.getShallowHash = function getShallowHash() {
    var shallowHash = this.__shallowHash__;
    if (!shallowHash) {
      this.__shallowHash__ = shallowHash = this.getSchemaName() + serializeCalls(this.getCallsWithValues());
    }
    return shallowHash;
  };

  /**
   * The name which Relay internals can use to reference this field, without
   * collisions.
   *
   * Given the GraphQL
   *   `field(first: 10, foo: "bar", baz: "bat")`, or
   *   `field(baz: "bat", foo: "bar", first: 10)`
   *
   * ...the following storage key will be produced:
   *   `'field{bar:"bat",foo:"bar"}'`
   */


  RelayQueryField.prototype.getStorageKey = function getStorageKey() {
    var _this13 = this;

    var storageKey = this.__storageKey__;
    if (!storageKey) {
      this.__storageKey__ = storageKey = this.getSchemaName() + serializeCalls(this.getCallsWithValues().filter(function (call) {
        return _this13._isCoreArg(call);
      }));
    }
    return storageKey;
  };

  /**
   * The name by which this field's results should be made available to the
   * application.
   */


  RelayQueryField.prototype.getApplicationName = function getApplicationName() {
    var concreteNode = this.__concreteNode__;
    return concreteNode.alias || concreteNode.fieldName;
  };

  RelayQueryField.prototype.getInferredRootCallName = function getInferredRootCallName() {
    return this.__concreteNode__.metadata.inferredRootCallName;
  };

  RelayQueryField.prototype.getInferredPrimaryKey = function getInferredPrimaryKey() {
    return this.__concreteNode__.metadata.inferredPrimaryKey;
  };

  RelayQueryField.prototype.getCallsWithValues = function getCallsWithValues() {
    var calls = this.__calls__;
    if (!calls) {
      var concreteCalls = this.__concreteNode__.calls;
      if (concreteCalls) {
        calls = require('./callsFromGraphQL')(concreteCalls, this.__variables__);
      } else {
        calls = EMPTY_CALLS;
      }
      this.__calls__ = calls;
    }
    return calls;
  };

  RelayQueryField.prototype.getCallType = function getCallType(callName) {
    var concreteCalls = this.__concreteNode__.calls;
    var concreteCall = concreteCalls && concreteCalls.filter(function (call) {
      return call.name === callName;
    })[0];
    if (concreteCall) {
      return concreteCall.metadata.type;
    }
  };

  RelayQueryField.prototype.equals = function equals(that) {
    if (this === that) {
      return true;
    }
    if (!(that instanceof RelayQueryField)) {
      return false;
    }
    if (this.getSchemaName() !== that.getSchemaName() || this.getApplicationName() !== that.getApplicationName() || !areCallValuesEqual(this.getCallsWithValues(), that.getCallsWithValues())) {
      return false;
    }
    return _RelayQueryNode4.prototype.equals.call(this, that);
  };

  RelayQueryField.prototype.cloneAsRefQueryDependency = function cloneAsRefQueryDependency() {
    var field = new RelayQueryField(this.__concreteNode__, this.__route__, this.__variables__);
    field.__children__ = [];
    field.__isRefQueryDependency__ = true;
    return field;
  };

  RelayQueryField.prototype.cloneFieldWithCalls = function cloneFieldWithCalls(children, calls) {
    if (!this.canHaveSubselections()) {
      // Compact new children *after* this check, for consistency.
      require('fbjs/lib/invariant')(children.length === 0, 'RelayQueryNode: Cannot add children to field `%s` because it does ' + 'not support sub-selections (sub-fields).', this.getSchemaName());
    }

    // use `clone()` if call values do not change
    if (require('fbjs/lib/areEqual')(this.getCallsWithValues(), calls)) {
      var clone = this.clone(children);
      return clone;
    }

    var nextChildren = cloneChildren(this.getChildren(), children);
    if (!nextChildren.length) {
      return null;
    }

    var field = new RelayQueryField(this.__concreteNode__, this.__route__, this.__variables__);
    field.__children__ = nextChildren;
    field.__calls__ = calls;

    return field;
  };

  /**
   * The following types of arguments are non-core:
   * - Range calls such as `first` or `find` on connections.
   * - Conditionals when the field is present.
   */


  RelayQueryField.prototype._isCoreArg = function _isCoreArg(arg) {
    return(
      // `name(if:true)`, `name(unless:false)`, and `name` are equivalent.
      !(arg.name === IF && String(arg.value) === TRUE) && !(arg.name === UNLESS && String(arg.value) === FALSE) &&
      // Connection arguments can be stripped out.
      !(this.isConnection() && require('./RelayConnectionInterface').isConnectionCall(arg))
    );
  };

  return RelayQueryField;
}(RelayQueryNode);

function createNode(concreteNode, route, variables) {
  require('fbjs/lib/invariant')(typeof concreteNode === 'object' && concreteNode !== null, 'RelayQueryNode: Expected a GraphQL object created with `Relay.QL`, got' + '`%s`.', concreteNode);
  var kind = concreteNode.kind;
  var type = RelayQueryNode;
  if (kind === 'Field') {
    type = RelayQueryField;
  } else if (kind === 'Fragment') {
    type = RelayQueryFragment;
  } else if (kind === 'FragmentReference') {
    type = RelayQueryFragment;
    var fragment = require('./QueryBuilder').getFragment(concreteNode.fragment);
    // TODO #9171213: Reference directives should override fragment directives
    if (fragment) {
      return createMemoizedFragment(fragment, route, {}, {
        isDeferred: false,
        isContainerFragment: true,
        isTypeConditional: false
      });
    }
  } else if (kind === 'Query') {
    type = RelayQueryRoot;
  } else if (kind === 'Mutation') {
    type = RelayQueryMutation;
  } else if (kind === 'Subscription') {
    type = RelayQuerySubscription;
  } else if (concreteNode instanceof require('./RelayRouteFragment')) {
    var _fragment = concreteNode.getFragmentForRoute(route);
    // May be null if no value was defined for this route.
    if (Array.isArray(_fragment)) {
      // A route-conditional function may return a single fragment reference
      // or an array of fragment references.
      return _fragment.map(function (frag) {
        return createNode(frag, route, variables);
      });
    } else if (_fragment) {
      return createNode(_fragment, route, variables);
    }
    return null;
  } else if (concreteNode instanceof require('./RelayFragmentReference')) {
    var _fragment2 = concreteNode.getFragment(variables);
    var fragmentVariables = concreteNode.getVariables(route, variables);
    if (_fragment2) {
      // the fragment may be null when `if` or `unless` conditions are not met.
      return createMemoizedFragment(_fragment2, route, fragmentVariables, {
        isDeferred: concreteNode.isDeferred(),
        isContainerFragment: concreteNode.isContainerFragment(),
        isTypeConditional: concreteNode.isTypeConditional()
      });
    }
    return null;
  } else {}
  return new type(concreteNode, route, variables);
}

/**
 * Memoizes the `RelayQueryFragment` equivalent of a given GraphQL fragment
 * for the given route, variables, and deferred status.
 */
function createMemoizedFragment(concreteFragment, route, variables, metadata) {
  var cacheKey = route.name + ':' + require('./stableStringify')(variables) + ':' + require('./stableStringify')(metadata);
  var fragment = concreteFragment.__cachedFragment__;
  var fragmentCacheKey = concreteFragment.__cacheKey__;
  if (!fragment || fragmentCacheKey !== cacheKey) {
    fragment = new RelayQueryFragment(concreteFragment, route, variables, metadata);
    concreteFragment.__cachedFragment__ = fragment;
    concreteFragment.__cacheKey__ = cacheKey;
  }
  return fragment;
}

/**
 * Compacts new children and compares them to the previous children.
 * - If all items are equal, returns previous array
 * - If any items differ, returns new array
 */
function cloneChildren(prevChildren, nextChildren) {
  var children = [];
  var isSameChildren = true;

  var prevIndex = 0;
  for (var ii = 0; ii < nextChildren.length; ii++) {
    var child = nextChildren[ii];
    if (child) {
      children.push(child);
      isSameChildren = isSameChildren && child === prevChildren[prevIndex++];
    }
  }

  if (isSameChildren && children.length === prevChildren.length) {
    return prevChildren;
  } else {
    return children;
  }
}

/**
 * Creates an opaque serialization of calls.
 */
function serializeCalls(calls) {
  if (calls.length) {
    var _ret5 = function () {
      var callMap = {};
      calls.forEach(function (call) {
        callMap[call.name] = call.value;
      });
      return {
        v: require('./stableStringify')(callMap)
      };
    }();

    if (typeof _ret5 === "object") return _ret5.v;
  } else {
    return '';
  }
}

/**
 * Checks if two sets of calls have equal names and values. This skips testing
 * argument types because type metadata for scalar arguments may be omitted by
 * the Babel plugin.
 */
function areCallValuesEqual(thisCalls, thatCalls) {
  if (thisCalls.length !== thatCalls.length) {
    return false;
  }
  return thisCalls.every(function (_ref2, ii) {
    var thisName = _ref2.name;
    var thisValue = _ref2.value;
    var _thatCalls$ii = thatCalls[ii];
    var thatName = _thatCalls$ii.name;
    var thatValue = _thatCalls$ii.value;

    if (thisName !== thatName) {
      return false;
    }
    if (thisValue instanceof require('./RelayVariable')) {
      return thisValue.equals(thatValue);
    }
    return require('fbjs/lib/areEqual')(thisValue, thatValue);
  });
}

require('./RelayProfiler').instrumentMethods(RelayQueryNode.prototype, {
  clone: '@RelayQueryNode.prototype.clone',
  equals: '@RelayQueryNode.prototype.equals',
  getChildren: '@RelayQueryNode.prototype.getChildren',
  getDirectives: '@RelayQueryNode.prototype.getDirectives',
  hasDeferredDescendant: '@RelayQueryNode.prototype.hasDeferredDescendant',
  getFieldByStorageKey: '@RelayQueryNode.prototype.getFieldByStorageKey'
});

require('./RelayProfiler').instrumentMethods(RelayQueryField.prototype, {
  getStorageKey: '@RelayQueryField.prototype.getStorageKey',
  getSerializationKey: '@RelayQueryField.prototype.getSerializationKey'
});

module.exports = {
  Field: RelayQueryField,
  Fragment: RelayQueryFragment,
  Mutation: RelayQueryMutation,
  Node: RelayQueryNode,
  Operation: RelayQueryOperation,
  Root: RelayQueryRoot,
  Subscription: RelayQuerySubscription
};