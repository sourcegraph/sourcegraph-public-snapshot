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

var FaBattery1 = function (_React$Component) {
    _inherits(FaBattery1, _React$Component);

    function FaBattery1() {
        _classCallCheck(this, FaBattery1);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaBattery1).apply(this, arguments));
    }

    _createClass(FaBattery1, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm4.444444444444445 26.666666666666668v-13.333333333333334h8.88888888888889v13.333333333333334h-8.88888888888889z m33.333333333333336-12.222222222222223q0.9200000000000017 0 1.571111111111108 0.6511111111111116t0.6511111111111134 1.5711111111111116v6.666666666666668q0 0.9200000000000017-0.6511111111111134 1.5711111111111116t-1.571111111111108 0.6511111111111099v2.7777777777777786q0 1.1444444444444457-0.8155555555555551 1.9622222222222234t-1.9622222222222234 0.8155555555555551h-32.22222222222222q-1.1444444444444446 0-1.9622222222222223-0.8155555555555551t-0.8155555555555563-1.9622222222222234v-16.666666666666668q0-1.144444444444444 0.8155555555555556-1.9622222222222216t1.962222222222222-0.8155555555555569h32.22222222222222q1.1444444444444457 0 1.9622222222222234 0.8155555555555551t0.8155555555555551 1.9622222222222234v2.777777777777777z m0 8.888888888888891v-6.666666666666668h-2.2222222222222214v-5q0-0.24444444444444358-0.15555555555555856-0.40000000000000036t-0.3999999999999986-0.15555555555555678h-32.22222222222222q-0.24444444444444446 0-0.3999999999999999 0.155555555555555t-0.15555555555555634 0.40000000000000213v16.666666666666668q0 0.24444444444444358 0.15555555555555545 0.3999999999999986t0.3999999999999999 0.155555555555555h32.22222222222222q0.24444444444444713 0 0.3999999999999986-0.155555555555555t0.15555555555555856-0.3999999999999986v-5h2.2222222222222214z' })
                )
            );
        }
    }]);

    return FaBattery1;
}(React.Component);

exports.default = FaBattery1;
module.exports = exports['default'];