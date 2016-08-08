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

var GoSync = function (_React$Component) {
    _inherits(GoSync, _React$Component);

    function GoSync() {
        _classCallCheck(this, GoSync);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoSync).apply(this, arguments));
    }

    _createClass(GoSync, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30.60375 18.494999999999997c0.4624999999999986 3.192499999999999-0.5100000000000016 6.553750000000001-3 9.0075-3.678750000000001 3.629999999999999-9.355 4.077500000000001-13.53 1.3500000000000014l2.924999999999999-2.8500000000000014-10.748750000000001-1.5024999999999977 1.5 10.5 3.28125-3.1449999999999996c5.895 4.346250000000001 14.254999999999999 3.9349999999999987 19.6125-1.3500000000000014 3.1037500000000016-3.0625 4.52-7.131249999999998 4.345000000000002-11.145l-4.383749999999999-0.8649999999999984z m-18.20375-5.992499999999996c3.6787499999999973-3.629999999999999 9.353749999999998-4.08 13.528749999999999-1.3499999999999996l-2.9299999999999997 2.8499999999999996 10.75 1.5-1.4987500000000011-10.502500000000001-3.2787500000000023 3.1500000000000004c-5.896249999999995-4.34625-14.253749999999997-3.9312500000000004-19.612499999999997 1.3499999999999996-3.102500000000001 3.0625-4.5175 7.1325-4.343750000000001 11.144999999999996l4.3875 0.8625000000000007c-0.46499999999999986-3.192499999999999 0.5124999999999993-6.550000000000001 3-9.004999999999999z' })
                )
            );
        }
    }]);

    return GoSync;
}(React.Component);

exports.default = GoSync;
module.exports = exports['default'];