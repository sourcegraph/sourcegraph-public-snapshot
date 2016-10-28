/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayNetworkDebug
 * 
 */

'use strict';

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

var _promise2 = _interopRequireDefault(require('fbjs/lib/Promise'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var RelayNetworkDebugger = function () {
  function RelayNetworkDebugger(environment) {
    var _this = this;

    (0, _classCallCheck3['default'])(this, RelayNetworkDebugger);

    this._initTime = require('fbjs/lib/performanceNow')();
    this._queryID = 0;
    this._subscription = environment.addNetworkSubscriber(function (request) {
      return _this.logRequest(createDebuggableFromRequest('Relay Query', request));
    }, function (request) {
      return _this.logRequest(createDebuggableFromRequest('Relay Mutation', request));
    });
  }

  RelayNetworkDebugger.prototype.uninstall = function uninstall() {
    this._subscription.remove();
  };

  RelayNetworkDebugger.prototype.logRequest = function logRequest(_ref) {
    var _this2 = this;

    var name = _ref.name;
    var type = _ref.type;
    var promise = _ref.promise;
    var logResult = _ref.logResult;

    var id = this._queryID++;
    var timerName = '[' + id + '] Request Duration';

    console.timeStamp && console.timeStamp('START: [' + id + '] ' + type + ': ' + name + ' →');
    console.time && console.time(timerName);

    var onSettled = function onSettled(error, response) {
      var time = (require('fbjs/lib/performanceNow')() - _this2._initTime) / 1000;
      console.timeStamp && console.timeStamp('← END: [' + id + '] ' + type + ': ' + name);
      var groupName = '%c[' + id + '] ' + type + ': ' + name + ' @ ' + time + 's';
      console.groupCollapsed(groupName, 'color:' + (error ? 'red' : 'black') + ';');
      console.timeEnd && console.timeEnd(timerName);
      logResult(error, response);
      console.groupEnd();
    };

    promise.then(function (response) {
      return onSettled(null, response);
    }, function (error) {
      return onSettled(error, null);
    });
  };

  return RelayNetworkDebugger;
}();

function createDebuggableFromRequest(type, request) {
  return {
    name: request.getDebugName(),
    type: type,
    promise: request.getPromise(),
    logResult: function logResult(error, response) {
      /* eslint-disable no-console-disallow */
      var requestSize = formatSize(require('fbjs/lib/xhrSimpleDataSerializer')({
        q: request.getQueryString(),
        query_params: request.getVariables()
      }).length);
      var requestVariables = request.getVariables();

      console.groupCollapsed('Request Query (Estimated Size: %s)', requestSize);
      console.debug('%c%s\n', 'font-size:10px; color:#333; font-family:mplus-2m-regular,menlo,' + 'monospaced;', request.getQueryString());
      console.groupEnd();

      if ((0, _keys2['default'])(requestVariables).length > 0) {
        console.log('Request Variables\n', request.getVariables());
      }

      error && console.error(error);
      response && console.log(response);
      /* eslint-enable no-console-disallow */
    }
  };
}

var ALL_UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
function formatSize(bytes) {
  var sign = bytes < 0 ? -1 : 1;
  bytes = Math.abs(bytes);
  var i = 0;
  while (bytes >= Math.pow(1024, i + 1) && i < ALL_UNITS.length) {
    i++;
  }
  var value = sign * bytes * 1.0 / Math.pow(1024, i);
  return Number(value.toFixed(2)) + ALL_UNITS[i];
}

var networkDebugger = void 0;

var RelayNetworkDebug = {
  init: function init() {
    var environment = arguments.length <= 0 || arguments[0] === undefined ? require('./RelayPublic').Store : arguments[0];

    networkDebugger && networkDebugger.uninstall();
    // Without `groupCollapsed`, RelayNetworkDebug is too noisy.
    if (console.groupCollapsed) {
      networkDebugger = new RelayNetworkDebugger(environment);
    }
  },
  logRequest: function logRequest(request) {
    networkDebugger && networkDebugger.logRequest(request);
  }
};

module.exports = RelayNetworkDebug;