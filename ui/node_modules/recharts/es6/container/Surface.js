'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _classnames = require('classnames');

var _classnames2 = _interopRequireDefault(_classnames);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

/**
 * @fileOverview Surface
 */


var propTypes = {
  width: _react.PropTypes.number.isRequired,
  height: _react.PropTypes.number.isRequired,
  viewBox: _react.PropTypes.shape({
    x: _react.PropTypes.number,
    y: _react.PropTypes.number,
    width: _react.PropTypes.number,
    height: _react.PropTypes.number
  }),
  className: _react.PropTypes.string,
  style: _react.PropTypes.object,
  children: _react.PropTypes.oneOfType([_react.PropTypes.arrayOf(_react.PropTypes.node), _react.PropTypes.node])
};
function Surface(props) {
  var children = props.children;
  var width = props.width;
  var height = props.height;
  var viewBox = props.viewBox;
  var className = props.className;
  var style = props.style;

  var svgView = viewBox || { width: width, height: height, x: 0, y: 0 };
  var layerClass = (0, _classnames2.default)('recharts-surface', className);

  return _react2.default.createElement(
    'svg',
    {
      className: layerClass,
      width: width,
      height: height,
      style: style,
      viewBox: svgView.x + ' ' + svgView.y + ' ' + svgView.width + ' ' + svgView.height,
      xmlns: 'http://www.w3.org/2000/svg', version: '1.1'
    },
    children
  );
}

Surface.propTypes = propTypes;

exports.default = Surface;