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

var FaCompass = function (_React$Component) {
    _inherits(FaCompass, _React$Component);

    function FaCompass() {
        _classCallCheck(this, FaCompass);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCompass).apply(this, arguments));
    }

    _createClass(FaCompass, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.142857142857142 24.285714285714285l5.714285714285715-2.8571428571428577-5.714285714285715-2.8571428571428577v5.714285714285715z m8.571428571428573-13.191428571428572v12.1l-11.428571428571429 5.714285714285715v-12.100000000000001z m6.428571428571431 8.905714285714287q0-3.3028571428571425-1.62857142857143-6.094285714285714t-4.4228571428571435-4.42-6.0914285714285725-1.6285714285714281-6.097142857142858 1.6285714285714281-4.417142857142856 4.42-1.6328571428571426 6.094285714285714 1.6285714285714281 6.094285714285714 4.421428571428573 4.419999999999998 6.097142857142856 1.6285714285714334 6.092857142857142-1.62857142857143 4.420000000000002-4.419999999999998 1.6300000000000026-6.094285714285718z m5 0q0 4.665714285714287-2.299999999999997 8.604285714285716t-6.237142857142857 6.238571428571426-8.605714285714292 2.3000000000000043-8.6-2.3000000000000043-6.242857142857143-6.238571428571426-2.295714285714286-8.604285714285716 2.3000000000000003-8.604285714285714 6.234285714285714-6.238571428571428 8.604285714285714-2.3000000000000003 8.605714285714285 2.3000000000000003 6.238571428571426 6.238571428571428 2.298571428571435 8.604285714285714z' })
                )
            );
        }
    }]);

    return FaCompass;
}(React.Component);

exports.default = FaCompass;
module.exports = exports['default'];