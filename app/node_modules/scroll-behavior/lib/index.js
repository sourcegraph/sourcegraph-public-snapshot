'use strict';

exports.__esModule = true;

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

exports.default = withScroll;

var _ScrollBehavior = require('./ScrollBehavior');

var _ScrollBehavior2 = _interopRequireDefault(_ScrollBehavior);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function withScroll(history, shouldUpdateScroll) {
  // history will invoke the onChange callback synchronously, so
  // currentLocation will always be defined when needed.
  var currentLocation = null;

  function getCurrentLocation() {
    return currentLocation;
  }

  var listeners = [];
  var scrollBehavior = null;

  function onChange(location) {
    var prevLocation = currentLocation;
    currentLocation = location;

    listeners.forEach(function (listener) {
      return listener(location);
    });

    var scrollPosition = void 0;
    if (!shouldUpdateScroll) {
      scrollPosition = true;
    } else {
      scrollPosition = shouldUpdateScroll.call(scrollBehavior, prevLocation, location);
    }

    scrollBehavior.updateScroll(scrollPosition);
  }

  var unlisten = null;

  function listen(listener) {
    if (listeners.length === 0) {
      scrollBehavior = new _ScrollBehavior2.default(history, getCurrentLocation);
      unlisten = history.listen(onChange);
    }

    listeners.push(listener);
    listener(currentLocation);

    return function () {
      listeners = listeners.filter(function (item) {
        return item !== listener;
      });

      if (listeners.length === 0) {
        scrollBehavior.stop();
        unlisten();
      }
    };
  }

  return _extends({}, history, {
    listen: listen
  });
}
module.exports = exports['default'];