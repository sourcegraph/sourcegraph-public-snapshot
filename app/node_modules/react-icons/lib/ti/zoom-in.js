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

var TiZoomIn = function (_React$Component) {
    _inherits(TiZoomIn, _React$Component);

    function TiZoomIn() {
        _classCallCheck(this, TiZoomIn);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiZoomIn).apply(this, arguments));
    }

    _createClass(TiZoomIn, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.666666666666668 18.333333333333336h-3.333333333333332v-3.3333333333333357c0-0.46000000000000085-0.37333333333333485-0.8333333333333339-0.8333333333333321-0.8333333333333339s-0.8333333333333321 0.3733333333333331-0.8333333333333321 0.8333333333333339v3.333333333333332h-3.3333333333333375c-0.46000000000000085 0-0.8333333333333339 0.37333333333333485-0.8333333333333339 0.8333333333333321s0.3733333333333331 0.8333333333333321 0.8333333333333339 0.8333333333333321h3.333333333333334v3.333333333333332c0 0.46000000000000085 0.37333333333333485 0.8333333333333321 0.8333333333333321 0.8333333333333321s0.8333333333333321-0.37333333333333485 0.8333333333333321-0.8333333333333321v-3.333333333333332h3.333333333333332c0.46000000000000085 0 0.8333333333333321-0.37333333333333485 0.8333333333333321-0.8333333333333321s-0.37333333333333485-0.8333333333333321-0.8333333333333321-0.8333333333333321z m9.053333333333331 6.616666666666667l-2.1416666666666657-2.1449999999999996c0.37833333333333385-1.1466666666666683 0.5899999999999999-2.366666666666667 0.5899999999999999-3.638333333333332 0-6.433333333333334-5.233333333333334-11.666666666666668-11.666666666666668-11.666666666666668s-11.666666666666668 5.233333333333334-11.666666666666668 11.666666666666668 5.236666666666666 11.666666666666668 11.666666666666668 11.666666666666668c1.2733333333333334 0 2.495000000000001-0.211666666666666 3.6416666666666657-0.5899999999999999l3.871666666666666 3.866666666666667 0.10666666666666558 0.09166666666666856c1 0.8449999999999989 2.246666666666666 1.30833333333333 3.5066666666666677 1.30833333333333 2.876666666666665 0 5.216666666666669-2.3400000000000034 5.216666666666669-5.216666666666669 0-1.3999999999999986-0.5466666666666669-2.7166666666666686-1.5399999999999991-3.6999999999999993l-1.5833333333333321-1.6416666666666657z m-21.55333333333333-5.783333333333335c0-4.595000000000001 3.7383333333333333-8.333333333333334 8.333333333333332-8.333333333333334s8.333333333333336 3.7383333333333333 8.333333333333336 8.333333333333334-3.7383333333333333 8.333333333333332-8.333333333333336 8.333333333333332-8.333333333333334-3.7383333333333333-8.333333333333334-8.333333333333336z' })
                )
            );
        }
    }]);

    return TiZoomIn;
}(React.Component);

exports.default = TiZoomIn;
module.exports = exports['default'];