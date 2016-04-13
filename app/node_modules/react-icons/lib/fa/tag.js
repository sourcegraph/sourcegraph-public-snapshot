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

var FaTag = function (_React$Component) {
    _inherits(FaTag, _React$Component);

    function FaTag() {
        _classCallCheck(this, FaTag);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTag).apply(this, arguments));
    }

    _createClass(FaTag, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm12.857142857142858 10q0-1.1828571428571433-0.8371428571428563-2.0199999999999996t-2.0200000000000014-0.8371428571428572-2.0199999999999996 0.8371428571428572-0.8371428571428572 2.0199999999999996 0.8371428571428572 2.0199999999999996 2.0199999999999996 0.8371428571428581 2.0199999999999996-0.8371428571428563 0.8371428571428581-2.0200000000000014z m23.81714285714286 12.857142857142858q0 1.1828571428571415-0.8257142857142838 2.008571428571429l-10.96 10.982857142857146q-0.8714285714285701 0.8257142857142838-2.0314285714285703 0.8257142857142838-1.1828571428571415 0-2.008571428571429-0.8257142857142838l-15.960000000000004-15.982857142857146q-0.8485714285714288-0.8257142857142874-1.44-2.2542857142857144t-0.5914285714285707-2.611428571428572v-9.285714285714285q0-1.160000000000001 0.8485714285714283-2.0085714285714293t2.008571428571429-0.8485714285714288h9.285714285714285q1.1828571428571415 0 2.611428571428572 0.5914285714285712t2.277142857142856 1.4399999999999995l15.960000000000004 15.937142857142856q0.8257142857142838 0.8714285714285701 0.8257142857142838 2.0314285714285703z' })
                )
            );
        }
    }]);

    return FaTag;
}(React.Component);

exports.default = FaTag;
module.exports = exports['default'];