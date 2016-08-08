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

var TiCompass = function (_React$Component) {
    _inherits(TiCompass, _React$Component);

    function TiCompass() {
        _classCallCheck(this, TiCompass);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiCompass).apply(this, arguments));
    }

    _createClass(TiCompass, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 8.333333333333334c6.433333333333334 0 11.666666666666668 5.236666666666666 11.666666666666668 11.666666666666666s-5.233333333333334 11.666666666666668-11.666666666666668 11.666666666666668-11.666666666666668-5.233333333333334-11.666666666666668-11.666666666666668 5.233333333333334-11.666666666666668 11.666666666666668-11.666666666666668z m0-3.333333333333334c-8.283333333333333 0-15 6.716666666666669-15 15s6.716666666666669 15 15 15 15-6.716666666666669 15-15-6.716666666666669-15-15-15z m6.969999999999999 8.033333333333335c-0.216666666666665-0.21666666666666679-0.5249999999999986-0.3000000000000007-0.8166666666666664-0.21333333333333293l-9.919999999999998 2.843333333333332c-0.2766666666666673 0.08000000000000007-0.49166666666666714 0.29499999999999993-0.5700000000000003 0.5716666666666654l-2.846666666666666 9.920000000000002c-0.08333333333333393 0.29166666666666785 0 0.6050000000000004 0.21333333333333293 0.8166666666666664 0.15833333333333321 0.16000000000000014 0.37166666666666615 0.245000000000001 0.5899999999999999 0.245000000000001 0.07499999999999929 0 0.1533333333333342-0.011666666666666714 0.23000000000000043-0.03333333333333499l9.916666666666668-2.8500000000000014c0.2749999999999986-0.07666666666666799 0.49166666666666714-0.293333333333333 0.5700000000000003-0.5700000000000003l2.844999999999999-9.916666666666668c0.08333333333333215-0.28666666666666707 0-0.5999999999999996-0.2133333333333347-0.8166666666666664z m-12.136666666666665 12.133333333333333l2.3050000000000015-8.026666666666667 5.723333333333333 5.725000000000001-8.028333333333334 2.301666666666666z' })
                )
            );
        }
    }]);

    return TiCompass;
}(React.Component);

exports.default = TiCompass;
module.exports = exports['default'];