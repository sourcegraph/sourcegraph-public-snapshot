'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _isNumber2 = require('lodash/isNumber');

var _isNumber3 = _interopRequireDefault(_isNumber2);

var _isFunction2 = require('lodash/isFunction');

var _isFunction3 = _interopRequireDefault(_isFunction2);

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _class, _class2, _temp; /**
                             * @fileOverview Legend
                             */


var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _PureRender = require('../util/PureRender');

var _PureRender2 = _interopRequireDefault(_PureRender);

var _server = require('react-dom/server');

var _server2 = _interopRequireDefault(_server);

var _Surface = require('../container/Surface');

var _Surface2 = _interopRequireDefault(_Surface);

var _DefaultLegendContent = require('./DefaultLegendContent');

var _DefaultLegendContent2 = _interopRequireDefault(_DefaultLegendContent);

var _DOMUtils = require('../util/DOMUtils');

var _ReactUtils = require('../util/ReactUtils');

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var SIZE = 32;

var renderContent = function renderContent(content, props) {
  if (_react2.default.isValidElement(content)) {
    return _react2.default.cloneElement(content, props);
  } else if ((0, _isFunction3.default)(content)) {
    return content(props);
  }

  return _react2.default.createElement(_DefaultLegendContent2.default, props);
};

var Legend = (0, _PureRender2.default)(_class = (_temp = _class2 = function (_Component) {
  _inherits(Legend, _Component);

  function Legend() {
    _classCallCheck(this, Legend);

    return _possibleConstructorReturn(this, Object.getPrototypeOf(Legend).apply(this, arguments));
  }

  _createClass(Legend, [{
    key: 'getDefaultPosition',
    value: function getDefaultPosition(style) {
      var _props = this.props;
      var layout = _props.layout;
      var align = _props.align;
      var verticalAlign = _props.verticalAlign;
      var margin = _props.margin;
      var chartWidth = _props.chartWidth;
      var chartHeight = _props.chartHeight;

      var hPos = void 0;
      var vPos = void 0;

      if (!style || (style.left === undefined || style.left === null) && (style.right === undefined || style.right === null)) {
        if (align === 'center' && layout === 'vertical') {
          var box = Legend.getLegendBBox(this.props) || { width: 0 };
          hPos = { left: ((chartWidth || 0) - box.width) / 2 };
        } else {
          hPos = align === 'right' ? { right: margin && margin.right || 0 } : { left: margin && margin.left || 0 };
        }
      }

      if (!style || (style.top === undefined || style.top === null) && (style.bottom === undefined || style.bottom === null)) {
        if (verticalAlign === 'middle') {
          var _box = Legend.getLegendBBox(this.props) || { height: 0 };
          vPos = { top: ((chartHeight || 0) - _box.height) / 2 };
        } else {
          vPos = verticalAlign === 'bottom' ? { bottom: margin && margin.bottom || 0 } : { top: margin && margin.top || 0 };
        }
      }

      return _extends({}, hPos, vPos);
    }
  }, {
    key: 'render',
    value: function render() {
      var _props2 = this.props;
      var content = _props2.content;
      var width = _props2.width;
      var height = _props2.height;
      var layout = _props2.layout;
      var wrapperStyle = _props2.wrapperStyle;

      var outerStyle = _extends({
        position: 'absolute',
        width: width || 'auto',
        height: height || 'auto'
      }, this.getDefaultPosition(wrapperStyle), wrapperStyle);

      return _react2.default.createElement(
        'div',
        { className: 'recharts-legend-wrapper', style: outerStyle },
        renderContent(content, this.props)
      );
    }
  }], [{
    key: 'getWithHeight',
    value: function getWithHeight(item, chartWidth, chartHeight) {
      var layout = item.props.layout;


      if (layout === 'vertical' && (0, _isNumber3.default)(item.props.height)) {
        return {
          height: item.props.height
        };
      } else if (layout === 'horizontal') {
        return {
          width: item.props.width || chartWidth
        };
      }

      return null;
    }
  }, {
    key: 'getLegendBBox',
    value: function getLegendBBox(props) {
      if (!(0, _ReactUtils.isSsr)()) {
        var content = props.content;
        var width = props.width;
        var height = props.height;
        var wrapperStyle = props.wrapperStyle;

        var contentHtml = _server2.default.renderToStaticMarkup(renderContent(content, props));
        var style = _extends({
          position: 'absolute',
          width: width || 'auto',
          height: height || 'auto'
        }, wrapperStyle, {
          top: -20000,
          left: 0,
          display: 'block'
        });
        var wrapper = document.createElement('div');

        wrapper.setAttribute('style', (0, _DOMUtils.getStyleString)(style));
        wrapper.innerHTML = contentHtml;
        document.body.appendChild(wrapper);
        var box = wrapper.getBoundingClientRect();

        document.body.removeChild(wrapper);

        return box;
      }

      return null;
    }
  }]);

  return Legend;
}(_react.Component), _class2.displayName = 'Legend', _class2.propTypes = {
  content: _react.PropTypes.oneOfType([_react.PropTypes.element, _react.PropTypes.func]),
  wrapperStyle: _react.PropTypes.object,
  chartWidth: _react.PropTypes.number,
  chartHeight: _react.PropTypes.number,
  width: _react.PropTypes.number,
  height: _react.PropTypes.number,
  iconSize: _react.PropTypes.number,
  layout: _react.PropTypes.oneOf(['horizontal', 'vertical']),
  align: _react.PropTypes.oneOf(['center', 'left', 'right']),
  verticalAlign: _react.PropTypes.oneOf(['top', 'bottom', 'middle']),
  margin: _react.PropTypes.shape({
    top: _react.PropTypes.number,
    left: _react.PropTypes.number,
    bottom: _react.PropTypes.number,
    right: _react.PropTypes.number
  }),
  payload: _react.PropTypes.arrayOf(_react.PropTypes.shape({
    value: _react.PropTypes.any,
    id: _react.PropTypes.any,
    type: _react.PropTypes.oneOf(['line', 'scatter', 'square', 'rect'])
  }))
}, _class2.defaultProps = {
  iconSize: 14,
  layout: 'horizontal',
  align: 'center',
  verticalAlign: 'bottom'
}, _temp)) || _class;

exports.default = Legend;