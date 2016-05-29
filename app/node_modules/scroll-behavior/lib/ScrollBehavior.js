'use strict';

exports.__esModule = true;

var _off = require('dom-helpers/events/off');

var _off2 = _interopRequireDefault(_off);

var _on = require('dom-helpers/events/on');

var _on2 = _interopRequireDefault(_on);

var _scrollLeft = require('dom-helpers/query/scrollLeft');

var _scrollLeft2 = _interopRequireDefault(_scrollLeft);

var _scrollTop = require('dom-helpers/query/scrollTop');

var _scrollTop2 = _interopRequireDefault(_scrollTop);

var _requestAnimationFrame = require('dom-helpers/util/requestAnimationFrame');

var _requestAnimationFrame2 = _interopRequireDefault(_requestAnimationFrame);

var _Actions = require('history/lib/Actions');

var _DOMStateStorage = require('history/lib/DOMStateStorage');

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } } /* eslint-disable no-underscore-dangle */

// FIXME: Stop using this gross hack. This won't collide with any actual
// history location keys, but it's dirty to sneakily use the same storage here.
var KEY_PREFIX = 's/';

// Try at most this many times to scroll, to avoid getting stuck.
var MAX_SCROLL_ATTEMPTS = 2;

var ScrollBehavior = function () {
  function ScrollBehavior(history, getCurrentLocation) {
    var _this = this;

    _classCallCheck(this, ScrollBehavior);

    this._onScroll = function () {
      // It's possible that this scroll operation was triggered by what will be a
      // `POP` transition. Instead of updating the saved location immediately, we
      // have to enqueue the update, then potentially cancel it if we observe a
      // location update.
      if (_this._savePositionHandle === null) {
        _this._savePositionHandle = (0, _requestAnimationFrame2.default)(_this._savePosition);
      }

      if (_this._scrollTarget) {
        var _scrollTarget = _this._scrollTarget;
        var xTarget = _scrollTarget[0];
        var yTarget = _scrollTarget[1];

        var x = (0, _scrollLeft2.default)(window);
        var y = (0, _scrollTop2.default)(window);

        if (x === xTarget && y === yTarget) {
          _this._scrollTarget = null;
          _this._cancelCheckScroll();
        }
      }
    };

    this._savePosition = function () {
      _this._savePositionHandle = null;

      // We have to directly update `DOMStateStorage`, because actually updating
      // the location could cause e.g. React Router to re-render the entire page,
      // which would lead to observably bad scroll performance.
      (0, _DOMStateStorage.saveState)(_this._getKey(_this._getCurrentLocation()), [(0, _scrollLeft2.default)(window), (0, _scrollTop2.default)(window)]);
    };

    this._checkScrollPosition = function () {
      _this._checkScrollHandle = null;

      // We can only get here if scrollTarget is set. Every code path that unsets
      // scroll target also cancels the handle to avoid calling this handler.
      // Still, check anyway just in case.
      /* istanbul ignore if: paranoid guard */
      if (!_this._scrollTarget) {
        return;
      }

      var _scrollTarget2 = _this._scrollTarget;
      var x = _scrollTarget2[0];
      var y = _scrollTarget2[1];

      window.scrollTo(x, y);

      ++_this._numScrollAttempts;

      /* istanbul ignore if: paranoid guard */
      if (_this._numScrollAttempts >= MAX_SCROLL_ATTEMPTS) {
        _this._scrollTarget = null;
        return;
      }

      _this._checkScrollHandle = (0, _requestAnimationFrame2.default)(_this._checkScrollPosition);
    };

    this._history = history;
    this._getCurrentLocation = getCurrentLocation;

    // This helps avoid some jankiness in fighting against the browser's
    // default scroll behavior on `POP` transitions.
    /* istanbul ignore if: not supported by any browsers on Travis */
    if ('scrollRestoration' in window.history) {
      this._oldScrollRestoration = window.history.scrollRestoration;
      window.history.scrollRestoration = 'manual';
    } else {
      this._oldScrollRestoration = null;
    }

    this._savePositionHandle = null;
    this._checkScrollHandle = null;
    this._scrollTarget = null;
    this._numScrollAttempts = 0;

    // We have to listen to each scroll update rather than to just location
    // updates, because some browsers will update scroll position before
    // emitting the location change.
    (0, _on2.default)(window, 'scroll', this._onScroll);

    this._unlistenBefore = history.listenBefore(function () {
      if (_this._savePositionHandle !== null) {
        _requestAnimationFrame2.default.cancel(_this._savePositionHandle);
        _this._savePositionHandle = null;
      }
    });
  }

  ScrollBehavior.prototype.stop = function stop() {
    /* istanbul ignore if: not supported by any browsers on Travis */
    if (this._oldScrollRestoration) {
      window.history.scrollRestoration = this._oldScrollRestoration;
    }

    (0, _off2.default)(window, 'scroll', this._onScroll);
    this._cancelCheckScroll();

    this._unlistenBefore();
  };

  ScrollBehavior.prototype.updateScroll = function updateScroll(scrollPosition) {
    // Whatever we were doing before isn't relevant any more.
    this._cancelCheckScroll();

    if (scrollPosition && !Array.isArray(scrollPosition)) {
      this._scrollTarget = this._getDefaultScrollTarget();
    } else {
      this._scrollTarget = scrollPosition;
    }

    // Check the scroll position to see if we even need to scroll.
    this._onScroll();

    if (!this._scrollTarget) {
      return;
    }

    this._numScrollAttempts = 0;
    this._checkScrollPosition();
  };

  ScrollBehavior.prototype.readPosition = function readPosition(location) {
    return (0, _DOMStateStorage.readState)(this._getKey(location));
  };

  ScrollBehavior.prototype._getKey = function _getKey(location) {
    // Use fallback key when actual key is unavailable.
    var key = location.key || this._history.createPath(location);

    return '' + KEY_PREFIX + key;
  };

  ScrollBehavior.prototype._cancelCheckScroll = function _cancelCheckScroll() {
    if (this._checkScrollHandle !== null) {
      _requestAnimationFrame2.default.cancel(this._checkScrollHandle);
      this._checkScrollHandle = null;
    }
  };

  ScrollBehavior.prototype._getDefaultScrollTarget = function _getDefaultScrollTarget() {
    var location = this._getCurrentLocation();
    if (location.action === _Actions.PUSH) {
      return [0, 0];
    }

    return this.readPosition(location) || [0, 0];
  };

  return ScrollBehavior;
}();

exports.default = ScrollBehavior;
module.exports = exports['default'];