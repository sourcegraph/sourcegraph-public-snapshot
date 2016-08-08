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

var TiDeleteOutline = function (_React$Component) {
    _inherits(TiDeleteOutline, _React$Component);

    function TiDeleteOutline() {
        _classCallCheck(this, TiDeleteOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiDeleteOutline).apply(this, arguments));
    }

    _createClass(TiDeleteOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 5c-8.271666666666667 0-15 6.73-15 15s6.7283333333333335 15 15 15 15-6.73 15-15-6.728333333333332-15-15-15z m0 26.666666666666668c-6.433333333333334 0-11.666666666666668-5.233333333333334-11.666666666666668-11.666666666666668s5.233333333333334-11.666666666666668 11.666666666666668-11.666666666666668 11.666666666666668 5.233333333333334 11.666666666666668 11.666666666666668-5.233333333333334 11.666666666666668-11.666666666666668 11.666666666666668z m1.1783333333333346-11.666666666666668l4.41-4.41c0.32333333333333414-0.32333333333333414 0.32333333333333414-0.8533333333333335 0-1.1783333333333328-0.3249999999999993-0.32333333333333414-0.8550000000000004-0.32333333333333414-1.1783333333333346 0l-4.41 4.409999999999998-4.41-4.4116666666666635c-0.3249999999999993-0.32333333333333414-0.8550000000000004-0.32333333333333414-1.1783333333333328 0-0.3249999999999993 0.3249999999999993-0.3249999999999993 0.8550000000000004 0 1.1783333333333328l4.409999999999998 4.411666666666665-4.41 4.41c-0.3249999999999993 0.3249999999999993-0.3249999999999993 0.8550000000000004 0 1.1783333333333346 0.16166666666666707 0.163333333333334 0.375 0.245000000000001 0.5883333333333329 0.245000000000001s0.42666666666666586-0.08333333333333215 0.5899999999999999-0.2433333333333323l4.410000000000002-4.411666666666669 4.41 4.41c0.163333333333334 0.163333333333334 0.37666666666666515 0.245000000000001 0.5899999999999999 0.245000000000001s0.42666666666666586-0.08333333333333215 0.5899999999999999-0.2433333333333323c0.32333333333333414-0.32333333333333414 0.32333333333333414-0.8533333333333317 0-1.1783333333333346l-4.411666666666665-4.411666666666669z' })
                )
            );
        }
    }]);

    return TiDeleteOutline;
}(React.Component);

exports.default = TiDeleteOutline;
module.exports = exports['default'];