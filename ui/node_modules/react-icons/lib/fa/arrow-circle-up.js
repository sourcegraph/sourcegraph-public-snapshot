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

var FaArrowCircleUp = function (_React$Component) {
    _inherits(FaArrowCircleUp, _React$Component);

    function FaArrowCircleUp() {
        _classCallCheck(this, FaArrowCircleUp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaArrowCircleUp).apply(this, arguments));
    }

    _createClass(FaArrowCircleUp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.517142857142858 19.97714285714286q0-0.6028571428571432-0.3999999999999986-1.0042857142857144l-10.11714285714286-10.110000000000001q-0.3999999999999986-0.40000000000000036-1.0028571428571418-0.40000000000000036t-1.0042857142857144 0.40000000000000036l-10.107142857142858 10.108571428571429q-0.40285714285714214 0.4028571428571439-0.40285714285714214 1.0057142857142871t0.40000000000000036 1.0042857142857144l2.032857142857143 2.0314285714285703q0.40285714285714214 0.3999999999999986 1.0057142857142853 0.3999999999999986t1.0042857142857144-0.3999999999999986l4.217142857142855-4.220000000000002v11.207142857142859q0 0.5799999999999983 0.4242857142857126 1.0042857142857144t1.004285714285718 0.42428571428571615h2.8571428571428577q0.5799999999999983 0 1.0042857142857144-0.4242857142857126t0.4242857142857126-1.004285714285718v-11.205714285714286l4.21857142857143 4.21857142857143q0.4242857142857126 0.4228571428571435 1.0042857142857144 0.4228571428571435t1.0042857142857144-0.4242857142857126l2.0285714285714285-2.032857142857143q0.4028571428571439-0.3999999999999986 0.4028571428571439-1.0042857142857144z m5.625714285714288 0.022857142857141355q0 4.665714285714287-2.299999999999997 8.604285714285716t-6.237142857142857 6.238571428571426-8.605714285714292 2.3000000000000043-8.6-2.3000000000000043-6.242857142857143-6.238571428571426-2.295714285714286-8.604285714285716 2.3000000000000003-8.604285714285714 6.234285714285714-6.238571428571428 8.604285714285714-2.3000000000000003 8.605714285714285 2.3000000000000003 6.238571428571426 6.238571428571428 2.298571428571435 8.604285714285714z' })
                )
            );
        }
    }]);

    return FaArrowCircleUp;
}(React.Component);

exports.default = FaArrowCircleUp;
module.exports = exports['default'];