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

var FaGamepad = function (_React$Component) {
    _inherits(FaGamepad, _React$Component);

    function FaGamepad() {
        _classCallCheck(this, FaGamepad);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaGamepad).apply(this, arguments));
    }

    _createClass(FaGamepad, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.333333333333332 24v-2.666666666666668q0-0.293333333333333-0.18666666666666742-0.4800000000000004t-0.4800000000000004-0.18666666666666742h-4v-4q0-0.293333333333333-0.18666666666666742-0.4800000000000004t-0.4799999999999969-0.18666666666666387h-2.666666666666668q-0.293333333333333 0-0.4800000000000004 0.18666666666666742t-0.18666666666666742 0.4800000000000004v4h-4q-0.293333333333333 0-0.4800000000000004 0.18666666666666742t-0.18666666666666387 0.4799999999999969v2.666666666666668q0 0.293333333333333 0.18666666666666654 0.4800000000000004t0.47999999999999954 0.18666666666666742h4v4q0 0.293333333333333 0.18666666666666742 0.4800000000000004t0.47999999999999865 0.18666666666666387h2.666666666666666q0.293333333333333 0 0.4800000000000004-0.18666666666666742t0.18666666666666742-0.4800000000000004v-4h3.9999999999999982q0.293333333333333 0 0.4800000000000004-0.18666666666666742t0.18666666666666742-0.4799999999999969z m12 1.3333333333333321q0-1.1039999999999992-0.7813333333333325-1.8853333333333318t-1.8853333333333353-0.7813333333333361-1.8853333333333318 0.7813333333333325-0.7813333333333325 1.8853333333333353 0.7813333333333325 1.8853333333333318 1.8853333333333318 0.7813333333333361 1.8853333333333318-0.7813333333333325 0.7813333333333361-1.8853333333333353z m5.333333333333332-5.333333333333332q0-1.1039999999999992-0.7813333333333361-1.8853333333333318t-1.8853333333333282-0.7813333333333361-1.8853333333333318 0.7813333333333325-0.7813333333333361 1.8853333333333353 0.7813333333333325 1.8853333333333318 1.8853333333333353 0.7813333333333325 1.8853333333333353-0.7813333333333325 0.781333333333329-1.8853333333333318z m5.333333333333336 2.666666666666668q0 4.417333333333335-3.12533333333333 7.541333333333331t-7.541333333333338 3.12533333333333q-4 0-7.039999999999999-2.666666666666668h-4.58666666666667q-3.039999999999999 2.666666666666668-7.039999999999999 2.666666666666668-4.418666666666667 0-7.542666666666666-3.1253333333333337t-3.123999999999998-7.541333333333331 3.126666666666667-7.541333333333332 7.539999999999999-3.125333333333332h18.666666666666664q4.418666666666667 0 7.542666666666669 3.1253333333333337t3.1240000000000023 7.541333333333331z' })
                )
            );
        }
    }]);

    return FaGamepad;
}(React.Component);

exports.default = FaGamepad;
module.exports = exports['default'];