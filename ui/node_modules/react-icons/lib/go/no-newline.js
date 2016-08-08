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

var GoNoNewline = function (_React$Component) {
    _inherits(GoNoNewline, _React$Component);

    function GoNoNewline() {
        _classCallCheck(this, GoNoNewline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoNoNewline).apply(this, arguments));
    }

    _createClass(GoNoNewline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 12.5v5h-5v-5l-7.5 7.5 7.5 7.5v-5h7.5s2.5-0.015000000000000568 2.5-2.5v-7.5h-5z m-26.25-1.25c-4.8325000000000005 0-8.75 3.9175000000000004-8.75 8.75s3.9175 8.75 8.75 8.75 8.75-3.916249999999998 8.75-8.75c0-4.8325-3.9175000000000004-8.75-8.75-8.75z m-5 8.75c0-2.7600000000000016 2.24-5 5-5 0.7324999999999999 0 1.4224999999999994 0.16499999999999915 2.0500000000000007 0.4499999999999993l-6.6000000000000005 6.600000000000001c-0.2825000000000002-0.6300000000000026-0.4500000000000002-1.3187499999999979-0.4500000000000002-2.0500000000000007z m5 5c-0.7324999999999999 0-1.4224999999999994-0.16625000000000156-2.05-0.4499999999999993l6.6000000000000005-6.600000000000001c0.28500000000000014 0.6287500000000001 0.4499999999999993 1.3187500000000014 0.4499999999999993 2.0500000000000007 0 2.7600000000000016-2.241250000000001 5-5 5z' })
                )
            );
        }
    }]);

    return GoNoNewline;
}(React.Component);

exports.default = GoNoNewline;
module.exports = exports['default'];