'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _class, _class2, _temp; /**
                             * @fileOverview Composed Chart
                             */


var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _classnames = require('classnames');

var _classnames2 = _interopRequireDefault(_classnames);

var _Surface = require('../container/Surface');

var _Surface2 = _interopRequireDefault(_Surface);

var _Layer = require('../container/Layer');

var _Layer2 = _interopRequireDefault(_Layer);

var _Tooltip = require('../component/Tooltip');

var _Tooltip2 = _interopRequireDefault(_Tooltip);

var _Line = require('../cartesian/Line');

var _Line2 = _interopRequireDefault(_Line);

var _Bar = require('../cartesian/Bar');

var _Bar2 = _interopRequireDefault(_Bar);

var _Area = require('../cartesian/Area');

var _Area2 = _interopRequireDefault(_Area);

var _Curve = require('../shape/Curve');

var _Curve2 = _interopRequireDefault(_Curve);

var _Dot = require('../shape/Dot');

var _Dot2 = _interopRequireDefault(_Dot);

var _Rectangle = require('../shape/Rectangle');

var _Rectangle2 = _interopRequireDefault(_Rectangle);

var _generateCategoricalChart = require('./generateCategoricalChart');

var _generateCategoricalChart2 = _interopRequireDefault(_generateCategoricalChart);

var _DataUtils = require('../util/DataUtils');

var _ReactUtils = require('../util/ReactUtils');

var _PureRender = require('../util/PureRender');

var _PureRender2 = _interopRequireDefault(_PureRender);

var _CartesianUtils = require('../util/CartesianUtils');

var _AreaChart = require('./AreaChart');

var _LineChart = require('./LineChart');

var _BarChart = require('./BarChart');

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var ComposedChart = (0, _PureRender2.default)(_class = (_temp = _class2 = function (_Component) {
  _inherits(ComposedChart, _Component);

  function ComposedChart() {
    _classCallCheck(this, ComposedChart);

    return _possibleConstructorReturn(this, Object.getPrototypeOf(ComposedChart).apply(this, arguments));
  }

  _createClass(ComposedChart, [{
    key: 'renderCursor',
    value: function renderCursor(xAxisMap, yAxisMap, offset) {
      var _props = this.props;
      var children = _props.children;
      var isTooltipActive = _props.isTooltipActive;

      var tooltipItem = (0, _ReactUtils.findChildByType)(children, _Tooltip2.default);
      if (!tooltipItem || !tooltipItem.props.cursor || !isTooltipActive) {
        return null;
      }

      var _props2 = this.props;
      var layout = _props2.layout;
      var activeTooltipIndex = _props2.activeTooltipIndex;

      var axisMap = layout === 'horizontal' ? xAxisMap : yAxisMap;
      var axis = (0, _DataUtils.getAnyElementOfObject)(axisMap);
      var bandSize = (0, _DataUtils.getBandSizeOfScale)(axis.scale);

      var ticks = (0, _CartesianUtils.getTicksOfAxis)(axis);
      var start = ticks[activeTooltipIndex].coordinate;
      var cursorProps = _extends({
        fill: '#f1f1f1'
      }, (0, _ReactUtils.getPresentationAttributes)(tooltipItem.props.cursor), {
        x: layout === 'horizontal' ? start : offset.left + 0.5,
        y: layout === 'horizontal' ? offset.top + 0.5 : start,
        width: layout === 'horizontal' ? bandSize : offset.width - 1,
        height: layout === 'horizontal' ? offset.height - 1 : bandSize
      });

      return _react2.default.isValidElement(tooltipItem.props.cursor) ? _react2.default.cloneElement(tooltipItem.props.cursor, cursorProps) : _react2.default.createElement(_Rectangle2.default, cursorProps);
    }
  }, {
    key: 'render',
    value: function render() {
      var _props3 = this.props;
      var xAxisMap = _props3.xAxisMap;
      var yAxisMap = _props3.yAxisMap;
      var offset = _props3.offset;
      var graphicalItems = _props3.graphicalItems;
      var stackGroups = _props3.stackGroups;

      var areaItems = graphicalItems.filter(function (item) {
        return item.type.displayName === 'Area';
      });
      var lineItems = graphicalItems.filter(function (item) {
        return item.type.displayName === 'Line';
      });
      var barItems = graphicalItems.filter(function (item) {
        return item.type.displayName === 'Bar';
      });

      return _react2.default.createElement(
        _Layer2.default,
        { className: 'recharts-composed' },
        this.renderCursor(xAxisMap, yAxisMap, offset),
        _react2.default.createElement(_AreaChart.AreaChart, _extends({}, this.props, { graphicalItems: areaItems, isComposed: true })),
        _react2.default.createElement(_BarChart.BarChart, _extends({}, this.props, { graphicalItems: barItems, isComposed: true })),
        _react2.default.createElement(_LineChart.LineChart, _extends({}, this.props, { graphicalItems: lineItems, isComposed: true }))
      );
    }
  }]);

  return ComposedChart;
}(_react.Component), _class2.displayName = 'ComposedChart', _class2.propTypes = {
  layout: _react.PropTypes.oneOf(['horizontal', 'vertical']),
  dataStartIndex: _react.PropTypes.number,
  dataEndIndex: _react.PropTypes.number,
  isTooltipActive: _react.PropTypes.bool,
  activeTooltipIndex: _react.PropTypes.number,
  xAxisMap: _react.PropTypes.object,
  yAxisMap: _react.PropTypes.object,
  offset: _react.PropTypes.object,
  graphicalItems: _react.PropTypes.array,
  stackGroups: _react.PropTypes.object,
  children: _react.PropTypes.oneOfType([_react.PropTypes.arrayOf(_react.PropTypes.node), _react.PropTypes.node])
}, _temp)) || _class;

exports.default = (0, _generateCategoricalChart2.default)(ComposedChart, [_Line2.default, _Area2.default, _Bar2.default]);