'use strict';

exports.__esModule = true;

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var _invariant = require('invariant');

var _invariant2 = _interopRequireDefault(_invariant);

var _createMemoryHistory = require('./createMemoryHistory');

var _createMemoryHistory2 = _interopRequireDefault(_createMemoryHistory);

var _createTransitionManager = require('./createTransitionManager');

var _createTransitionManager2 = _interopRequireDefault(_createTransitionManager);

var _RouteUtils = require('./RouteUtils');

var _RouterUtils = require('./RouterUtils');

/**
 * A high-level API to be used for server-side rendering.
 *
 * This function matches a location to a set of routes and calls
 * callback(error, redirectLocation, renderProps) when finished.
 *
 * Note: You probably don't want to use this in a browser unless you're using
 * server-side rendering with async routes.
 */
function match(_ref, callback) {
  var history = _ref.history;
  var routes = _ref.routes;
  var location = _ref.location;

  var options = _objectWithoutProperties(_ref, ['history', 'routes', 'location']);

  !(history || location) ? process.env.NODE_ENV !== 'production' ? _invariant2['default'](false, 'match needs a history or a location') : _invariant2['default'](false) : undefined;

  history = history ? history : _createMemoryHistory2['default'](options);
  var transitionManager = _createTransitionManager2['default'](history, _RouteUtils.createRoutes(routes));

  var unlisten = undefined;

  if (location) {
    // Allow match({ location: '/the/path', ... })
    location = history.createLocation(location);
  } else {
    // Pick up the location from the history via synchronous history.listen
    // call if needed.
    unlisten = history.listen(function (historyLocation) {
      location = historyLocation;
    });
  }

  var router = _RouterUtils.createRouterObject(history, transitionManager);
  history = _RouterUtils.createRoutingHistory(history, transitionManager);

  transitionManager.match(location, function (error, redirectLocation, nextState) {
    callback(error, redirectLocation, nextState && _extends({}, nextState, {
      history: history,
      router: router,
      matchContext: { history: history, transitionManager: transitionManager, router: router }
    }));

    // Defer removing the listener to here to prevent DOM histories from having
    // to unwind DOM event listeners unnecessarily, in case callback renders a
    // <Router> and attaches another history listener.
    if (unlisten) {
      unlisten();
    }
  });
}

exports['default'] = match;
module.exports = exports['default'];