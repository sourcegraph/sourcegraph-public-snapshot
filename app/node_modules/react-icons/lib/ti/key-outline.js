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

var TiKeyOutline = function (_React$Component) {
    _inherits(TiKeyOutline, _React$Component);

    function TiKeyOutline() {
        _classCallCheck(this, TiKeyOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiKeyOutline).apply(this, arguments));
    }

    _createClass(TiKeyOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.666666666666668 35h-10v-7.356666666666666l6.305-6.305c-0.3116666666666674-0.9783333333333353-0.47166666666666757-1.9883333333333333-0.47166666666666757-3.004999999999999 0-5.516666666666666 4.483333333333334-10 10-10s10 4.483333333333334 10 10-4.483333333333334 10-10 10h-2.5v3.333333333333332h-3.333333333333332v3.333333333333332z m-6.666666666666668-3.333333333333332h3.333333333333334v-3.333333333333332h3.333333333333334v-3.333333333333332h5.833333333333332c3.676666666666666 0 6.666666666666668-2.9899999999999984 6.666666666666668-6.666666666666668s-2.9899999999999984-6.666666666666668-6.666666666666668-6.666666666666668-6.666666666666668 2.99-6.666666666666668 6.666666666666668c0 0.9333333333333336 0.1999999999999993 1.8500000000000014 0.6000000000000014 2.7333333333333343l0.4733333333333327 1.0500000000000007-6.906666666666666 6.906666666666666v2.643333333333331z m12.5-15.003333333333334c0.9166666666666679 0 1.6666666666666679 0.75 1.6666666666666679 1.6700000000000017s-0.75 1.6666666666666679-1.6666666666666679 1.6666666666666679-1.6666666666666679-0.7466666666666661-1.6666666666666679-1.6666666666666679c0-0.9216666666666669 0.75-1.6700000000000017 1.6666666666666679-1.6700000000000017z m0-1.666666666666666c-1.8399999999999999 0-3.333333333333332 1.493333333333334-3.333333333333332 3.336666666666668 0 1.8399999999999999 1.4933333333333323 3.333333333333332 3.333333333333332 3.333333333333332 1.8416666666666686 0 3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332 0-1.8416666666666686-1.4916666666666671-3.336666666666666-3.333333333333332-3.336666666666666z' })
                )
            );
        }
    }]);

    return TiKeyOutline;
}(React.Component);

exports.default = TiKeyOutline;
module.exports = exports['default'];