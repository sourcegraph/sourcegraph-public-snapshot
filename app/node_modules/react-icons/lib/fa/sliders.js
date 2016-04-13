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

var FaSliders = function (_React$Component) {
    _inherits(FaSliders, _React$Component);

    function FaSliders() {
        _classCallCheck(this, FaSliders);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaSliders).apply(this, arguments));
    }

    _createClass(FaSliders, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10.714285714285715 31.42857142857143v2.857142857142854h-7.857142857142858v-2.8571428571428577h7.857142857142858z m7.857142857142858-2.8571428571428577q0.5800000000000018 0 1.0042857142857144 0.4242857142857126t0.4242857142857126 1.0042857142857144v5.714285714285715q0 0.5799999999999983-0.4242857142857126 1.0042857142857144t-1.0042857142857144 0.42428571428571615h-5.714285714285715q-0.5800000000000001 0-1.0042857142857144-0.42428571428571615t-0.4242857142857144-1.0042857142857144v-5.714285714285715q0-0.5799999999999983 0.4242857142857144-1.0042857142857144t1.0042857142857144-0.4242857142857126h5.714285714285715z m3.571428571428573-8.571428571428573v2.8571428571428577h-19.28571428571429v-2.8571428571428577h19.285714285714285z m-14.285714285714286-11.428571428571429v2.8571428571428577h-5.000000000000002v-2.8571428571428577h5z m29.285714285714285 22.85714285714286v2.857142857142854h-16.42857142857143v-2.8571428571428577h16.42857142857143z m-21.42857142857143-25.714285714285715q0.5799999999999983-8.881784197001252e-16 1.0042857142857144 0.4242857142857135t0.4242857142857126 1.0042857142857144v5.714285714285714q0 0.5800000000000001-0.4242857142857126 1.0042857142857144t-1.0042857142857144 0.4242857142857144h-5.714285714285715q-0.5800000000000001 0-1.0042857142857144-0.4242857142857144t-0.4242857142857144-1.0042857142857144v-5.714285714285714q0-0.5800000000000001 0.4242857142857144-1.0042857142857144t1.0042857142857144-0.4242857142857144h5.714285714285715z m14.285714285714285 11.428571428571427q0.5799999999999983 0 1.0042857142857144 0.4242857142857126t0.42428571428571615 1.004285714285718v5.714285714285715q0 0.5800000000000018-0.4242857142857126 1.0042857142857144t-1.004285714285718 0.4242857142857126h-5.714285714285715q-0.5799999999999983 0-1.0042857142857144-0.4242857142857126t-0.4242857142857126-1.004285714285718v-5.714285714285715q0-0.5799999999999983 0.4242857142857126-1.0042857142857144t1.0042857142857144-0.4242857142857126h5.714285714285715z m7.142857142857146 2.8571428571428577v2.8571428571428577h-5v-2.8571428571428577h5z m0-11.428571428571429v2.8571428571428577h-19.28571428571429v-2.8571428571428577h19.28571428571429z' })
                )
            );
        }
    }]);

    return FaSliders;
}(React.Component);

exports.default = FaSliders;
module.exports = exports['default'];