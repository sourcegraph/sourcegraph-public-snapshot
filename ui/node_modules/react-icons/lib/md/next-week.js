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

var MdNextWeek = function (_React$Component) {
    _inherits(MdNextWeek, _React$Component);

    function MdNextWeek() {
        _classCallCheck(this, MdNextWeek);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdNextWeek).apply(this, arguments));
    }

    _createClass(MdNextWeek, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.36 30.86l6.640000000000001-6.716666666666665-6.640000000000001-6.643333333333334-1.7166666666666686 1.6416666666666657 5 5-5 5z m-1.7199999999999989-22.5v3.283333333333333h6.716666666666669v-3.283333333333333h-6.716666666666669z m6.719999999999999-3.3599999999999994q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.383333333333333v3.283333333333333h6.716666666666669q1.3299999999999983 0 2.306666666666665 1.0133333333333336t0.9766666666666737 2.3433333333333337v18.36q0 1.3283333333333331-0.9766666666666666 2.3049999999999997t-2.306666666666665 0.9750000000000014h-26.713333333333342q-1.330000000000001 0-2.3066666666666675-0.9766666666666666t-0.9766666666666666-2.306666666666665v-18.35666666666667q0-1.3283333333333331 0.9766666666666666-2.3433333333333337t2.3050000000000006-1.0166666666666657h6.716666666666667v-3.2783333333333324q0-1.3283333333333331 0.9783333333333335-2.3433333333333337t2.304999999999998-1.0183333333333344h6.716666666666669z' })
                )
            );
        }
    }]);

    return MdNextWeek;
}(React.Component);

exports.default = MdNextWeek;
module.exports = exports['default'];