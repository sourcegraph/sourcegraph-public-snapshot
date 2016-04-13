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

var MdSettingsInputHdmi = function (_React$Component) {
    _inherits(MdSettingsInputHdmi, _React$Component);

    function MdSettingsInputHdmi() {
        _classCallCheck(this, MdSettingsInputHdmi);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSettingsInputHdmi).apply(this, arguments));
    }

    _createClass(MdSettingsInputHdmi, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.360000000000001 6.640000000000001v5h3.283333333333333v-3.283333333333333h1.716666666666665v3.283333333333333h3.2833333333333314v-3.283333333333333h1.7166666666666686v3.283333333333333h3.2833333333333314v-5h-13.283333333333333z m16.64 5h1.6400000000000006v10l-5 10v5h-13.283333333333333v-5l-5-10v-10h1.6433333333333326v-5q0-1.3283333333333331 1.0133333333333336-2.3049999999999997t2.3433333333333337-0.9750000000000001h13.283333333333333q1.326666666666668 0 2.3416666666666686 0.976666666666667t1.0133333333333319 2.3066666666666666v5z' })
                )
            );
        }
    }]);

    return MdSettingsInputHdmi;
}(React.Component);

exports.default = MdSettingsInputHdmi;
module.exports = exports['default'];