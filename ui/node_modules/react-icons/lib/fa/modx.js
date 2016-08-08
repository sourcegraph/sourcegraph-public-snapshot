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

var FaModx = function (_React$Component) {
    _inherits(FaModx, _React$Component);

    function FaModx() {
        _classCallCheck(this, FaModx);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaModx).apply(this, arguments));
    }

    _createClass(FaModx, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.852857142857143 15.825714285714286l-13.705714285714283-8.614285714285716 2.0528571428571425-3.3714285714285714h19.085714285714282z m-22.814285714285717 5.914285714285713l-4.105714285714286-2.588571428571427v-19.15142857142857l26.40571428571429 16.585714285714285z m22.747142857142858-3.014285714285716l3.2814285714285703 2.12142857142857v19.152857142857147l-11.875714285714288-7.478571428571428z m-0.8257142857142838-0.4671428571428571l-11.16 17.902857142857147h-19.085714285714285l7.948571428571427-12.745714285714286z' })
                )
            );
        }
    }]);

    return FaModx;
}(React.Component);

exports.default = FaModx;
module.exports = exports['default'];