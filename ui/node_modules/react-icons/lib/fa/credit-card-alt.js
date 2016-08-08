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

var FaCreditCardAlt = function (_React$Component) {
    _inherits(FaCreditCardAlt, _React$Component);

    function FaCreditCardAlt() {
        _classCallCheck(this, FaCreditCardAlt);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCreditCardAlt).apply(this, arguments));
    }

    _createClass(FaCreditCardAlt, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm0 30.555555555555557v-10.555555555555557h40v10.555555555555557q0 1.1444444444444457-0.8155555555555551 1.9622222222222234t-1.9622222222222234 0.8155555555555551h-34.44444444444444q-1.1444444444444446 0-1.9622222222222223-0.8155555555555551t-0.8155555555555563-1.9622222222222234z m11.11111111111111-3.8888888888888893v2.2222222222222214h6.666666666666668v-2.2222222222222214h-6.666666666666668z m-6.666666666666667 0v2.2222222222222214h4.4444444444444455v-2.2222222222222214h-4.444444444444445z m32.77777777777778-20q1.1444444444444457-8.881784197001252e-16 1.9622222222222234 0.8155555555555551t0.8155555555555551 1.9622222222222216v3.8888888888888893h-40v-3.8888888888888893q0-1.144444444444444 0.8155555555555556-1.9622222222222225t1.962222222222222-0.8155555555555551h34.44444444444444z' })
                )
            );
        }
    }]);

    return FaCreditCardAlt;
}(React.Component);

exports.default = FaCreditCardAlt;
module.exports = exports['default'];