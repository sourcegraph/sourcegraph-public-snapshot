/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule validateRelayReadQuery
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var validateRelayReadQuery = require('fbjs/lib/emptyFunction');

if (process.env.NODE_ENV !== 'production') {
  // Wrap in an IIFE to avoid unwanted function hoisting.
  (function () {
    /**
     * @internal
     *
     * `validateRelayReadQuery` is a `__DEV__`-only validator that checks that a
     * query used to read data from `RelayStore` is well-formed. Validation
     * problems are reported via `console.error`.
     *
     * At the moment, "well-formed" means that the query does not contain
     * duplicate aliases.
     */
    validateRelayReadQuery = function _validateRelayReadQuery(queryNode, options) {
      var validator = new RelayStoreReadValidator(options);
      validator.visit(queryNode, {
        children: {},
        hash: null
      });
    };

    /**
     * Returns the nested AliasMap for `node`, initializing if it necessary.
     */
    function getAliasMap(node, parentAliasMap) {
      var applicationName = node.getApplicationName();
      var hash = node.getShallowHash();
      var children = parentAliasMap.children;

      if (!children.hasOwnProperty(applicationName)) {
        children[applicationName] = {
          children: {},
          hash: hash
        };
      } else if (children[applicationName].hash !== hash) {
        console.error('`%s` is used as an alias more than once. Please use unique aliases.', applicationName);
      }
      return children[applicationName];
    }

    var RelayStoreReadValidator = function (_RelayQueryVisitor) {
      (0, _inherits3['default'])(RelayStoreReadValidator, _RelayQueryVisitor);

      function RelayStoreReadValidator(options) {
        (0, _classCallCheck3['default'])(this, RelayStoreReadValidator);

        var _this = (0, _possibleConstructorReturn3['default'])(this, _RelayQueryVisitor.call(this));

        _this._traverseFragmentReferences = options && options.traverseFragmentReferences || false;
        return _this;
      }

      RelayStoreReadValidator.prototype.visitField = function visitField(node, parentAliasMap) {
        var aliasMap = getAliasMap(node, parentAliasMap);

        if (node.isGenerated()) {
          return;
        } else if (!node.canHaveSubselections()) {
          return;
        } else if (node.isPlural()) {
          this._readPlural(node, aliasMap);
        } else {
          // No special handling needed for connections, edges, page_info etc.
          this._readLinkedField(node, aliasMap);
        }
      };

      RelayStoreReadValidator.prototype.visitFragment = function visitFragment(node, aliasMap) {
        if (this._traverseFragmentReferences || !node.isContainerFragment()) {
          this.traverse(node, aliasMap);
        }
      };

      RelayStoreReadValidator.prototype._readPlural = function _readPlural(node, aliasMap) {
        var _this2 = this;

        node.getChildren().forEach(function (child) {
          return _this2.visit(child, aliasMap);
        });
      };

      RelayStoreReadValidator.prototype._readLinkedField = function _readLinkedField(node, aliasMap) {
        aliasMap = getAliasMap(node, aliasMap);
        this.traverse(node, aliasMap);
      };

      return RelayStoreReadValidator;
    }(require('./RelayQueryVisitor'));
  })();
}

module.exports = validateRelayReadQuery;