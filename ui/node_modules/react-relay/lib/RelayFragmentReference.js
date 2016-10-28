/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayFragmentReference
 * 
 */

'use strict';

var _extends3 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @internal
 *
 * RelayFragmentReference is the return type of fragment composition:
 *
 *   fragment on Foo {
 *     ${Child.getFragment('bar', {baz: variables.qux})}
 *   }
 *
 * Whereas a fragment defines a sub-query's structure, a fragment reference is
 * a particular instantiation of the fragment as it is composed within a query
 * or another fragment. It encodes the source fragment, initial variables, and
 * a mapping from variables in the composing query's (or fragment's) scope to
 * variables in the fragment's scope.
 *
 * The variable mapping is represented by `variableMapping`, a dictionary that
 * maps from names of variables in the parent scope to variables that exist in
 * the fragment. Example:
 *
 * ```
 * // Fragment:
 * var Container = Relay.createContainer(..., {
 *   initialVariables: {
 *     private: 'foo',
 *     public: 'bar',
 *     variable: null,
 *   },
 *   fragments: {
 *     foo: ...
 *   }
 * });
 *
 * // Reference:
 * ${Container.getQuery(
 *   'foo',
 *   // Variable Mapping:
 *   {
 *     public: 'BAR',
 *     variable: variables.source,
 *   }
 * )}
 * ```
 *
 * When evaluating the referenced fragment, `$public` will be overridden with
 * `'Bar'`. The value of `$variable` will become the value of `$source` in the
 * outer scope. This is analagous to:
 *
 * ```
 * function inner(private = 'foo', public = 'bar', variable) {}
 * function outer(source) {
 *   inner(public = 'BAR', variable = source);
 * }
 * ```
 *
 * Where the value of the inner `variable` depends on how `outer` is called.
 *
 * The `prepareVariables` function allows for variables to be modified based on
 * the runtime environment or route name.
 */

var RelayFragmentReference = function () {
  RelayFragmentReference.createForContainer = function createForContainer(fragmentGetter, initialVariables, variableMapping, prepareVariables) {
    var reference = new RelayFragmentReference(fragmentGetter, initialVariables, variableMapping, prepareVariables);
    reference._isContainerFragment = true;
    return reference;
  };

  function RelayFragmentReference(fragmentGetter, initialVariables, variableMapping, prepareVariables) {
    (0, _classCallCheck3['default'])(this, RelayFragmentReference);

    this._conditions = null;
    this._initialVariables = initialVariables || {};
    this._fragment = undefined;
    this._fragmentGetter = fragmentGetter;
    this._isContainerFragment = false;
    this._isDeferred = false;
    this._isTypeConditional = false;
    this._variableMapping = variableMapping;
    this._prepareVariables = prepareVariables;
  }

  RelayFragmentReference.prototype.conditionOnType = function conditionOnType() {
    this._isTypeConditional = true;
    return this;
  };

  RelayFragmentReference.prototype.getConditions = function getConditions() {
    return this._conditions;
  };

  RelayFragmentReference.prototype.getFragmentUnconditional = function getFragmentUnconditional() {
    var fragment = this._fragment;
    if (fragment == null) {
      fragment = this._fragmentGetter();
      this._fragment = fragment;
    }
    return fragment;
  };

  RelayFragmentReference.prototype.getInitialVariables = function getInitialVariables() {
    return this._initialVariables;
  };

  RelayFragmentReference.prototype.getVariableMapping = function getVariableMapping() {
    return this._variableMapping;
  };

  /**
   * Mark this usage of the fragment as deferred.
   */


  RelayFragmentReference.prototype.defer = function defer() {
    this._isDeferred = true;
    return this;
  };

  /**
   * Mark this fragment for inclusion only if the given variable is truthy.
   */


  RelayFragmentReference.prototype['if'] = function _if(value) {
    var callVariable = require('./QueryBuilder').getCallVariable(value);
    require('fbjs/lib/invariant')(callVariable, 'RelayFragmentReference: Invalid value `%s` supplied to `if()`. ' + 'Expected a variable.', callVariable);
    this._addCondition({
      passingValue: true,
      variable: callVariable.callVariableName
    });
    return this;
  };

  /**
   * Mark this fragment for inclusion only if the given variable is falsy.
   */


  RelayFragmentReference.prototype.unless = function unless(value) {
    var callVariable = require('./QueryBuilder').getCallVariable(value);
    require('fbjs/lib/invariant')(callVariable, 'RelayFragmentReference: Invalid value `%s` supplied to `unless()`. ' + 'Expected a variable.', callVariable);
    this._addCondition({
      passingValue: false,
      variable: callVariable.callVariableName
    });
    return this;
  };

  /**
   * Get the referenced fragment if all conditions are met.
   */


  RelayFragmentReference.prototype.getFragment = function getFragment(variables) {
    // determine if the variables match the supplied if/unless conditions
    var conditions = this._conditions;
    if (conditions && !conditions.every(function (_ref) {
      var variable = _ref.variable;
      var passingValue = _ref.passingValue;

      return !!variables[variable] === passingValue;
    })) {
      return null;
    }
    return this.getFragmentUnconditional();
  };

  /**
   * Get the variables to pass to the referenced fragment, accounting for
   * initial values, overrides, and route-specific variables.
   */


  RelayFragmentReference.prototype.getVariables = function getVariables(route, variables) {
    var _this = this;

    var innerVariables = (0, _extends3['default'])({}, this._initialVariables);

    // map variables from outer -> inner scope
    var variableMapping = this._variableMapping;
    if (variableMapping) {
      require('fbjs/lib/forEachObject')(variableMapping, function (value, name) {
        var callVariable = require('./QueryBuilder').getCallVariable(value);
        if (callVariable) {
          value = variables[callVariable.callVariableName];
        }
        if (value === undefined) {
          require('fbjs/lib/warning')(false, 'RelayFragmentReference: Variable `%s` is undefined in fragment ' + '`%s`.', name, _this.getFragmentUnconditional().name);
        } else {
          innerVariables[name] = value;
        }
      });
    }

    var prepareVariables = this._prepareVariables;
    if (prepareVariables) {
      innerVariables = prepareVariables(innerVariables, route);
    }

    return innerVariables;
  };

  RelayFragmentReference.prototype.isContainerFragment = function isContainerFragment() {
    return this._isContainerFragment;
  };

  RelayFragmentReference.prototype.isDeferred = function isDeferred() {
    return this._isDeferred;
  };

  RelayFragmentReference.prototype.isTypeConditional = function isTypeConditional() {
    return this._isTypeConditional;
  };

  RelayFragmentReference.prototype._addCondition = function _addCondition(condition) {
    var conditions = this._conditions;
    if (!conditions) {
      conditions = [];
      this._conditions = conditions;
    }
    conditions.push(condition);
  };

  return RelayFragmentReference;
}();

module.exports = RelayFragmentReference;