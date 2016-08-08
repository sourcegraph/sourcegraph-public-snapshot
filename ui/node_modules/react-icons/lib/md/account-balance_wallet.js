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

var MdAccountBalanceWallet = function (_React$Component) {
    _inherits(MdAccountBalanceWallet, _React$Component);

    function MdAccountBalanceWallet() {
        _classCallCheck(this, MdAccountBalanceWallet);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAccountBalanceWallet).apply(this, arguments));
    }

    _createClass(MdAccountBalanceWallet, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 22.5q1.0166666666666657 0 1.7583333333333329-0.7033333333333331t0.740000000000002-1.7966666666666669-0.7416666666666671-1.7966666666666669-1.7600000000000016-0.7033333333333331-1.7583333333333329 0.7033333333333331-0.7399999999999984 1.7966666666666669 0.7416666666666671 1.7966666666666669 1.7566666666666677 0.7033333333333331z m-6.640000000000001 4.140000000000001v-13.283333333333333h16.64v13.283333333333333h-16.64z m15 3.3599999999999994v1.6400000000000006q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657h-23.28333333333333q-1.405000000000002 0-2.3833333333333346-1.0166666666666657t-0.9749999999999996-2.34333333333333v-23.28333333333334q0-1.3266666666666653 0.9766666666666666-2.341666666666665t2.3833333333333346-1.0150000000000006h23.28333333333334q1.326666666666668 0 2.3416666666666686 1.0166666666666666t1.0149999999999935 2.3400000000000007v1.6433333333333326h-15q-1.4066666666666663 0-2.383333333333333 1.0133333333333336t-0.9766666666666666 2.3433333333333337v13.283333333333333q0 1.326666666666668 0.9766666666666666 2.3416666666666686t2.383333333333333 1.0133333333333319h15z' })
                )
            );
        }
    }]);

    return MdAccountBalanceWallet;
}(React.Component);

exports.default = MdAccountBalanceWallet;
module.exports = exports['default'];