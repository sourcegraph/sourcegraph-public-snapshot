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

var FaMinusCircle = function (_React$Component) {
    _inherits(FaMinusCircle, _React$Component);

    function FaMinusCircle() {
        _classCallCheck(this, FaMinusCircle);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMinusCircle).apply(this, arguments));
    }

    _createClass(FaMinusCircle, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 21.42857142857143v-2.8571428571428577q0-0.5800000000000018-0.4242857142857126-1.0042857142857144t-1.0042857142857144-0.42428571428571615h-17.142857142857146q-0.5799999999999983 0-1.0042857142857127 0.4242857142857126t-0.4242857142857144 1.004285714285718v2.8571428571428577q0 0.5799999999999983 0.4242857142857144 1.0042857142857144t1.0042857142857144 0.4242857142857126h17.14285714285714q0.5799999999999983 0 1.0042857142857144-0.4242857142857126t0.42428571428571615-1.0042857142857144z m7.142857142857146-1.428571428571427q0 4.665714285714284-2.299999999999997 8.604285714285712t-6.237142857142857 6.238571428571426-8.605714285714292 2.3000000000000043-8.6-2.3000000000000043-6.242857142857143-6.238571428571426-2.295714285714286-8.604285714285716 2.3000000000000003-8.604285714285714 6.234285714285714-6.238571428571428 8.604285714285714-2.3000000000000003 8.605714285714285 2.3000000000000003 6.238571428571426 6.238571428571428 2.298571428571435 8.604285714285714z' })
                )
            );
        }
    }]);

    return FaMinusCircle;
}(React.Component);

exports.default = FaMinusCircle;
module.exports = exports['default'];