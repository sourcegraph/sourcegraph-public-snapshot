/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayReadyState
 * 
 */

'use strict';

var _extends3 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _stringify2 = _interopRequireDefault(require('babel-runtime/core-js/json/stringify'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @internal
 */

var RelayReadyState = function () {
  function RelayReadyState(onReadyStateChange) {
    (0, _classCallCheck3['default'])(this, RelayReadyState);

    this._onReadyStateChange = onReadyStateChange;
    this._readyState = {
      aborted: false,
      done: false,
      error: null,
      events: [],
      ready: false,
      stale: false
    };
    this._scheduled = false;
  }

  RelayReadyState.prototype.update = function update(nextReadyState, newEvents) {
    var prevReadyState = this._readyState;
    if (prevReadyState.aborted) {
      return;
    }
    if (prevReadyState.done || prevReadyState.error) {
      if (nextReadyState.stale) {
        if (prevReadyState.error) {
          this._mergeState(nextReadyState, newEvents);
        }
        // Do nothing if stale data comes after server data.
      } else if (!nextReadyState.aborted) {
          require('fbjs/lib/warning')(false, 'RelayReadyState: Invalid state change from `%s` to `%s`.', (0, _stringify2['default'])(prevReadyState), (0, _stringify2['default'])(nextReadyState));
        }
      return;
    }
    this._mergeState(nextReadyState, newEvents);
  };

  RelayReadyState.prototype._mergeState = function _mergeState(nextReadyState, newEvents) {
    var _this = this;

    this._readyState = (0, _extends3['default'])({}, this._readyState, nextReadyState, {
      events: newEvents && newEvents.length ? [].concat(this._readyState.events, newEvents) : this._readyState.events
    });
    if (this._scheduled) {
      return;
    }
    this._scheduled = true;
    require('fbjs/lib/resolveImmediate')(function () {
      _this._scheduled = false;
      _this._onReadyStateChange(_this._readyState);
    });
  };

  return RelayReadyState;
}();

module.exports = RelayReadyState;