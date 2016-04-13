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

var TiSocialTumblerCircular = function (_React$Component) {
    _inherits(TiSocialTumblerCircular, _React$Component);

    function TiSocialTumblerCircular() {
        _classCallCheck(this, TiSocialTumblerCircular);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiSocialTumblerCircular).apply(this, arguments));
    }

    _createClass(TiSocialTumblerCircular, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.96166666666667 25.983333333333334v-2.056666666666665c-0.6666666666666679 0.44666666666666544-1.3133333333333326 0.6666666666666679-1.9433333333333316 0.6666666666666679-0.2950000000000017 0-0.6266666666666652-0.09499999999999886-1-0.2766666666666673-0.22166666666666757-0.14999999999999858-0.3500000000000014-0.31666666666666643-0.38833333333333186-0.5-0.10999999999999943-0.22333333333333272-0.16666666666666785-0.7049999999999983-0.16666666666666785-1.4466666666666654v-3.2766666666666673h3.0566666666666684v-2.0549999999999997h-3.0549999999999997v-3.278333333333336h-1.7816666666666698c-0.14666666666666828 0.7783333333333342-0.2950000000000017 1.333333333333334-0.44333333333333513 1.666666666666666-0.18333333333333357 0.4066666666666663-0.4800000000000004 0.7783333333333342-0.8883333333333319 1.1100000000000012-0.4066666666666663 0.33333333333333215-0.8333333333333321 0.5749999999999993-1.2766666666666673 0.7233333333333327v1.8333333333333321h1.3883333333333319v4.5c0 0.5199999999999996 0.07333333333333414 0.961666666666666 0.22333333333333272 1.3333333333333321 0.11166666666666814 0.29666666666666686 0.33333333333333215 0.591666666666665 0.6666666666666679 0.8883333333333319 0.25833333333333286 0.25833333333333286 0.629999999999999 0.4800000000000004 1.1116666666666681 0.6666666666666679 0.591666666666665 0.14999999999999858 1.1099999999999994 0.21999999999999886 1.5566666666666684 0.21999999999999886 0.5199999999999996 0 1-0.054999999999999716 1.4450000000000003-0.16666666666666785 0.5199999999999996-0.11166666666666814 1.0199999999999996-0.29666666666666686 1.5-0.5533333333333346z m-3.9616666666666696 9.016666666666666c-8.271666666666667 0-15-6.728333333333332-15-15s6.7283333333333335-15 15-15 15 6.7283333333333335 15 15-6.728333333333332 15-15 15z m0-26.666666666666668c-6.433333333333334 0-11.666666666666668 5.233333333333334-11.666666666666668 11.666666666666668s5.233333333333334 11.666666666666668 11.666666666666668 11.666666666666668 11.666666666666668-5.233333333333334 11.666666666666668-11.666666666666668-5.233333333333334-11.666666666666668-11.666666666666668-11.666666666666668z' })
                )
            );
        }
    }]);

    return TiSocialTumblerCircular;
}(React.Component);

exports.default = TiSocialTumblerCircular;
module.exports = exports['default'];