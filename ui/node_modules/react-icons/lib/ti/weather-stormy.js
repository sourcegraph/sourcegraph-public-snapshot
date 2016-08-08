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

var TiWeatherStormy = function (_React$Component) {
    _inherits(TiWeatherStormy, _React$Component);

    function TiWeatherStormy() {
        _classCallCheck(this, TiWeatherStormy);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiWeatherStormy).apply(this, arguments));
    }

    _createClass(TiWeatherStormy, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.333333333333336 30c-0.9216666666666669 0-1.6666666666666679-0.745000000000001-1.6666666666666679-1.6666666666666679s0.745000000000001-1.6666666666666679 1.6666666666666679-1.6666666666666679c2.7566666666666677 0 5-2.2433333333333323 5-5s-2.2433333333333323-5-5-5c-0.3966666666666683 0-0.826666666666668 0.07000000000000028-1.3550000000000004 0.216666666666665l-1.783333333333335 0.5033333333333339-0.3116666666666674-1.83c-0.543333333333333-3.2199999999999953-3.3000000000000007-5.556666666666663-6.549999999999997-5.556666666666663-3.6750000000000007 0-6.666666666666668 2.99-6.666666666666668 6.666666666666668 0 0.456666666666667 0.04499999999999993 0.908333333333335 0.13666666666666671 1.3433333333333337l0.43333333333333357 2.0666666666666664-2.3933333333333344-0.086666666666666c-1.6833333333333336 0.00999999999999801-3.1766666666666667 1.504999999999999-3.1766666666666667 3.3433333333333337s1.4933333333333332 3.333333333333332 3.333333333333333 3.333333333333332c0.9216666666666669 0 1.666666666666666 0.745000000000001 1.666666666666666 1.6666666666666679s-0.7449999999999992 1.6666666666666679-1.666666666666666 1.6666666666666679c-3.6750000000000007 0-6.666666666666667-2.9899999999999984-6.666666666666667-6.666666666666668 0-3.1000000000000014 2.128333333333333-5.716666666666669 5.003333333333333-6.456666666666667-0.003333333333332078-0.07000000000000028-0.003333333333332078-0.14000000000000057-0.003333333333332078-0.21000000000000085 0-5.516666666666666 4.4833333333333325-10 10.000000000000002-10 4.311666666666667 0 8.04 2.7300000000000004 9.416666666666668 6.691666666666666 4.871666666666666-0.40000000000000036 8.916666666666668 3.5199999999999996 8.916666666666668 8.308333333333334 0 4.595000000000002-3.7383333333333297 8.333333333333332-8.333333333333336 8.333333333333332z m-7.266666666666666-6.666666666666664l-7.5 6.75 5 2.4166666666666643-2.5 5.833333333333336 7.5-6.75-5-2.416666666666668z' })
                )
            );
        }
    }]);

    return TiWeatherStormy;
}(React.Component);

exports.default = TiWeatherStormy;
module.exports = exports['default'];