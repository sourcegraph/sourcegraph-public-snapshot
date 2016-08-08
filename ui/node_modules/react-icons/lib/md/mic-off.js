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

var MdMicOff = function (_React$Component) {
    _inherits(MdMicOff, _React$Component);

    function MdMicOff() {
        _classCallCheck(this, MdMicOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMicOff).apply(this, arguments));
    }

    _createClass(MdMicOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm7.11 5l27.89 27.890000000000008-2.1099999999999923 2.1099999999999923-6.953333333333333-6.953333333333333q-1.7166666666666686 1.0933333333333337-4.296666666666667 1.4833333333333343v5.469999999999999h-3.2833333333333314v-5.466666666666669q-4.138333333333334-0.6266666666666652-7.066666666666666-3.789999999999999t-2.9333333333333336-7.383333333333333h2.8166666666666664q0 3.671666666666667 2.616666666666667 6.055t6.20999999999999 2.384999999999998q1.9533333333333331 0 3.828333333333333-0.8616666666666681l-2.7333333333333343-2.7333333333333343q-0.6266666666666652 0.15500000000000114-1.0949999999999989 0.15500000000000114-2.0333333333333314 0-3.5166666666666657-1.4833333333333343t-1.4833333333333343-3.5183333333333273v-1.25l-10-10z m17.89 13.593333333333334l-10-9.921666666666667v-0.3133333333333326q0-2.033333333333333 1.4833333333333343-3.5166666666666666t3.5166666666666657-1.4833333333333334 3.5166666666666657 1.4833333333333334 1.4833333333333343 3.5166666666666666v10.233333333333334z m6.640000000000001-0.23333333333333428q0 2.8900000000000006-1.4833333333333343 5.466666666666669l-2.0333333333333314-2.1083333333333343q0.7033333333333331-1.5616666666666674 0.7033333333333331-3.3583333333333343h2.8133333333333326z' })
                )
            );
        }
    }]);

    return MdMicOff;
}(React.Component);

exports.default = MdMicOff;
module.exports = exports['default'];