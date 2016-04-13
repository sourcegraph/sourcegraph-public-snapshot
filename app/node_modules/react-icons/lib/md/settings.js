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

var MdSettings = function (_React$Component) {
    _inherits(MdSettings, _React$Component);

    function MdSettings() {
        _classCallCheck(this, MdSettings);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSettings).apply(this, arguments));
    }

    _createClass(MdSettings, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 25.86q2.421666666666667 0 4.140000000000001-1.7166666666666686t1.716666666666665-4.143333333333331-1.716666666666665-4.138333333333334-4.140000000000001-1.716666666666665-4.140000000000001 1.7166666666666668-1.716666666666665 4.138333333333332 1.7166666666666668 4.141666666666666 4.139999999999999 1.7166666666666686z m12.421666666666667-4.219999999999999l3.5166666666666657 2.7333333333333343q0.5450000000000017 0.3916666666666657 0.15500000000000114 1.0949999999999989l-3.3599999999999994 5.783333333333335q-0.3133333333333326 0.5466666666666669-1.0166666666666657 0.3133333333333326l-4.138333333333335-1.6400000000000006q-1.6400000000000006 1.1716666666666669-2.8133333333333326 1.6400000000000006l-0.6266666666666652 4.3749999999999964q-0.1566666666666663 0.7033333333333331-0.783333333333335 0.7033333333333331h-6.716666666666669q-0.625 0-0.7833333333333332-0.7033333333333331l-0.6233333333333331-4.375q-1.4833333333333343-0.625-2.8133333333333326-1.6400000000000006l-4.140000000000001 1.6400000000000006q-0.7033333333333331 0.23333333333333428-1.0166666666666666-0.3133333333333326l-3.358333333333334-5.783333333333335q-0.3900000000000001-0.6999999999999993 0.15666666666666673-1.091666666666665l3.5166666666666666-2.7333333333333343q-0.08000000000000007-0.5500000000000007-0.08000000000000007-1.6416666666666657t0.0766666666666671-1.6400000000000006l-3.5166666666666666-2.7333333333333343q-0.5449999999999999-0.3916666666666675-0.1549999999999998-1.0950000000000006l3.3650000000000015-5.78333333333333q0.31166666666666654-0.5466666666666669 1.0133333333333336-0.3133333333333326l4.140000000000001 1.6400000000000006q1.6400000000000006-1.1716666666666669 2.8133333333333326-1.6400000000000006l0.625-4.375q0.1566666666666663-0.7033333333333331 0.783333333333335-0.7033333333333331h6.716666666666669q0.625 0 0.783333333333335 0.7033333333333331l0.6233333333333348 4.375q1.4833333333333343 0.625 2.8116666666666674 1.6400000000000006l4.140000000000001-1.6400000000000006q0.7049999999999983-0.2333333333333325 1.0166666666666657 0.3133333333333326l3.3599999999999994 5.783333333333335q0.39000000000000057 0.6999999999999993-0.1566666666666663 1.0916666666666668l-3.5166666666666657 2.7333333333333325q0.0799999999999983 0.5500000000000007 0.0799999999999983 1.6416666666666657t-0.07833333333333314 1.6400000000000006z' })
                )
            );
        }
    }]);

    return MdSettings;
}(React.Component);

exports.default = MdSettings;
module.exports = exports['default'];