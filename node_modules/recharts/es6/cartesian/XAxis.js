'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _class, _class2, _temp; /**
                             * @fileOverview X Axis
                             */


var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _PureRender = require('../util/PureRender');

var _PureRender2 = _interopRequireDefault(_PureRender);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var XAxis = (0, _PureRender2.default)(_class = (_temp = _class2 = function (_Component) {
  _inherits(XAxis, _Component);

  function XAxis() {
    _classCallCheck(this, XAxis);

    return _possibleConstructorReturn(this, Object.getPrototypeOf(XAxis).apply(this, arguments));
  }

  _createClass(XAxis, [{
    key: 'render',
    value: function render() {
      return null;
    }
  }]);

  return XAxis;
}(_react.Component), _class2.displayName = 'XAxis', _class2.propTypes = {
  hide: _react.PropTypes.bool,
  // The name of data displayed in the axis
  name: _react.PropTypes.oneOfType([_react.PropTypes.string, _react.PropTypes.number]),
  // The unit of data displayed in the axis
  unit: _react.PropTypes.oneOfType([_react.PropTypes.string, _react.PropTypes.number]),
  // The unique id of x-axis
  xAxisId: _react.PropTypes.oneOfType([_react.PropTypes.string, _react.PropTypes.number]),
  domain: _react.PropTypes.arrayOf(_react.PropTypes.oneOfType([_react.PropTypes.number, _react.PropTypes.oneOf(['auto', 'dataMin', 'dataMax'])])),
  // The key of data displayed in the axis
  dataKey: _react.PropTypes.oneOfType([_react.PropTypes.string, _react.PropTypes.number]),
  // The width of axis which is usually calculated internally
  width: _react.PropTypes.number,
  // The height of axis, which need to be setted by user
  height: _react.PropTypes.number,
  // The orientation of axis
  orientation: _react.PropTypes.oneOf(['top', 'bottom']),
  type: _react.PropTypes.oneOf(['number', 'category']),
  // Ticks can be any type when the axis is the type of category
  // Ticks must be numbers when the axis is the type of number
  ticks: _react.PropTypes.array,
  // The count of ticks
  tickCount: _react.PropTypes.number,
  // The formatter function of tick
  tickFormatter: _react.PropTypes.func
}, _class2.defaultProps = {
  hide: false,
  orientation: 'bottom',
  width: 0,
  height: 30,
  xAxisId: 0,
  tickCount: 5,
  type: 'category',
  domain: [0, 'auto']
}, _temp)) || _class;

exports.default = XAxis;