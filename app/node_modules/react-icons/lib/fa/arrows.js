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

var FaArrows = function (_React$Component) {
    _inherits(FaArrows, _React$Component);

    function FaArrows() {
        _classCallCheck(this, FaArrows);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaArrows).apply(this, arguments));
    }

    _createClass(FaArrows, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm40 20q0 0.5799999999999983-0.42428571428571615 1.0042857142857144l-5.714285714285715 5.714285714285715q-0.42428571428570905 0.4242857142857126-1.0042857142857073 0.4242857142857126t-1.0042857142857144-0.4242857142857126-0.42428571428571615-1.0042857142857144v-2.8571428571428577h-8.57142857142857v8.571428571428573h2.8571428571428577q0.5800000000000018 0 1.0042857142857144 0.4242857142857126t0.42428571428570905 1.004285714285718-0.4242857142857126 1.0042857142857144l-5.714285714285715 5.714285714285715q-0.4242857142857126 0.42428571428570905-1.0042857142857144 0.42428571428570905t-1.0042857142857144-0.42428571428571615l-5.714285714285715-5.714285714285715q-0.4242857142857126-0.42428571428570905-0.4242857142857126-1.0042857142857073t0.4242857142857144-1.0042857142857144 1.0042857142857144-0.42428571428571615h2.857142857142856v-8.57142857142857h-8.571428571428571v2.8571428571428577q0 0.5800000000000018-0.4242857142857144 1.0042857142857144t-1.0042857142857136 0.42428571428570905-1.0042857142857144-0.4242857142857126l-5.714285714285714-5.714285714285715q-0.4242857142857144-0.4242857142857126-0.4242857142857144-1.0042857142857144t0.42428571428571427-1.0042857142857144l5.714285714285714-5.714285714285715q0.4242857142857144-0.4242857142857126 1.0042857142857144-0.4242857142857126t1.0042857142857136 0.4242857142857144 0.4242857142857144 1.0042857142857144v2.857142857142856h8.571428571428571v-8.571428571428571h-2.8571428571428577q-0.5800000000000001 0-1.0042857142857144-0.4242857142857144t-0.4242857142857126-1.0042857142857136 0.4242857142857144-1.0042857142857144l5.7142857142857135-5.714285714285714q0.4242857142857126-0.4242857142857144 1.0042857142857144-0.4242857142857144t1.0042857142857144 0.42428571428571427l5.714285714285715 5.714285714285714q0.4242857142857126 0.4242857142857144 0.4242857142857126 1.0042857142857144t-0.4242857142857126 1.0042857142857136-1.0042857142857144 0.4242857142857144h-2.8571428571428577v8.571428571428571h8.571428571428573v-2.8571428571428577q0-0.5800000000000001 0.4242857142857126-1.0042857142857144t1.004285714285718-0.4242857142857126 1.0042857142857144 0.4242857142857144l5.714285714285715 5.7142857142857135q0.42428571428570905 0.4242857142857126 0.42428571428570905 1.0042857142857144z' })
                )
            );
        }
    }]);

    return FaArrows;
}(React.Component);

exports.default = FaArrows;
module.exports = exports['default'];