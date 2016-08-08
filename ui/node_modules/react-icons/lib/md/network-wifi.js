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

var MdNetworkWifi = function (_React$Component) {
    _inherits(MdNetworkWifi, _React$Component);

    function MdNetworkWifi() {
        _classCallCheck(this, MdNetworkWifi);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdNetworkWifi).apply(this, arguments));
    }

    _createClass(MdNetworkWifi, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm34.06333333333333 18.203333333333337l0.07833333333333314 0.07833333333333314-14.141666666666666 17.50166666666666-14.138333333333335-17.5 0.07833333333333314-0.0799999999999983-5.313333333333331-6.563333333333333q9.453333333333333-6.640000000000001 19.373333333333335-6.640000000000001t19.376666666666665 6.640000000000001z' })
                )
            );
        }
    }]);

    return MdNetworkWifi;
}(React.Component);

exports.default = MdNetworkWifi;
module.exports = exports['default'];