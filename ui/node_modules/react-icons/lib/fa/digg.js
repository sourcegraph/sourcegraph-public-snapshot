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

var FaDigg = function (_React$Component) {
    _inherits(FaDigg, _React$Component);

    function FaDigg() {
        _classCallCheck(this, FaDigg);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaDigg).apply(this, arguments));
    }

    _createClass(FaDigg, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm6.40625 8.0075h3.9837500000000006v19.2h-10.39v-13.61625h6.404999999999999v-5.587499999999999z m0 15.99625v-7.2075h-2.4025v7.2075h2.4025z m5.587499999999999-10.41v13.6125h4.0024999999999995v-13.611250000000002h-4.003749999999998z m0-5.587499999999999v3.9849999999999994h4.0024999999999995v-3.9837500000000006h-4.003749999999998z m5.603749999999998 5.587499999999999h10.41v18.400000000000002h-10.41v-3.1849999999999987h6.40625v-1.6000000000000014h-6.40625v-13.61625z m6.40625 10.41v-7.2075h-2.403749999999995v7.2075h2.4037499999999987z m5.606249999999999-10.41h10.390000000000004v18.400000000000002h-10.3875v-3.1849999999999987h6.387499999999999v-1.6000000000000014h-6.387499999999999v-13.61625z m6.387499999999999 10.41v-7.2075h-2.4037500000000023v7.2075h2.4024999999999963z' })
                )
            );
        }
    }]);

    return FaDigg;
}(React.Component);

exports.default = FaDigg;
module.exports = exports['default'];