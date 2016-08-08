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

var TiFlash = function (_React$Component) {
    _inherits(TiFlash, _React$Component);

    function TiFlash() {
        _classCallCheck(this, TiFlash);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiFlash).apply(this, arguments));
    }

    _createClass(TiFlash, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.169999999999998 20.055l-7.066666666666666-4.096666666666668 3.5616666666666674-8.55c0.10833333333333428-0.22499999999999964 0.1700000000000017-0.47666666666666657 0.1700000000000017-0.7416666666666671 0-0.9199999999999999-0.7416666666666671-1.666666666666667-1.6616666666666653-1.666666666666667-0.41666666666666785 0.006666666666666821-0.7616666666666667 0.13833333333333364-1.0366666666666653 0.35666666666666647-0.054999999999999716 0.043333333333333-0.08833333333333471 0.07000000000000028-0.11666666666666714 0.09999999999999964l-12.5 11.833333333333336c-0.38333333333333286 0.36166666666666814-0.5700000000000003 0.8833333333333329-0.5099999999999998 1.4033333333333324s0.36666666666666714 0.9833333333333343 0.8166666666666664 1.25l7.071666666666667 4.100000000000001-3.6050000000000004 8.650000000000002c-0.3049999999999997 0.7266666666666666-0.05666666666666664 1.56666666666667 0.5899999999999999 2.0133333333333354 0.28999999999999915 0.19666666666666544 0.6216666666666661 0.29333333333333655 0.9499999999999993 0.29333333333333655 0.413333333333334 0 0.826666666666668-0.15500000000000114 1.1466666666666683-0.45666666666666345l12.5-11.836666666666666c0.38333333333333286-0.36166666666666814 0.5700000000000003-0.8833333333333329 0.5100000000000016-1.4033333333333324-0.061666666666667425-0.5216666666666683-0.36666666666666714-0.9833333333333343-0.8200000000000003-1.25z' })
                )
            );
        }
    }]);

    return TiFlash;
}(React.Component);

exports.default = TiFlash;
module.exports = exports['default'];