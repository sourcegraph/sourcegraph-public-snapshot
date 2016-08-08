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

var GoDeviceMobile = function (_React$Component) {
    _inherits(GoDeviceMobile, _React$Component);

    function GoDeviceMobile() {
        _classCallCheck(this, GoDeviceMobile);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoDeviceMobile).apply(this, arguments));
    }

    _createClass(GoDeviceMobile, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 0h-20c-1.3787500000000001 0-2.5 1.12-2.5 2.5v35c0 1.3812500000000014 1.120000000000001 2.5 2.5 2.5h20c1.3812500000000014 0 2.5-1.1187499999999986 2.5-2.5v-35c0-1.37875-1.1187499999999986-2.5-2.5-2.5z m-11.25 2.5h2.5c0.6875 0 1.25 0.56 1.25 1.25s-0.5625 1.25-1.25 1.25h-2.5c-0.6900000000000013 0-1.25-0.5600000000000005-1.25-1.25s0.5599999999999987-1.25 1.25-1.25z m2.5 35h-2.5c-0.6900000000000013 0-1.25-0.5625-1.25-1.25s0.5599999999999987-1.25 1.25-1.25h2.5c0.6875 0 1.25 0.5625 1.25 1.25s-0.5625 1.25-1.25 1.25z m8.75-5h-20v-25h20v25z' })
                )
            );
        }
    }]);

    return GoDeviceMobile;
}(React.Component);

exports.default = GoDeviceMobile;
module.exports = exports['default'];