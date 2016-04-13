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

var FaFastBackward = function (_React$Component) {
    _inherits(FaFastBackward, _React$Component);

    function FaFastBackward() {
        _classCallCheck(this, FaFastBackward);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFastBackward).apply(this, arguments));
    }

    _createClass(FaFastBackward, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm38.995714285714286 3.147142857142857q0.42428571428571615-0.4242857142857144 0.7142857142857153-0.29000000000000004t0.28999999999999915 0.7142857142857144v32.85714285714286q0 0.5799999999999983-0.28999999999999915 0.7142857142857153t-0.7142857142857153-0.28999999999999915l-15.848571428571429-15.848571428571432q-0.1999999999999993-0.1999999999999993-0.28999999999999915-0.4242857142857126v15.848571428571429q0 0.5799999999999983-0.28999999999999915 0.7142857142857153t-0.7142857142857153-0.28999999999999915l-15.848571428571429-15.848571428571432q-0.20000000000000018-0.1999999999999993-0.29000000000000004-0.4242857142857126v15.134285714285713q0 0.5799999999999983-0.4242857142857144 1.0042857142857144t-1.0042857142857144 0.42428571428571615h-2.8571428571428568q-0.5800000000000003 0-1.0042857142857144-0.42428571428571615t-0.4242857142857144-1.0042857142857144v-31.42857142857143q0-0.5799999999999992 0.42428571428571427-1.0042857142857136t1.0042857142857144-0.42428571428571393h2.8571428571428568q0.5800000000000001 0 1.0042857142857144 0.4242857142857144t0.4242857142857144 1.004285714285714v15.134285714285713q0.08999999999999986-0.24571428571428555 0.29000000000000004-0.4242857142857126l15.848571428571429-15.848571428571429q0.4242857142857126-0.4242857142857144 0.7142857142857153-0.29000000000000004t0.28999999999999915 0.7142857142857144v15.84857142857143q0.08999999999999986-0.24571428571428555 0.28999999999999915-0.4242857142857126z' })
                )
            );
        }
    }]);

    return FaFastBackward;
}(React.Component);

exports.default = FaFastBackward;
module.exports = exports['default'];