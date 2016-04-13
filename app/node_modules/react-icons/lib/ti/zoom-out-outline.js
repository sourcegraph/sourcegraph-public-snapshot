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

var TiZoomOutOutline = function (_React$Component) {
    _inherits(TiZoomOutOutline, _React$Component);

    function TiZoomOutOutline() {
        _classCallCheck(this, TiZoomOutOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiZoomOutOutline).apply(this, arguments));
    }

    _createClass(TiZoomOutOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.666666666666668 20h-8.333333333333334c-0.4583333333333339 0-0.8333333333333339-0.375-0.8333333333333339-0.8333333333333321s0.375-0.8333333333333321 0.8333333333333339-0.8333333333333321h8.333333333333334c0.45833333333333215 0 0.8333333333333321 0.375 0.8333333333333321 0.8333333333333321s-0.375 0.8333333333333321-0.8333333333333321 0.8333333333333321z m10.63333333333333 6.593333333333334l-3.740000000000002-3.8049999999999997c0.379999999999999-1.1449999999999996 0.6066666666666656-2.3533333333333317 0.6066666666666656-3.621666666666666 0-6.433333333333334-5.233333333333334-11.666666666666668-11.666666666666668-11.666666666666668s-11.666666666666668 5.233333333333334-11.666666666666668 11.666666666666668 5.233333333333334 11.666666666666668 11.666666666666668 11.666666666666668c1.2666666666666657 0 2.4800000000000004-0.22833333333333172 3.621666666666666-0.6066666666666656l3.995000000000001 3.97666666666667c1 0.8433333333333337 2.246666666666666 1.3049999999999997 3.5066666666666677 1.3049999999999997 2.878333333333334 0 5.216666666666669-2.3400000000000034 5.216666666666669-5.216666666666669 0-1.3999999999999986-0.5450000000000017-2.7166666666666686-1.538333333333334-3.6999999999999993z m-6.5-1.8500000000000014l4.153333333333332 4.216666666666669c0.341666666666665 0.33999999999999986 0.5533333333333346 0.8116666666666674 0.5533333333333346 1.3299999999999983 0 1.0416666666666679-0.8449999999999989 1.8833333333333329-1.8833333333333329 1.8833333333333329-0.5216666666666683 0-0.9916666666666671-0.21000000000000085-1.3616666666666681-0.5199999999999996l-4.186666666666667-4.183333333333334c1.076666666666668-0.7283333333333317 2-1.6533333333333324 2.7250000000000014-2.7266666666666666z m-16.63333333333333-5.576666666666664c0-4.595000000000001 3.7383333333333333-8.333333333333334 8.333333333333332-8.333333333333334s8.333333333333336 3.7383333333333333 8.333333333333336 8.333333333333334-3.7383333333333333 8.333333333333332-8.333333333333336 8.333333333333332-8.333333333333334-3.7383333333333333-8.333333333333334-8.333333333333336z' })
                )
            );
        }
    }]);

    return TiZoomOutOutline;
}(React.Component);

exports.default = TiZoomOutOutline;
module.exports = exports['default'];