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

var TiPin = function (_React$Component) {
    _inherits(TiPin, _React$Component);

    function TiPin() {
        _classCallCheck(this, TiPin);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiPin).apply(this, arguments));
    }

    _createClass(TiPin, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.883333333333336 7.116666666666666c-0.6499999999999986-0.6500000000000004-1.7033333333333331-0.6533333333333333-2.3583333333333343-0.004999999999999893-0.173333333333332 0.17333333333333378-0.293333333333333 0.37833333333333297-0.375 0.5916666666666668-1.3866666666666667 2.8933333333333344-2.913333333333334 4.525-4.84 5.488333333333333-2.16 1.0666666666666664-4.643333333333334 1.8083333333333336-8.643333333333334 1.8083333333333336-0.21666666666666679 0-0.43333333333333357 0.041666666666666075-0.6366666666666667 0.12666666666666693-0.4083333333333332 0.16999999999999993-0.7333333333333343 0.4949999999999992-0.9000000000000004 0.9000000000000004-0.16999999999999993 0.408333333333335-0.16999999999999993 0.8666666666666671 0 1.2749999999999986 0.08333333333333393 0.206666666666667 0.20500000000000007 0.3916666666666657 0.35999999999999943 0.5450000000000017l5.404999999999999 5.405000000000001-7.561666666666666 10.081666666666667 10.083333333333334-7.561666666666667 5.403333333333332 5.403333333333332c0.15333333333333243 0.1566666666666663 0.3383333333333347 0.2766666666666673 0.543333333333333 0.36166666666666814 0.20333333333333314 0.08333333333333215 0.4200000000000017 0.129999999999999 0.6366666666666667 0.129999999999999s0.43333333333333357-0.045000000000001705 0.6366666666666667-0.129999999999999c0.408333333333335-0.1700000000000017 0.7333333333333343-0.49166666666666714 0.8999999999999986-0.8999999999999986 0.086666666666666-0.20333333333333314 0.129999999999999-0.42166666666666686 0.129999999999999-0.6366666666666667 0-4 0.7399999999999984-6.483333333333334 1.8049999999999997-8.610000000000003 0.961666666666666-1.9266666666666659 2.5933333333333337-3.453333333333333 5.488333333333337-4.84 0.21666666666666856-0.08333333333333215 0.4166666666666643-0.1999999999999993 0.5900000000000034-0.375 0.6499999999999986-0.6549999999999994 0.6450000000000031-1.708333333333334-0.006666666666667709-2.3566666666666656l-6.66-6.701666666666669z' })
                )
            );
        }
    }]);

    return TiPin;
}(React.Component);

exports.default = TiPin;
module.exports = exports['default'];