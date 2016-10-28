/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayDefaultNetworkLayer
 * 
 */

'use strict';

var _extends3 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _stringify2 = _interopRequireDefault(require('babel-runtime/core-js/json/stringify'));

var _promise2 = _interopRequireDefault(require('fbjs/lib/Promise'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var RelayDefaultNetworkLayer = function () {
  // InitWithRetries

  function RelayDefaultNetworkLayer(uri, init) {
    (0, _classCallCheck3['default'])(this, RelayDefaultNetworkLayer);

    this._uri = uri;
    this._init = (0, _extends3['default'])({}, init);

    // Facilitate reuse when creating custom network layers.
    this.sendMutation = this.sendMutation.bind(this);
    this.sendQueries = this.sendQueries.bind(this);
    this.supports = this.supports.bind(this);
  }

  RelayDefaultNetworkLayer.prototype.sendMutation = function sendMutation(request) {
    return this._sendMutation(request).then(function (result) {
      return result.json();
    }).then(function (payload) {
      if (payload.hasOwnProperty('errors')) {
        var error = createRequestError(request, '200', payload);
        request.reject(error);
      } else {
        request.resolve({ response: payload.data });
      }
    })['catch'](function (error) {
      return request.reject(error);
    });
  };

  RelayDefaultNetworkLayer.prototype.sendQueries = function sendQueries(requests) {
    var _this = this;

    return _promise2['default'].all(requests.map(function (request) {
      return _this._sendQuery(request).then(function (result) {
        return result.json();
      }).then(function (payload) {
        if (payload.hasOwnProperty('errors')) {
          var error = createRequestError(request, '200', payload);
          request.reject(error);
        } else if (!payload.hasOwnProperty('data')) {
          request.reject(new Error('Server response was missing for query ' + ('`' + request.getDebugName() + '`.')));
        } else {
          request.resolve({ response: payload.data });
        }
      })['catch'](function (error) {
        return request.reject(error);
      });
    }));
  };

  RelayDefaultNetworkLayer.prototype.supports = function supports() {
    // Does not support the only defined option, "defer".
    return false;
  };

  /**
   * Sends a POST request with optional files.
   */


  RelayDefaultNetworkLayer.prototype._sendMutation = function _sendMutation(request) {
    var init = void 0;
    var files = request.getFiles();
    if (files) {
      if (!global.FormData) {
        throw new Error('Uploading files without `FormData` not supported.');
      }
      var formData = new FormData();
      formData.append('query', request.getQueryString());
      formData.append('variables', (0, _stringify2['default'])(request.getVariables()));
      for (var filename in files) {
        if (files.hasOwnProperty(filename)) {
          formData.append(filename, files[filename]);
        }
      }
      init = (0, _extends3['default'])({}, this._init, {
        body: formData,
        method: 'POST'
      });
    } else {
      init = (0, _extends3['default'])({}, this._init, {
        body: (0, _stringify2['default'])({
          query: request.getQueryString(),
          variables: request.getVariables()
        }),
        headers: (0, _extends3['default'])({}, this._init.headers, {
          'Accept': '*/*',
          'Content-Type': 'application/json'
        }),
        method: 'POST'
      });
    }
    return require('fbjs/lib/fetch')(this._uri, init).then(function (response) {
      return throwOnServerError(request, response);
    });
  };

  /**
   * Sends a POST request and retries if the request fails or times out.
   */


  RelayDefaultNetworkLayer.prototype._sendQuery = function _sendQuery(request) {
    return require('fbjs/lib/fetchWithRetries')(this._uri, (0, _extends3['default'])({}, this._init, {
      body: (0, _stringify2['default'])({
        query: request.getQueryString(),
        variables: request.getVariables()
      }),
      headers: (0, _extends3['default'])({}, this._init.headers, {
        'Accept': '*/*',
        'Content-Type': 'application/json'
      }),
      method: 'POST'
    }));
  };

  return RelayDefaultNetworkLayer;
}();

/**
 * Rejects HTTP responses with a status code that is not >= 200 and < 300.
 * This is done to follow the internal behavior of `fetchWithRetries`.
 */


function throwOnServerError(request, response) {
  if (response.status >= 200 && response.status < 300) {
    return response;
  } else {
    return response.text().then(function (payload) {
      throw createRequestError(request, response.status, payload);
    });
  }
}

/**
 * Formats an error response from GraphQL server request.
 */
function formatRequestErrors(request, errors) {
  var CONTEXT_BEFORE = 20;
  var CONTEXT_LENGTH = 60;

  var queryLines = request.getQueryString().split('\n');
  return errors.map(function (_ref, ii) {
    var locations = _ref.locations;
    var message = _ref.message;

    var prefix = ii + 1 + '. ';
    var indent = ' '.repeat(prefix.length);

    //custom errors thrown in graphql-server may not have locations
    var locationMessage = locations ? '\n' + locations.map(function (_ref2) {
      var column = _ref2.column;
      var line = _ref2.line;

      var queryLine = queryLines[line - 1];
      var offset = Math.min(column - 1, CONTEXT_BEFORE);
      return [queryLine.substr(column - 1 - offset, CONTEXT_LENGTH), ' '.repeat(Math.max(0, offset)) + '^^^'].map(function (messageLine) {
        return indent + messageLine;
      }).join('\n');
    }).join('\n') : '';

    return prefix + message + locationMessage;
  }).join('\n');
}

function createRequestError(request, responseStatus, payload) {
  var requestType = request instanceof require('./RelayMutationRequest') ? 'mutation' : 'query';
  var errorReason = typeof payload === 'object' ? formatRequestErrors(request, payload.errors) : 'Server response had an error status: ' + responseStatus;
  var error = new Error('Server request for ' + requestType + ' `' + request.getDebugName() + '` ' + ('failed for the following reasons:\n\n' + errorReason));
  error.source = payload;
  error.status = responseStatus;
  return error;
}

module.exports = RelayDefaultNetworkLayer;