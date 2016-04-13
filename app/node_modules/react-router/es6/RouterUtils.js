'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

export { createRouterObject };
export { createRoutingHistory };
import deprecateObjectProperties from './deprecateObjectProperties';

function createRouterObject(history, transitionManager) {
  return _extends({}, history, {
    setRouteLeaveHook: transitionManager.listenBeforeLeavingRoute,
    isActive: transitionManager.isActive
  });
}

// deprecated

function createRoutingHistory(history, transitionManager) {
  history = _extends({}, history, transitionManager);

  if (process.env.NODE_ENV !== 'production') {
    history = deprecateObjectProperties(history, '`props.history` and `context.history` are deprecated. Please use `context.router`. http://tiny.cc/router-contextchanges');
  }

  return history;
}