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

var MdKeyboardHide = function (_React$Component) {
    _inherits(MdKeyboardHide, _React$Component);

    function MdKeyboardHide() {
        _classCallCheck(this, MdKeyboardHide);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdKeyboardHide).apply(this, arguments));
    }

    _createClass(MdKeyboardHide, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 38.36l-6.640000000000001-6.716666666666669h13.283333333333331z m11.64-25v-3.3599999999999994h-3.2833333333333314v3.3599999999999994h3.2833333333333314z m0 5v-3.3599999999999994h-3.2833333333333314v3.3599999999999994h3.2833333333333314z m-5-5v-3.3599999999999994h-3.2833333333333314v3.3599999999999994h3.2833333333333314z m0 5v-3.3599999999999994h-3.2833333333333314v3.3599999999999994h3.2833333333333314z m0 6.640000000000001v-3.3599999999999994h-13.283333333333333v3.3599999999999994h13.283333333333333z m-15-11.64v-3.3599999999999994h-3.283333333333333v3.3599999999999994h3.283333333333333z m0 5v-3.3599999999999994h-3.283333333333333v3.3599999999999994h3.283333333333333z m1.7200000000000006-3.3599999999999994v3.3599999999999994h3.283333333333333v-3.3599999999999994h-3.283333333333333z m0-5v3.3599999999999994h3.283333333333333v-3.3599999999999994h-3.283333333333333z m4.999999999999998 5v3.3599999999999994h3.2833333333333314v-3.3599999999999994h-3.2833333333333314z m0-5v3.3599999999999994h3.2833333333333314v-3.3599999999999994h-3.2833333333333314z m15-5q1.3283333333333331 0 2.3049999999999997 1.0166666666666666t0.9750000000000014 2.3416666666666677v16.641666666666666q0 1.326666666666668-0.9766666666666666 2.3416666666666686t-2.306666666666665 1.0166666666666657h-26.713333333333335q-1.330000000000001 0-2.3066666666666675-1.0166666666666657t-0.9766666666666666-2.3416666666666686v-16.64333333333333q0-1.328333333333335 0.9766666666666666-2.3433333333333355t2.3049999999999997-1.0133333333333336h26.71666666666667z' })
                )
            );
        }
    }]);

    return MdKeyboardHide;
}(React.Component);

exports.default = MdKeyboardHide;
module.exports = exports['default'];