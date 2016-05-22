'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.translateStyle = exports.AnimateGroup = exports.configBezier = exports.configSpring = undefined;

var _Animate = require('./Animate');

var _Animate2 = _interopRequireDefault(_Animate);

var _easing = require('./easing');

var _util = require('./util');

var _AnimateGroup = require('./AnimateGroup');

var _AnimateGroup2 = _interopRequireDefault(_AnimateGroup);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

exports.configSpring = _easing.configSpring;
exports.configBezier = _easing.configBezier;
exports.AnimateGroup = _AnimateGroup2.default;
exports.translateStyle = _util.translateStyle;
exports.default = _Animate2.default;