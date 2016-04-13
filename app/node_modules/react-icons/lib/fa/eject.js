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

var FaEject = function (_React$Component) {
    _inherits(FaEject, _React$Component);

    function FaEject() {
        _classCallCheck(this, FaEject);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaEject).apply(this, arguments));
    }

    _createClass(FaEject, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm3.1714285714285717 21.852857142857143l15.845714285714287-15.848571428571429q0.4242857142857126-0.4242857142857144 1.0042857142857144-0.4242857142857144t1.0042857142857144 0.4242857142857144l15.847142857142856 15.848571428571429q0.42428571428571615 0.4242857142857126 0.28999999999999915 0.7142857142857153t-0.7142857142857153 0.28999999999999915h-32.85714285714286q-0.5800000000000001 0-0.7142857142857144-0.28999999999999915t0.29000000000000004-0.7142857142857153z m32.56428571428572 12.432857142857141h-31.42857142857143q-0.5800000000000001 0-1.0042857142857144-0.42428571428571615t-0.4228571428571426-1.0042857142857073v-5.714285714285715q0-0.5799999999999983 0.4242857142857144-1.0042857142857144t1.0042857142857144-0.42428571428571615h31.428571428571427q0.5799999999999983 0 1.0042857142857144 0.4242857142857126t0.42285714285713993 1.0042857142857144v5.714285714285712q0 0.5799999999999983-0.42428571428571615 1.0042857142857144t-1.0057142857142836 0.42428571428571615z' })
                )
            );
        }
    }]);

    return FaEject;
}(React.Component);

exports.default = FaEject;
module.exports = exports['default'];