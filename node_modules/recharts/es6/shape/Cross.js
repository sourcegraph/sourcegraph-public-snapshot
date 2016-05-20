'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _isNumber2 = require('lodash/isNumber');

var _isNumber3 = _interopRequireDefault(_isNumber2);

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _class, _class2, _temp; /**
                             * @fileOverview Cross
                             */


var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _PureRender = require('../util/PureRender');

var _PureRender2 = _interopRequireDefault(_PureRender);

var _classnames = require('classnames');

var _classnames2 = _interopRequireDefault(_classnames);

var _ReactUtils = require('../util/ReactUtils');

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var Cross = (0, _PureRender2.default)(_class = (_temp = _class2 = function (_Component) {
  _inherits(Cross, _Component);

  function Cross() {
    _classCallCheck(this, Cross);

    return _possibleConstructorReturn(this, Object.getPrototypeOf(Cross).apply(this, arguments));
  }

  _createClass(Cross, [{
    key: 'getPath',
    value: function getPath(x, y, width, height, top, left) {
      return 'M' + x + ',' + top + 'v' + height + 'M' + left + ',' + y + 'h' + width;
    }
  }, {
    key: 'render',
    value: function render() {
      var _props = this.props;
      var x = _props.x;
      var y = _props.y;
      var width = _props.width;
      var height = _props.height;
      var top = _props.top;
      var left = _props.left;
      var className = _props.className;


      if (!(0, _isNumber3.default)(x) || !(0, _isNumber3.default)(y) || !(0, _isNumber3.default)(width) || !(0, _isNumber3.default)(height) || !(0, _isNumber3.default)(top) || !(0, _isNumber3.default)(left)) {
        return null;
      }

      return _react2.default.createElement('path', _extends({}, (0, _ReactUtils.getPresentationAttributes)(this.props), {
        className: (0, _classnames2.default)('recharts-cross', className),
        d: this.getPath(x, y, width, height, top, left)
      }));
    }
  }]);

  return Cross;
}(_react.Component), _class2.displayName = 'Cross', _class2.propTypes = _extends({}, _ReactUtils.PRESENTATION_ATTRIBUTES, {
  x: _react.PropTypes.number,
  y: _react.PropTypes.number,
  width: _react.PropTypes.number,
  height: _react.PropTypes.number,
  top: _react.PropTypes.number,
  left: _react.PropTypes.number,
  className: _react.PropTypes.string
}), _class2.defaultProps = {
  x: 0,
  y: 0,
  top: 0,
  left: 0,
  width: 0,
  height: 0,
  stroke: '#000',
  fill: 'none'
}, _temp)) || _class;

exports.default = Cross;