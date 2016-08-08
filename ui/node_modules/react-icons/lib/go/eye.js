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

var GoEye = function (_React$Component) {
    _inherits(GoEye, _React$Component);

    function GoEye() {
        _classCallCheck(this, GoEye);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoEye).apply(this, arguments));
    }

    _createClass(GoEye, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 5c-7.5 0-16.25 5-20 15 3.75 7.5 11.25 12.5 20 12.5s16.25-5 20-12.5c-3.75-10-12.5-15-20-15z m0 25c-7.5 0-13.75-5-15-10 1.25-5 7.5-10 15-10s13.75 5 15 10c-1.25 5-7.5 10-15 10z m0-17.5c-0.78125 0-1.4849999999999994 0.15625-2.1875 0.34999999999999964 1.2875000000000014 0.5875000000000004 2.1875 1.875 2.1875 3.4000000000000004 0 2.0700000000000003-1.6799999999999997 3.75-3.75 3.75-1.5250000000000004 0-2.8125-0.8999999999999986-3.4000000000000004-2.1875-0.19374999999999964 0.7037499999999994-0.34999999999999964 1.40625-0.34999999999999964 2.1875 0 4.141249999999999 3.3599999999999994 7.5 7.5 7.5s7.5-3.3599999999999994 7.5-7.5-3.3599999999999994-7.5-7.5-7.5z' })
                )
            );
        }
    }]);

    return GoEye;
}(React.Component);

exports.default = GoEye;
module.exports = exports['default'];