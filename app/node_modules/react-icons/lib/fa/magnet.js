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

var FaMagnet = function (_React$Component) {
    _inherits(FaMagnet, _React$Component);

    function FaMagnet() {
        _classCallCheck(this, FaMagnet);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMagnet).apply(this, arguments));
    }

    _createClass(FaMagnet, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.142857142857146 18.571428571428573v2.8571428571428577q0 4.4857142857142875-2.200000000000003 8.079999999999998t-6.114285714285714 5.614285714285714-8.82857142857143 2.020000000000003-8.82857142857143-2.020000000000003-6.114285714285715-5.614285714285714-2.199999999999999-8.079999999999998v-2.8571428571428577q0-0.5800000000000018 0.4242857142857144-1.0042857142857144t1.004285714285714-0.42428571428571615h8.571428571428573q0.5800000000000001 0 1.0042857142857144 0.4242857142857126t0.4242857142857144 1.004285714285718v2.8571428571428577q0 1.1600000000000001 0.524285714285714 2.008571428571429t1.194285714285714 1.2714285714285722 1.5857142857142854 0.6714285714285708 1.428571428571427 0.28999999999999915 0.9814285714285731 0.044285714285713595 0.9828571428571422-0.04285714285714448 1.428571428571427-0.2914285714285718 1.5857142857142854-0.6714285714285708 1.192857142857143-1.2714285714285722 0.5242857142857176-2.0085714285714253v-2.8571428571428577q0-0.5800000000000018 0.4242857142857126-1.0042857142857144t1.0042857142857144-0.42428571428571615h8.571428571428573q0.5799999999999983 0 1.0042857142857144 0.4242857142857126t0.42428571428571615 1.004285714285718z m-22.85714285714286-14.285714285714288v8.571428571428573q1.7763568394002505e-15 0.5800000000000001-0.4242857142857126 1.0042857142857144t-1.0042857142857144 0.4242857142857144h-8.571428571428573q-0.5799999999999992 0-1.0042857142857136-0.4242857142857144t-0.42428571428571393-1.0042857142857144v-8.571428571428573q0-0.5799999999999992 0.4242857142857144-1.0042857142857136t1.004285714285714-0.42428571428571393h8.571428571428573q0.5800000000000001 0 1.0042857142857144 0.4242857142857144t0.4242857142857144 1.004285714285714z m22.85714285714286 8.881784197001252e-16v8.571428571428573q0 0.5800000000000001-0.42428571428571615 1.0042857142857144t-1.0042857142857144 0.4242857142857144h-8.57142857142857q-0.5799999999999983 0-1.0042857142857144-0.4242857142857144t-0.42428571428571615-1.0042857142857144v-8.571428571428573q0-0.5799999999999992 0.4242857142857126-1.0042857142857136t1.0042857142857144-0.42428571428571393h8.571428571428573q0.5799999999999983 0 1.0042857142857144 0.4242857142857144t0.42428571428571615 1.004285714285714z' })
                )
            );
        }
    }]);

    return FaMagnet;
}(React.Component);

exports.default = FaMagnet;
module.exports = exports['default'];