/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayRoute
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var createURI = function createURI() {
  return null;
};

/**
 * Describes the root queries, param definitions and other metadata for a given
 * path (URI).
 */

var RelayRoute = function (_RelayQueryConfig) {
  (0, _inherits3['default'])(RelayRoute, _RelayQueryConfig);

  function RelayRoute(initialVariables, uri) {
    (0, _classCallCheck3['default'])(this, RelayRoute);

    var _this = (0, _possibleConstructorReturn3['default'])(this, _RelayQueryConfig.call(this, initialVariables));

    var constructor = _this.constructor;
    var routeName = constructor.routeName;
    var path = constructor.path;


    require('fbjs/lib/invariant')(constructor !== RelayRoute, 'RelayRoute: Abstract class cannot be instantiated.');
    require('fbjs/lib/invariant')(routeName, '%s: Subclasses of RelayRoute must define a `routeName`.', constructor.name || '<<anonymous>>');

    // $FlowIssue #9905535 - Object.defineProperty doesn't understand getters
    Object.defineProperty(_this, 'uri', {
      enumerable: true,
      get: function get() {
        if (!uri && path) {
          uri = createURI(constructor, this.params);
        }
        return uri;
      }
    });
    return _this;
  }

  RelayRoute.prototype.prepareVariables = function prepareVariables(prevVariables) {
    var _constructor = this.constructor;
    var paramDefinitions = _constructor.paramDefinitions;
    var prepareParams = _constructor.prepareParams;
    var routeName = _constructor.routeName;

    var params = prevVariables;
    if (prepareParams) {
      /* $FlowFixMe(>=0.17.0) - params is ?Tv but prepareParams expects Tv */
      params = prepareParams(params);
    }
    require('fbjs/lib/forEachObject')(paramDefinitions, function (paramDefinition, paramName) {
      if (params) {
        if (params.hasOwnProperty(paramName)) {
          return;
        } else {
          // Backfill param so that a call variable is created for it.
          params[paramName] = undefined;
        }
      }
      require('fbjs/lib/invariant')(!paramDefinition.required, 'RelayRoute: Missing required parameter `%s` in `%s`. Check the ' + 'supplied params or URI.', paramName, routeName);
    });
    return params;
  };

  RelayRoute.injectURICreator = function injectURICreator(creator) {
    createURI = creator;
  };

  return RelayRoute;
}(require('./RelayQueryConfig'));

module.exports = RelayRoute;