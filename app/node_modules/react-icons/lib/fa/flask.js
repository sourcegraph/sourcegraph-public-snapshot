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

var FaFlask = function (_React$Component) {
    _inherits(FaFlask, _React$Component);

    function FaFlask() {
        _classCallCheck(this, FaFlask);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFlask).apply(this, arguments));
    }

    _createClass(FaFlask, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.51428571428571 32.32142857142857q1.2485714285714309 1.9857142857142875 0.47857142857142776 3.404285714285713t-3.135714285714279 1.4171428571428635h-25.71428571428572q-2.3657142857142848 0-3.1357142857142843-1.4171428571428564t0.4800000000000004-3.404285714285713l11.227142857142857-17.700000000000006v-8.907142857142855h-1.4285714285714288q-0.5800000000000001-8.881784197001252e-16-1.0042857142857144-0.42428571428571527t-0.4242857142857144-1.0042857142857144 0.4242857142857144-1.0042857142857144 1.0042857142857144-0.42428571428571393h11.428571428571429q0.5799999999999983 0 1.0042857142857144 0.4242857142857144t0.4242857142857126 1.004285714285714-0.4242857142857126 1.0042857142857144-1.0042857142857144 0.4242857142857144h-1.428571428571427v8.905714285714286z m-17.389999999999997-16.182857142857138l-6.070000000000002 9.575714285714284h15.892857142857144l-6.071428571428569-9.575714285714284-0.4471428571428575-0.6914285714285722v-9.732857142857146h-2.8571428571428577v9.732857142857144z' })
                )
            );
        }
    }]);

    return FaFlask;
}(React.Component);

exports.default = FaFlask;
module.exports = exports['default'];