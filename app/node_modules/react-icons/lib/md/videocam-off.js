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

var MdVideocamOff = function (_React$Component) {
    _inherits(MdVideocamOff, _React$Component);

    function MdVideocamOff() {
        _classCallCheck(this, MdVideocamOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdVideocamOff).apply(this, arguments));
    }

    _createClass(MdVideocamOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm5.466666666666667 3.3600000000000003l29.53333333333333 29.53333333333334-2.1099999999999923 2.106666666666662-5.313333333333333-5.311666666666667q-0.466666666666665 0.3133333333333326-0.9383333333333326 0.3133333333333326h-20q-0.7033333333333331 0-1.1716666666666669-0.466666666666665t-0.46666666666667567-1.173333333333332v-16.71666666666667q0-0.7049999999999983 0.4666666666666668-1.173333333333332t1.1716666666666669-0.4666666666666668h1.25l-4.533333333333334-4.533333333333334z m29.53333333333333 7.500000000000001v17.81333333333334l-18.671666666666663-18.67333333333334h10.313333333333333q0.7033333333333331 0 1.211666666666666 0.47000000000000064t0.5083333333333329 1.1716666666666669v5.8583333333333325z' })
                )
            );
        }
    }]);

    return MdVideocamOff;
}(React.Component);

exports.default = MdVideocamOff;
module.exports = exports['default'];