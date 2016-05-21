'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _class, _class2, _temp; /**
                             * @fileOverview Default Legend Content
                             */


var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _PureRender = require('../util/PureRender');

var _PureRender2 = _interopRequireDefault(_PureRender);

var _Surface = require('../container/Surface');

var _Surface2 = _interopRequireDefault(_Surface);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var SIZE = 32;

var DefaultLegendContent = (0, _PureRender2.default)(_class = (_temp = _class2 = function (_Component) {
  _inherits(DefaultLegendContent, _Component);

  function DefaultLegendContent() {
    _classCallCheck(this, DefaultLegendContent);

    return _possibleConstructorReturn(this, Object.getPrototypeOf(DefaultLegendContent).apply(this, arguments));
  }

  _createClass(DefaultLegendContent, [{
    key: 'renderIcon',


    /**
     * Render the path of icon
     * @param {Object} data Data of each legend item
     * @return {String} Path element
     */
    value: function renderIcon(data) {
      var halfSize = SIZE / 2;
      var sixthSize = SIZE / 6;
      var thirdSize = SIZE / 3;
      var path = void 0;
      var fill = data.color;
      var stroke = data.color;

      switch (data.type) {
        case 'line':
          fill = 'none';
          path = 'M0,' + halfSize + 'h' + thirdSize + 'A' + sixthSize + ',' + sixthSize + ',' + ('0,1,1,' + 2 * thirdSize + ',' + halfSize) + ('H' + SIZE + 'M' + 2 * thirdSize + ',' + halfSize) + ('A' + sixthSize + ',' + sixthSize + ',0,1,1,' + thirdSize + ',' + halfSize);
          break;
        case 'scatter':
          stroke = 'none';
          path = 'M' + halfSize + ',0A' + halfSize + ',' + halfSize + ',0,1,1,' + halfSize + ',' + SIZE + ('A' + halfSize + ',' + halfSize + ',0,1,1,' + halfSize + ',0Z');
          break;
        case 'rect':
          stroke = 'none';
          path = 'M0,' + SIZE / 8 + 'h' + SIZE + 'v' + SIZE * 3 / 4 + 'h' + -SIZE + 'z';
          break;
        default:
          stroke = 'none';
          path = 'M0,0h' + SIZE + 'v' + SIZE + 'h' + -SIZE + 'z';
          break;
      }

      return _react2.default.createElement('path', {
        strokeWidth: 4,
        fill: fill,
        stroke: stroke,
        d: path,
        className: 'recharts-legend-icon'
      });
    }

    /**
     * Draw items of legend
     * @return {ReactElement} Items
     */

  }, {
    key: 'renderItems',
    value: function renderItems() {
      var _this2 = this;

      var _props = this.props;
      var payload = _props.payload;
      var iconSize = _props.iconSize;
      var layout = _props.layout;

      var viewBox = { x: 0, y: 0, width: SIZE, height: SIZE };
      var itemStyle = {
        display: layout === 'horizontal' ? 'inline-block' : 'block',
        marginRight: 10
      };
      var svgStyle = { display: 'inline-block', verticalAlign: 'middle', marginRight: 4 };

      return payload.map(function (entry, i) {
        return _react2.default.createElement(
          'li',
          {
            className: 'recharts-legend-item legend-item-' + i,
            style: itemStyle,
            key: 'legend-item-' + i
          },
          _react2.default.createElement(
            _Surface2.default,
            { width: iconSize, height: iconSize, viewBox: viewBox, style: svgStyle },
            _this2.renderIcon(entry)
          ),
          _react2.default.createElement(
            'span',
            { className: 'recharts-legend-item-text' },
            entry.value
          )
        );
      });
    }
  }, {
    key: 'render',
    value: function render() {
      var _props2 = this.props;
      var payload = _props2.payload;
      var layout = _props2.layout;
      var align = _props2.align;


      if (!payload || !payload.length) {
        return null;
      }

      var finalStyle = {
        padding: 0,
        margin: 0,
        textAlign: layout === 'horizontal' ? align : 'left'
      };

      return _react2.default.createElement(
        'ul',
        { className: 'recharts-default-legend', style: finalStyle },
        this.renderItems()
      );
    }
  }]);

  return DefaultLegendContent;
}(_react.Component), _class2.displayName = 'Legend', _class2.propTypes = {
  content: _react.PropTypes.element,
  iconSize: _react.PropTypes.number,
  layout: _react.PropTypes.oneOf(['horizontal', 'vertical']),
  align: _react.PropTypes.oneOf(['center', 'left', 'right']),
  verticalAlign: _react.PropTypes.oneOf(['top', 'bottom', 'middle']),
  payload: _react.PropTypes.arrayOf(_react.PropTypes.shape({
    value: _react.PropTypes.any,
    id: _react.PropTypes.any,
    type: _react.PropTypes.oneOf(['line', 'scatter', 'square', 'rect'])
  }))
}, _class2.defaultProps = {
  iconSize: 14,
  layout: 'horizontal',
  align: 'center',
  verticalAlign: 'middle'
}, _temp)) || _class;

exports.default = DefaultLegendContent;