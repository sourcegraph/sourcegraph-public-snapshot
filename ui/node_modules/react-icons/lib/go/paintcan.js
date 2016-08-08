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

var GoPaintcan = function (_React$Component) {
    _inherits(GoPaintcan, _React$Component);

    function GoPaintcan() {
        _classCallCheck(this, GoPaintcan);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoPaintcan).apply(this, arguments));
    }

    _createClass(GoPaintcan, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 0c-8.283749999999998 0-15 6.7162500000000005-15 15v2.5c0 1.3812500000000014 1.1187500000000004 2.5 2.5 2.5v12.5c0 2.762500000000003 5.596250000000001 5 12.5 5s12.5-2.237499999999997 12.5-5v-12.5c1.3812500000000014 0 2.5-1.1187499999999986 2.5-2.5v-2.5c0-8.28375-6.716250000000002-15-15-15z m7.5 25v1.25c0 0.6900000000000013-0.5599999999999987 1.25-1.25 1.25s-1.25-0.5599999999999987-1.25-1.25v-1.25c0-0.6900000000000013-0.5599999999999987-1.25-1.25-1.25s-1.25 0.5599999999999987-1.25 1.25v6.25c0 0.6900000000000013-0.5599999999999987 1.25-1.25 1.25s-1.25-0.5599999999999987-1.25-1.25v-5c0-0.6900000000000013-0.5599999999999987-1.25-1.25-1.25s-1.25 0.5599999999999987-1.25 1.25v1.25c0 1.3812500000000014-1.1187499999999986 2.5-2.5 2.5s-2.5-1.1187499999999986-2.5-2.5v-2.5c-1.3812499999999996 0-2.5-1.1187499999999986-2.5-2.5v-4.5c2.280000000000001 1.2124999999999986 5.912500000000001 2 10 2s7.719999999999999-0.7850000000000001 10-2v4.5c0 1.3812500000000014-1.1187499999999986 2.5-2.5 2.5z m-7.5-7.5c-4.196249999999999 0-7.787500000000001-1.0337500000000013-9.2725-2.5 1.4837500000000006-1.4662500000000005 5.074999999999999-2.5 9.2725-2.5s7.787500000000001 1.0337499999999995 9.2725 2.5c-1.4837500000000006 1.4662499999999987-5.074999999999999 2.5-9.2725 2.5z m0-7.5c-6.899999999999999 0-12.4925 2.237499999999999-12.5 4.995000000000001 0.0024999999999995026-6.9 5.600000000000001-12.495000000000001 12.5-12.495000000000001 6.903749999999999 0 12.5 5.596250000000001 12.5 12.5 0-2.7624999999999993-5.596250000000001-5-12.5-5z' })
                )
            );
        }
    }]);

    return GoPaintcan;
}(React.Component);

exports.default = GoPaintcan;
module.exports = exports['default'];