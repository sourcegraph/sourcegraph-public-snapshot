/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule transformRelayQueryPayload
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * Transforms "client" payloads with property keys that match the "application"
 * names (i.e. property names are schema names or aliases) into "server"
 * payloads that match what the server would return for the given query (i.e.
 * property names are serialization keys instead).
 */
function transformRelayQueryPayload(root, clientData, config) {
  if (clientData == null) {
    return clientData;
  } else {
    return require('fbjs/lib/mapObject')(clientData, function (item) {
      // Handle both FB & OSS formats for root payloads on plural calls: FB
      // returns objects, OSS returns arrays.
      if (Array.isArray(item)) {
        return item.map(function (innerItem) {
          return transform(root, innerItem, config);
        });
      }
      return transform(root, item, config);
    });
  }
}

function transform(root, clientData, config) {
  if (clientData == null) {
    return clientData;
  }
  var transformer = new RelayPayloadTransformer(config);
  var serverData = {};
  transformer.visit(root, {
    client: clientData,
    server: serverData
  });
  return serverData;
}

var RelayPayloadTransformer = function (_RelayQueryVisitor) {
  (0, _inherits3['default'])(RelayPayloadTransformer, _RelayQueryVisitor);

  function RelayPayloadTransformer(config) {
    (0, _classCallCheck3['default'])(this, RelayPayloadTransformer);

    var _this = (0, _possibleConstructorReturn3['default'])(this, _RelayQueryVisitor.call(this));

    if (config) {
      _this._getKeyForClientData = config.getKeyForClientData;
      _this._traverseChildren = config.traverseChildren;
    }
    return _this;
  }

  RelayPayloadTransformer.prototype._getKeyForClientData = function _getKeyForClientData(field) {
    return field.getApplicationName();
  };

  RelayPayloadTransformer.prototype.traverseChildren = function traverseChildren(node, nextState, callback, context) {
    if (this._traverseChildren) {
      this._traverseChildren(node, callback, context);
    } else {
      _RelayQueryVisitor.prototype.traverseChildren.call(this, node, nextState, callback, context);
    }
  };

  RelayPayloadTransformer.prototype.visitField = function visitField(node, state) {
    var _this2 = this;

    var client = state.client;
    var server = state.server;

    var applicationName = this._getKeyForClientData(node);
    var serializationKey = node.getSerializationKey();
    var clientData = client[applicationName];
    var serverData = server[serializationKey];

    if (!client.hasOwnProperty(applicationName)) {
      return;
    } else if (!node.canHaveSubselections() || clientData == null) {
      server[serializationKey] = clientData;
    } else if (Array.isArray(clientData)) {
      if (serverData == null) {
        server[serializationKey] = serverData = [];
      }
      // $FlowFixMe(>=0.31.0)
      clientData.forEach(function (clientItem, index) {
        require('fbjs/lib/invariant')(Array.isArray(serverData), 'RelayPayloadTransformer: Got conflicting values for field `%s`: ' + 'expected values to be arrays.', applicationName);
        if (clientItem == null) {
          serverData[index] = clientItem;
          return;
        }
        var serverItem = serverData && serverData[index];
        if (serverItem == null) {
          serverData[index] = serverItem = {};
        }
        // $FlowFixMe(>=0.31.0)
        _this2.traverse(node, {
          client: clientItem,
          server: serverItem
        });
      });
    } else {
      require('fbjs/lib/invariant')(typeof clientData === 'object' && clientData !== null, 'RelayPayloadTransformer: Expected an object value for field `%s`.', applicationName);
      require('fbjs/lib/invariant')(serverData == null || typeof serverData === 'object', 'RelayPayloadTransformer: Got conflicting values for field `%s`: ' + 'expected values to be objects.', applicationName);
      if (serverData == null) {
        server[serializationKey] = serverData = {};
      }
      this.traverse(node, {
        client: clientData,
        server: serverData
      });
    }
  };

  return RelayPayloadTransformer;
}(require('./RelayQueryVisitor'));

module.exports = transformRelayQueryPayload;