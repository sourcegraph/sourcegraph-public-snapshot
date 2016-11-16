'use strict';

exports.__esModule = true;
exports.locationShape = exports.routerShape = undefined;

var _react = require('react');

var func = _react.PropTypes.func,
    object = _react.PropTypes.object,
    shape = _react.PropTypes.shape,
    string = _react.PropTypes.string;
var routerShape = exports.routerShape = shape({
  push: func.isRequired,
  replace: func.isRequired,
  go: func.isRequired,
  goBack: func.isRequired,
  goForward: func.isRequired,
  setRouteLeaveHook: func.isRequired,
  isActive: func.isRequired
});

var locationShape = exports.locationShape = shape({
  pathname: string.isRequired,
  search: string.isRequired,
  state: object,
  action: string.isRequired,
  key: string
});