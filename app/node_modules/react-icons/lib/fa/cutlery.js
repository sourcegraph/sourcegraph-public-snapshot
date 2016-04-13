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

var FaCutlery = function (_React$Component) {
    _inherits(FaCutlery, _React$Component);

    function FaCutlery() {
        _classCallCheck(this, FaCutlery);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCutlery).apply(this, arguments));
    }

    _createClass(FaCutlery, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.571428571428573 1.4285714285714286v14.285714285714286q0 1.361428571428572-0.7928571428571445 2.4771428571428586t-2.064285714285713 1.5628571428571405v17.38857142857143q0 1.1600000000000037-0.8485714285714288 2.008571428571429t-2.008571428571429 0.8485714285714252h-2.8571428571428577q-1.1600000000000001 0-2.008571428571429-0.8485714285714252t-0.8485714285714279-2.008571428571429v-17.38857142857143q-1.2714285714285714-0.4471428571428575-2.064285714285714-1.562857142857144t-0.7928571428571436-2.477142857142855v-14.285714285714286q0-0.5800000000000003 0.4242857142857144-1.0042857142857144t1.0042857142857144-0.4242857142857144 1.0042857142857144 0.42428571428571427 0.4242857142857144 1.0042857142857144v9.285714285714286q0 0.5800000000000001 0.4242857142857144 1.0042857142857144t1.0042857142857136 0.4242857142857126 1.0042857142857144-0.4242857142857144 0.4242857142857144-1.0042857142857127v-9.285714285714286q0-0.5800000000000003 0.4242857142857144-1.0042857142857144t1.0042857142857144-0.4242857142857144 1.0042857142857144 0.42428571428571427 0.4242857142857144 1.0042857142857144v9.285714285714286q0 0.5800000000000001 0.4242857142857144 1.0042857142857144t1.0042857142857144 0.4242857142857126 1.0042857142857144-0.4242857142857144 0.4242857142857144-1.0042857142857127v-9.285714285714286q0-0.5800000000000003 0.4242857142857126-1.0042857142857144t1.0042857142857144-0.4242857142857144 1.0042857142857144 0.42428571428571427 0.42428571428571615 1.0042857142857144z m17.142857142857142 0v35.714285714285715q0 1.1600000000000037-0.8485714285714252 2.008571428571429t-2.008571428571429 0.8485714285714252h-2.8571428571428577q-1.1600000000000001 0-2.008571428571429-0.8485714285714252t-0.8485714285714323-2.008571428571429v-11.42857142857143h-5q-0.28999999999999915 0-0.5028571428571418-0.21142857142856997t-0.21142857142856997-0.5028571428571453v-17.857142857142858q0-2.9471428571428566 2.1000000000000014-5.042857142857142t5.042857142857141-2.1000000000000005h5.714285714285712q0.5799999999999983 0 1.0042857142857144 0.42428571428571427t0.42428571428571615 1.0042857142857144z' })
                )
            );
        }
    }]);

    return FaCutlery;
}(React.Component);

exports.default = FaCutlery;
module.exports = exports['default'];