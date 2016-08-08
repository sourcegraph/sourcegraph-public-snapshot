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

var FaVolumeDown = function (_React$Component) {
    _inherits(FaVolumeDown, _React$Component);

    function FaVolumeDown() {
        _classCallCheck(this, FaVolumeDown);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaVolumeDown).apply(this, arguments));
    }

    _createClass(FaVolumeDown, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm24.285714285714285 7.857142857142858v24.28571428571428q0 0.5799999999999983-0.4242857142857126 1.0042857142857144t-1.0042857142857144 0.42428571428571615-1.0042857142857144-0.42428571428571615l-7.4328571428571415-7.432857142857138h-5.8485714285714305q-0.5800000000000001 0-1.0042857142857144-0.4242857142857126t-0.4242857142857135-1.004285714285718v-8.571428571428571q0-0.5800000000000001 0.4242857142857144-1.0042857142857144t1.0042857142857136-0.4242857142857126h5.848571428571429l7.432857142857143-7.432857142857143q0.4242857142857126-0.4242857142857144 1.0042857142857144-0.4242857142857144t1.0042857142857144 0.4242857142857144 0.4242857142857126 1.0042857142857144z m8.57142857142857 12.142857142857142q0 1.6971428571428575-0.9485714285714302 3.1571428571428584t-2.5114285714285707 2.088571428571427q-0.2228571428571442 0.1114285714285721-0.5571428571428569 0.1114285714285721-0.581428571428571 0-1.0057142857142871-0.4142857142857146t-0.4242857142857126-1.0142857142857125q0-0.46857142857142975 0.2671428571428578-0.7928571428571445t0.6471428571428568-0.5571428571428569 0.7571428571428562-0.5142857142857125 0.6485714285714295-0.7928571428571445 0.2700000000000031-1.2714285714285722-0.2671428571428578-1.274285714285714-0.6457142857142841-0.7928571428571445-0.7571428571428562-0.514285714285716-0.6485714285714295-0.5571428571428569-0.2671428571428578-0.7928571428571445q0-0.6028571428571432 0.4242857142857126-1.0142857142857142t1.0042857142857144-0.4142857142857146q0.33428571428571274 0 0.5571428571428569 0.1114285714285721 1.5642857142857132 0.6028571428571432 2.5142857142857125 2.0757142857142856t0.9428571428571502 3.1685714285714326z' })
                )
            );
        }
    }]);

    return FaVolumeDown;
}(React.Component);

exports.default = FaVolumeDown;
module.exports = exports['default'];