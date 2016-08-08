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

var TiVendorMicrosoft = function (_React$Component) {
    _inherits(TiVendorMicrosoft, _React$Component);

    function TiVendorMicrosoft() {
        _classCallCheck(this, TiVendorMicrosoft);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiVendorMicrosoft).apply(this, arguments));
    }

    _createClass(TiVendorMicrosoft, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.7 20.8c0-0.5-0.3999999999999986-0.8000000000000007-0.9000000000000004-0.8000000000000007h-10c-0.5 0-0.7999999999999998 0.3000000000000007-0.7999999999999998 0.8000000000000007v8.400000000000002c0 0.5 0.2999999999999998 0.8000000000000007 0.7999999999999998 1l10 1.1000000000000014c0.4999999999999982 0 0.9000000000000004-0.3000000000000007 0.9000000000000004-0.6000000000000014v-9.899999999999999z m2.5-0.8000000000000007c-0.5 0-0.8999999999999986 0.3000000000000007-0.8999999999999986 0.8000000000000007v9.900000000000002c0 0.5 0.3999999999999986 0.8000000000000007 0.8999999999999986 1l15.000000000000004 1.6000000000000014c0.5 0 0.7999999999999972-0.29999999999999716 0.7999999999999972-0.6000000000000014v-11.700000000000003c0-0.5-0.29999999999999716-0.8000000000000007-0.7999999999999972-0.8000000000000007l-15-0.1999999999999993z m-2.5-12.2c0-0.5000000000000009-0.3999999999999986-0.8000000000000007-0.9000000000000004-0.6000000000000005l-10 1.1000000000000005c-0.5 0-0.7999999999999998 0.40000000000000036-0.7999999999999998 0.9000000000000004v8.299999999999999c0 0.5 0.2999999999999998 0.8000000000000007 0.7999999999999998 0.8000000000000007h10c0.4999999999999982 0 0.9000000000000004-0.3000000000000007 0.9000000000000004-0.8000000000000007v-9.7z m2.5-1.0000000000000009c-0.5 0-0.8999999999999986 0.5-0.8999999999999986 1v9.899999999999999c0 0.5 0.3999999999999986 0.8000000000000007 0.8999999999999986 0.8000000000000007h15.000000000000004c0.5 0 0.7999999999999972-0.3000000000000007 0.7999999999999972-0.8000000000000007v-11.7c0-0.5-0.29999999999999716-0.7999999999999998-0.7999999999999972-0.7000000000000002l-15 1.5z' })
                )
            );
        }
    }]);

    return TiVendorMicrosoft;
}(React.Component);

exports.default = TiVendorMicrosoft;
module.exports = exports['default'];