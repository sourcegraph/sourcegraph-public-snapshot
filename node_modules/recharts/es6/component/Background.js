'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _classnames = require('classnames');

var _classnames2 = _interopRequireDefault(_classnames);

var _PureRender = require('../util/PureRender');

var _PureRender2 = _interopRequireDefault(_PureRender);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; } /**
                                                                                                                                                                                                                              * @fileOverview Background
                                                                                                                                                                                                                              */


var propTypes = {
  x: _react.PropTypes.number,
  y: _react.PropTypes.number,
  width: _react.PropTypes.number,
  height: _react.PropTypes.number,
  strokeWidth: _react.PropTypes.number,
  stroke: _react.PropTypes.string,
  fill: _react.PropTypes.string,
  className: _react.PropTypes.string
};

function Background(props) {
  var className = props.className;

  var others = _objectWithoutProperties(props, ['className']);

  return _react2.default.createElement(
    'g',
    { className: (0, _classnames2.default)('recharts-background', className) },
    _react2.default.createElement('rect', others)
  );
}

Background.propTypes = propTypes;

exports.default = Background;