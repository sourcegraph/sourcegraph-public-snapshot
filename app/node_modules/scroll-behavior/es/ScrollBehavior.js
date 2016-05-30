function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

/* eslint-disable no-underscore-dangle */

import off from 'dom-helpers/events/off';
import on from 'dom-helpers/events/on';
import scrollLeft from 'dom-helpers/query/scrollLeft';
import scrollTop from 'dom-helpers/query/scrollTop';
import requestAnimationFrame from 'dom-helpers/util/requestAnimationFrame';
import { PUSH } from 'history/lib/Actions';
import { readState, saveState } from 'history/lib/DOMStateStorage';

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
        _this._savePositionHandle = requestAnimationFrame(_this._savePosition);
      }

      if (_this._scrollTarget) {
        var _scrollTarget = _this._scrollTarget;
        var xTarget = _scrollTarget[0];
        var yTarget = _scrollTarget[1];

        var x = scrollLeft(window);
        var y = scrollTop(window);

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
      saveState(_this._getKey(_this._getCurrentLocation()), [scrollLeft(window), scrollTop(window)]);
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

      _this._checkScrollHandle = requestAnimationFrame(_this._checkScrollPosition);
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
    on(window, 'scroll', this._onScroll);

    this._unlistenBefore = history.listenBefore(function () {
      if (_this._savePositionHandle !== null) {
        requestAnimationFrame.cancel(_this._savePositionHandle);
        _this._savePositionHandle = null;
      }
    });
  }

  ScrollBehavior.prototype.stop = function stop() {
    /* istanbul ignore if: not supported by any browsers on Travis */
    if (this._oldScrollRestoration) {
      window.history.scrollRestoration = this._oldScrollRestoration;
    }

    off(window, 'scroll', this._onScroll);
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
    return readState(this._getKey(location));
  };

  ScrollBehavior.prototype._getKey = function _getKey(location) {
    // Use fallback key when actual key is unavailable.
    var key = location.key || this._history.createPath(location);

    return '' + KEY_PREFIX + key;
  };

  ScrollBehavior.prototype._cancelCheckScroll = function _cancelCheckScroll() {
    if (this._checkScrollHandle !== null) {
      requestAnimationFrame.cancel(this._checkScrollHandle);
      this._checkScrollHandle = null;
    }
  };

  ScrollBehavior.prototype._getDefaultScrollTarget = function _getDefaultScrollTarget() {
    var location = this._getCurrentLocation();
    if (location.action === PUSH) {
      return [0, 0];
    }

    return this.readPosition(location) || [0, 0];
  };

  return ScrollBehavior;
}();

export default ScrollBehavior;