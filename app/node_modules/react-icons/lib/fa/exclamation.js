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

var FaExclamation = function (_React$Component) {
    _inherits(FaExclamation, _React$Component);

    function FaExclamation() {
        _classCallCheck(this, FaExclamation);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaExclamation).apply(this, arguments));
    }

    _createClass(FaExclamation, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm24.285714285714285 27.857142857142858v5.0000000000000036q0 0.5799999999999983-0.4242857142857126 1.0042857142857144t-1.0042857142857144 0.42428571428570905h-5.714285714285715q-0.5800000000000018 0-1.0042857142857144-0.42428571428571615t-0.4242857142857126-1.0042857142857073v-5q0-0.5799999999999983 0.4242857142857126-1.0042857142857144t1.0042857142857144-0.42428571428571615h5.714285714285715q0.5800000000000018 0 1.0042857142857144 0.4242857142857126t0.4242857142857126 1.0042857142857144z m0.6714285714285708-23.571428571428573l-0.6285714285714299 17.142857142857142q-0.02142857142857224 0.5800000000000018-0.4571428571428555 1.0042857142857144t-1.0142857142857125 0.42428571428571615h-5.714285714285715q-0.5800000000000018 0-1.014285714285716-0.4242857142857126t-0.45714285714285374-1.0042857142857144l-0.6285714285714299-17.142857142857146q-0.02142857142857224-0.5799999999999992 0.39142857142857146-1.0042857142857136t0.9942857142857164-0.42428571428571393h7.142857142857142q0.5785714285714292 0 0.9914285714285711 0.4242857142857144t0.389999999999997 1.004285714285714z' })
                )
            );
        }
    }]);

    return FaExclamation;
}(React.Component);

exports.default = FaExclamation;
module.exports = exports['default'];