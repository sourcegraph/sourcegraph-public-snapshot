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

var TiSocialLinkedinCircular = function (_React$Component) {
    _inherits(TiSocialLinkedinCircular, _React$Component);

    function TiSocialLinkedinCircular() {
        _classCallCheck(this, TiSocialLinkedinCircular);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiSocialLinkedinCircular).apply(this, arguments));
    }

    _createClass(TiSocialLinkedinCircular, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.721666666666668 25.500000000000004h-2.666666666666668v-8.666666666666668h2.666666666666668v8.666666666666668z m-1.333333333333334-9.776666666666667c-0.961666666666666 0-1.4433333333333334-0.4450000000000003-1.4433333333333334-1.333333333333334 0-0.37166666666666615 0.13666666666666671-0.6866666666666674 0.4166666666666661-0.9450000000000003 0.2766666666666673-0.25833333333333286 0.6166666666666671-0.38833333333333364 1.0266666666666673-0.38833333333333364 0.9616666666666678 0 1.4433333333333334 0.4466666666666672 1.4433333333333334 1.333333333333334s-0.4833333333333343 1.333333333333334-1.4450000000000003 1.333333333333334z m11.223333333333334 9.776666666666667h-2.7216666666666676v-4.833333333333332c0-1.2583333333333329-0.44666666666666544-1.8883333333333319-1.3333333333333321-1.8883333333333319-0.7049999999999983 0-1.1666666666666679 0.3500000000000014-1.3916666666666657 1.0549999999999997-0.07333333333333414 0.11166666666666814-0.11166666666666814 0.33333333333333215-0.11166666666666814 0.6666666666666679v5h-2.719999999999999v-5.888333333333335c0-1.3333333333333321-0.019999999999999574-2.258333333333333-0.054999999999999716-2.7766666666666673h2.333333333333332l0.16666666666666785 1.1666666666666679c0.6116666666666681-0.9266666666666659 1.5-1.3883333333333319 2.7216666666666676-1.3883333333333319 0.9283333333333346 0 1.6766666666666659 0.32333333333333414 2.25 0.9716666666666676 0.576666666666668 0.6499999999999986 0.8633333333333333 1.5833333333333321 0.8633333333333333 2.806666666666665v5.108333333333331z m-6.611666666666668 9.499999999999996c-8.271666666666667 0-15-6.728333333333332-15-15s6.7283333333333335-15 15-15 15 6.7283333333333335 15 15-6.728333333333332 15-15 15z m0-26.666666666666668c-6.433333333333334 0-11.666666666666668 5.233333333333334-11.666666666666668 11.666666666666668s5.233333333333334 11.666666666666668 11.666666666666668 11.666666666666668 11.666666666666668-5.233333333333334 11.666666666666668-11.666666666666668-5.233333333333334-11.666666666666668-11.666666666666668-11.666666666666668z' })
                )
            );
        }
    }]);

    return TiSocialLinkedinCircular;
}(React.Component);

exports.default = TiSocialLinkedinCircular;
module.exports = exports['default'];