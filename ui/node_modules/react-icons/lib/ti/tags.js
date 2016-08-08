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

var TiTags = function (_React$Component) {
    _inherits(TiTags, _React$Component);

    function TiTags() {
        _classCallCheck(this, TiTags);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiTags).apply(this, arguments));
    }

    _createClass(TiTags, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.70333333333333 15.99l-10.774999999999999-10.89c-2.216666666666665-2.2166666666666663-5.145000000000003-3.4333333333333327-8.261666666666667-3.4333333333333327s-6.046666666666667 1.2166666666666668-8.25 3.416666666666666c-2.360000000000001 2.360000000000001-3.545000000000001 5.5966666666666685-3.378333333333334 8.86-1.1050000000000009 1.8066666666666684-1.7050000000000005 3.8816666666666713-1.7050000000000005 6.056666666666668 0 3.116666666666667 1.2166666666666672 6.045000000000002 3.4166666666666665 8.25l5.09 4.971666666666664 5.803333333333335 5.799999999999997c0.6499999999999986 0.653333333333336 1.5033333333333339 0.9783333333333317 2.3566666666666656 0.9783333333333317s1.7049999999999983-0.32500000000000284 2.3566666666666656-0.9766666666666666l11.666666666666668-11.666666666666668c1.2966666666666669-1.2966666666666669 1.3033333333333346-3.3966666666666647 0.013333333333335418-4.699999999999999l-0.15500000000000114-0.1566666666666663 1.8066666666666649-1.8099999999999987c1.2966666666666669-1.2966666666666669 1.3033333333333346-3.3966666666666647 0.013333333333335418-4.699999999999999z m-15.703333333333333 20.67666666666667l-5.830000000000002-5.828333333333333-5.061666666666667-4.946666666666665c-3.255000000000001-3.254999999999999-3.255000000000001-8.533333333333331 0-11.783333333333335 1.6250000000000036-1.6283333333333374 3.7583333333333364-2.44166666666667 5.891666666666669-2.44166666666667s4.266666666666666 0.8133333333333326 5.891666666666666 2.4416666666666664l10.775000000000002 10.891666666666666-11.666666666666668 11.666666666666671z m3.2616666666666667-24.9c-2.216666666666665-2.216666666666672-5.145-3.433333333333339-8.261666666666667-3.433333333333339-2.126666666666667 0-4.161666666666667 0.5783333333333331-5.941666666666666 1.6366666666666667 0.40000000000000036-0.9166666666666661 0.9666666666666668-1.7783333333333342 1.7166666666666668-2.5300000000000002 1.625-1.626666666666667 3.7583333333333346-2.4400000000000004 5.8916666666666675-2.4400000000000004s4.266666666666666 0.8133333333333335 5.891666666666669 2.4416666666666664l10.774999999999999 10.89166666666667-1.7966666666666669 1.7966666666666669-8.274999999999999-8.366666666666667z m-8.261666666666667 5.733333333333327c1.3833333333333329 0 2.5 1.1166666666666671 2.5 2.5s-1.1166666666666671 2.5-2.5 2.5-2.5-1.1166666666666671-2.5-2.5 1.1166666666666671-2.5 2.5-2.5z m0-1.666666666666666c-2.296666666666667 0-4.166666666666668 1.8666666666666654-4.166666666666668 4.166666666666666s1.870000000000001 4.166666666666668 4.166666666666668 4.166666666666668 4.166666666666668-1.870000000000001 4.166666666666668-4.166666666666668c0-2.3000000000000007-1.870000000000001-4.166666666666668-4.166666666666668-4.166666666666668z' })
                )
            );
        }
    }]);

    return TiTags;
}(React.Component);

exports.default = TiTags;
module.exports = exports['default'];