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

var MdReplay5 = function (_React$Component) {
    _inherits(MdReplay5, _React$Component);

    function MdReplay5() {
        _classCallCheck(this, MdReplay5);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdReplay5).apply(this, arguments));
    }

    _createClass(MdReplay5, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm19.688333333333336 22.656666666666666q-0.5466666666666669 0.23333333333333428-0.5466666666666669 0.3133333333333326l-0.1566666666666663 0.23333333333333428h-1.1699999999999982l0.39000000000000057-3.6700000000000017h3.9833333333333343v1.1716666666666669h-2.888333333333339l-0.158333333333335 1.4833333333333343q0.1566666666666663 0 0.1566666666666663-0.15500000000000114 0-0.07833333333333314 0.11666666666666714-0.11666666666666714t0.11666666666666714-0.11666666666666714h0.6233333333333348q0.625 0 0.8599999999999994 0.23333333333333428 0.07833333333333314 0.0799999999999983 0.3133333333333326 0.23666666666666814t0.3133333333333326 0.23333333333333428q0.7033333333333331 0.7049999999999983 0.7033333333333331 1.8000000000000007 0 0.6999999999999993-0.1566666666666663 0.8583333333333343-0.07833333333333314 0.07833333333333314-0.23333333333333428 0.39000000000000057t-0.31666666666666643 0.46999999999999886-0.3500000000000014 0.27333333333333343-0.27333333333333343 0.19500000000000028q-0.1566666666666663 0.1566666666666663-1.0166666666666657 0.1566666666666663-0.6999999999999993 0-0.8583333333333343-0.1566666666666663-0.07833333333333314-0.07833333333333314-0.3500000000000014-0.1566666666666663t-0.43333333333333357-0.158333333333335q-0.6999999999999993-0.39000000000000057-0.6999999999999993-1.4833333333333343h1.326666666666668q0 0.7800000000000011 1.0166666666666657 0.7800000000000011 0.3133333333333326 0 0.466666666666665-0.1566666666666663l0.3916666666666657-0.31666666666666643q0.1566666666666663-0.3116666666666674 0.1566666666666663-0.466666666666665v-1.0166666666666657l-0.158333333333335-0.3116666666666674-0.39000000000000057-0.39000000000000057q-0.3116666666666674-0.158333333333335-0.466666666666665-0.158333333333335h-0.3133333333333326z m0.31166666666666387-14.296666666666665q5.546666666666667 0 9.453333333333333 3.9066666666666663t3.9066666666666663 9.374999999999998q0 5.546666666666667-3.9450000000000003 9.453333333333333t-9.415 3.905000000000001-9.411666666666667-3.905000000000001-3.9450000000000003-9.453333333333333h3.360000000000001q0 4.140000000000001 2.966666666666667 7.07t7.033333333333333 2.9299999999999997 7.033333333333335-2.9299999999999997 2.963333333333331-7.07-2.9666666666666686-7.07-7.033333333333331-2.929999999999998v6.716666666666667l-8.356666666666667-8.358333333333334 8.360000000000001-8.36v6.716666666666667z' })
                )
            );
        }
    }]);

    return MdReplay5;
}(React.Component);

exports.default = MdReplay5;
module.exports = exports['default'];