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

var TiStopwatch = function (_React$Component) {
    _inherits(TiStopwatch, _React$Component);

    function TiStopwatch() {
        _classCallCheck(this, TiStopwatch);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiStopwatch).apply(this, arguments));
    }

    _createClass(TiStopwatch, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.35666666666667 14.836666666666666c0.173333333333332-0.08000000000000007 0.3433333333333337-0.17999999999999972 0.48833333333333684-0.3249999999999993l0.8333333333333357-0.8333333333333339c0.6499999999999986-0.6500000000000004 0.6499999999999986-1.705 0-2.3566666666666674s-1.7049999999999983-0.6500000000000004-2.3566666666666656 0l-0.8333333333333321 0.8333333333333339c-0.08333333333333215 0.08333333333333393-0.12833333333333385 0.19166666666666643-0.19166666666666643 0.288333333333334-2.3116666666666674-2.1866666666666674-5.313333333333333-3.6500000000000004-8.650000000000002-4.0166666666666675 0.003333333333326749-0.033333333333333215 0.019999999999992468-0.05999999999999872 0.019999999999992468-0.09333333333333194v-1.666666666666667h1.6666666666666679c0.9166666666666679 0 1.6666666666666679-0.75 1.6666666666666679-1.666666666666667s-0.75-1.666666666666667-1.6666666666666679-1.666666666666667h-6.666666666666668c-0.9166666666666661 0-1.666666666666666 0.75-1.666666666666666 1.666666666666667s0.75 1.666666666666667 1.666666666666666 1.666666666666667h1.6666666666666679v1.666666666666667c0 0.033333333333333215 0.01666666666666572 0.061666666666667425 0.019999999999999574 0.09500000000000064-7.510000000000002 0.8216666666666654-13.353333333333335 7.178333333333335-13.353333333333335 14.905000000000001 0 8.283333333333331 6.716666666666669 15 15 15s15-6.716666666666669 15-15c0-3.1566666666666663-0.9799999999999969-6.079999999999998-2.643333333333331-8.496666666666666z m-12.35666666666667 20.163333333333334c-6.433333333333334 0-11.666666666666668-5.233333333333334-11.666666666666668-11.666666666666668s5.233333333333334-11.666666666666668 11.666666666666668-11.666666666666668 11.666666666666668 5.233333333333334 11.666666666666668 11.666666666666668-5.233333333333334 11.666666666666668-11.666666666666668 11.666666666666668z m1.6666666666666679-13.333333333333332v-3.333333333333332c0-0.9166666666666679-0.75-1.6666666666666679-1.6666666666666679-1.6666666666666679s-1.6666666666666679 0.75-1.6666666666666679 1.6666666666666679v5c0 0.9166666666666679 0.75 1.6666666666666679 1.6666666666666679 1.6666666666666679h5c0.9166666666666679 0 1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679s-0.75-1.6666666666666679-1.6666666666666679-1.6666666666666679h-3.333333333333332z m-1.6666666666666679-8.333333333333334c-5.52 0-10 4.479999999999999-10 10.000000000000002s4.48 10 10 10 10-4.48 10-10-4.48-10-10-10z m0 18.333333333333336c-4.595000000000001 0-8.333333333333334-3.7383333333333333-8.333333333333334-8.333333333333336s3.7383333333333333-8.333333333333334 8.333333333333334-8.333333333333334 8.333333333333336 3.738333333333335 8.333333333333336 8.333333333333334-3.7383333333333333 8.333333333333336-8.333333333333336 8.333333333333336z' })
                )
            );
        }
    }]);

    return TiStopwatch;
}(React.Component);

exports.default = TiStopwatch;
module.exports = exports['default'];