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

var TiTime = function (_React$Component) {
    _inherits(TiTime, _React$Component);

    function TiTime() {
        _classCallCheck(this, TiTime);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiTime).apply(this, arguments));
    }

    _createClass(TiTime, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.666666666666668 21.666666666666668c0-0.9166666666666679-0.75-1.6666666666666679-1.6666666666666679-1.6666666666666679h-5c-0.9166666666666679 0-1.6666666666666679 0.75-1.6666666666666679 1.6666666666666679s0.75 1.6666666666666679 1.6666666666666679 1.6666666666666679h5c0.9166666666666679 0 1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679z m-6.666666666666668-11.666666666666668c6.433333333333334 0 11.666666666666668 5.233333333333334 11.666666666666668 11.666666666666668s-5.233333333333334 11.666666666666668-11.666666666666668 11.666666666666668-11.666666666666668-5.233333333333334-11.666666666666668-11.666666666666668 5.233333333333334-11.666666666666668 11.666666666666668-11.666666666666668z m0-3.333333333333334c-8.283333333333333 0-15 6.716666666666667-15 14.999999999999998s6.716666666666669 15 15 15 15-6.716666666666669 15-15-6.716666666666669-15-15-15z m1.6666666666666679 10.000000000000002c0-0.9166666666666661-0.75-1.666666666666666-1.6666666666666679-1.666666666666666s-1.6666666666666679 0.75-1.6666666666666679 1.666666666666666v5c0 0.9166666666666679 0.75 1.6666666666666679 1.6666666666666679 1.6666666666666679s1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679v-5z m-1.6666666666666679-3.333333333333334c4.594999999999999 0 8.333333333333336 3.7383333333333315 8.333333333333336 8.333333333333334s-3.7383333333333333 8.333333333333332-8.333333333333336 8.333333333333332-8.333333333333334-3.7383333333333333-8.333333333333334-8.333333333333336 3.7383333333333333-8.333333333333334 8.333333333333334-8.333333333333334z m0-1.666666666666666c-5.52 0-10 4.476666666666667-10 10 0 5.52 4.48 10 10 10s10-4.48 10-10c0-5.523333333333333-4.48-10-10-10z' })
                )
            );
        }
    }]);

    return TiTime;
}(React.Component);

exports.default = TiTime;
module.exports = exports['default'];