'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _class, _class2, _temp; /**
                             * @fileOverview Cartesian Grid
                             */


var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _PureRender = require('../util/PureRender');

var _PureRender2 = _interopRequireDefault(_PureRender);

var _ReactUtils = require('../util/ReactUtils');

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var CartesianGrid = (0, _PureRender2.default)(_class = (_temp = _class2 = function (_Component) {
  _inherits(CartesianGrid, _Component);

  function CartesianGrid() {
    _classCallCheck(this, CartesianGrid);

    return _possibleConstructorReturn(this, Object.getPrototypeOf(CartesianGrid).apply(this, arguments));
  }

  _createClass(CartesianGrid, [{
    key: 'renderHorizontal',


    /**
     * Draw the horizontal grid lines
     * @return {Group} Horizontal lines
     */
    value: function renderHorizontal() {
      var _props = this.props;
      var x = _props.x;
      var width = _props.width;
      var horizontalPoints = _props.horizontalPoints;


      if (!horizontalPoints || !horizontalPoints.length) {
        return null;
      }

      var props = (0, _ReactUtils.getPresentationAttributes)(this.props);
      var items = horizontalPoints.map(function (entry, i) {
        return _react2.default.createElement('line', _extends({}, props, { key: 'line-' + i, x1: x, y1: entry, x2: x + width, y2: entry }));
      });

      return _react2.default.createElement(
        'g',
        { className: 'recharts-cartesian-grid-horizontal' },
        items
      );
    }

    /**
     * Draw vertical grid lines
     * @return {Group} Vertical lines
     */

  }, {
    key: 'renderVertical',
    value: function renderVertical() {
      var _props2 = this.props;
      var y = _props2.y;
      var height = _props2.height;
      var verticalPoints = _props2.verticalPoints;


      if (!verticalPoints || !verticalPoints.length) {
        return null;
      }

      var props = (0, _ReactUtils.getPresentationAttributes)(this.props);

      var items = verticalPoints.map(function (entry, i) {
        return _react2.default.createElement('line', _extends({}, props, { key: 'line-' + i, x1: entry, y1: y, x2: entry, y2: y + height }));
      });

      return _react2.default.createElement(
        'g',
        { className: 'recharts-cartesian-grid-vertical' },
        items
      );
    }
  }, {
    key: 'render',
    value: function render() {
      var _props3 = this.props;
      var width = _props3.width;
      var height = _props3.height;
      var horizontal = _props3.horizontal;
      var vertical = _props3.vertical;


      if (width <= 0 || height <= 0) {
        return null;
      }

      return _react2.default.createElement(
        'g',
        { className: 'recharts-cartesian-grid' },
        horizontal && this.renderHorizontal(),
        vertical && this.renderVertical()
      );
    }
  }]);

  return CartesianGrid;
}(_react.Component), _class2.displayName = 'CartesianGrid', _class2.propTypes = _extends({}, _ReactUtils.PRESENTATION_ATTRIBUTES, {
  x: _react.PropTypes.number,
  y: _react.PropTypes.number,
  width: _react.PropTypes.number,
  height: _react.PropTypes.number,
  horizontal: _react.PropTypes.bool,
  vertical: _react.PropTypes.bool,
  horizontalPoints: _react.PropTypes.arrayOf(_react.PropTypes.number),
  verticalPoints: _react.PropTypes.arrayOf(_react.PropTypes.number)
}), _class2.defaultProps = {
  x: 0,
  y: 0,
  width: 0,
  height: 0,
  horizontal: true,
  vertical: true,
  // The ordinates of horizontal grid lines
  horizontalPoints: [],
  // The abscissas of vertical grid lines
  verticalPoints: [],

  stroke: '#ccc',
  fill: 'none'
}, _temp)) || _class;

exports.default = CartesianGrid;