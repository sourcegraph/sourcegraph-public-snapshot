'use strict';

Object.defineProperty(exports, "__esModule", {
    value: true
});

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var React = require('react');
var IconBase = require('react-icon-base');

var MdWrapText = function (_React$Component) {
    _inherits(MdWrapText, _React$Component);

    function MdWrapText() {
        _classCallCheck(this, MdWrapText);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWrapText).apply(this, arguments));
    }

    _createClass(MdWrapText, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 18.36q2.7333333333333343 0 4.688333333333333 1.9533333333333331t1.951666666666668 4.686666666666667-1.9500000000000028 4.690000000000001-4.690000000000001 1.9533333333333331h-3.359999999999996v3.360000000000003l-5-5 5-5v3.3599999999999994h3.75q1.3283333333333331 0 2.3433333333333337-1.0166666666666657t1.0166666666666657-2.3416666666666686-1.0166666666666657-2.3433333333333337-2.3433333333333337-1.0166666666666657h-22.11v-3.278333333333336h21.716666666666665z m5-10v3.283333333333333h-26.716666666666665v-3.283333333333333h26.716666666666665z m-26.72 23.280000000000005v-3.283333333333335h10v3.2833333333333314h-10z' })
                )
            );
        }
    }]);

    return MdWrapText;
}(React.Component);

exports.default = MdWrapText;
module.exports = exports['default'];