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

var FaWrench = function (_React$Component) {
    _inherits(FaWrench, _React$Component);

    function FaWrench() {
        _classCallCheck(this, FaWrench);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaWrench).apply(this, arguments));
    }

    _createClass(FaWrench, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10 32.85714285714286q0-0.5799999999999983-0.4242857142857144-1.0042857142857144t-1.0042857142857144-0.42428571428571615-1.0042857142857144 0.4242857142857126-0.4242857142857135 1.004285714285718 0.4242857142857144 1.0042857142857144 1.0042857142857136 0.42428571428570905 1.0042857142857144-0.42428571428571615 0.4242857142857144-1.0042857142857073z m14.375714285714288-9.374285714285715l-15.222857142857144 15.22285714285714q-0.8257142857142856 0.8257142857142838-2.008571428571429 0.8257142857142838-1.1571428571428575 0-2.0285714285714285-0.8257142857142838l-2.3700000000000006-2.4099999999999966q-0.8457142857142859-0.8028571428571496-0.8457142857142859-2.010000000000005 0-1.1814285714285688 0.8471428571428572-2.028571428571432l15.2-15.202857142857141q0.8714285714285701 2.185714285714287 2.557142857142857 3.87142857142857t3.87142857142857 2.557142857142857z m14.152857142857137-9.711428571428574q0 0.8714285714285719-0.5142857142857125 2.367142857142859-1.048571428571428 2.991428571428571-3.671428571428571 4.854285714285716t-5.771428571428569 1.8642857142857103q-4.12857142857143 0-7.062857142857144-2.935714285714287t-2.937142857142856-7.064285714285713 2.937142857142856-7.064285714285715 7.062857142857144-2.9357142857142855q1.2957142857142863 0 2.7142857142857153 0.36857142857142877t2.400000000000002 1.0385714285714283q0.3571428571428541 0.24571428571428555 0.3571428571428541 0.6257142857142854t-0.3571428571428541 0.6257142857142854l-6.542857142857148 3.774285714285716v5l4.309999999999999 2.3885714285714297q0.1114285714285721-0.0671428571428585 1.7628571428571433-1.0828571428571436t3.027142857142863-1.8100000000000005 1.5742857142857147-0.7928571428571427q0.33428571428571274 0 0.5242857142857176 0.22285714285714242t0.18999999999999773 0.5600000000000005z' })
                )
            );
        }
    }]);

    return FaWrench;
}(React.Component);

exports.default = FaWrench;
module.exports = exports['default'];