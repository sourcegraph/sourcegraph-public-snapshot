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

var GoLightBulb = function (_React$Component) {
    _inherits(GoLightBulb, _React$Component);

    function GoLightBulb() {
        _classCallCheck(this, GoLightBulb);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoLightBulb).apply(this, arguments));
    }

    _createClass(GoLightBulb, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 2.5c-6.903749999999999 0-12.5 5.596250000000001-12.5 12.5 0 4.09 1.9625000000000004 7.719999999999999 5 10v7.5c0 1.3812500000000014 1.1187500000000004 2.5 2.5 2.5 0 1.3812500000000014 1.1187499999999986 2.5 2.5 2.5h5c1.3812500000000014 0 2.5-1.1187499999999986 2.5-2.5 1.3812500000000014 0 2.5-1.1187499999999986 2.5-2.5v-7.5c3.0375000000000014-2.280000000000001 5-5.912500000000001 5-10 0-6.903749999999999-5.596250000000001-12.5-12.5-12.5z m5 28.75c0 0.6900000000000013-0.5599999999999987 1.25-1.25 1.25h-7.5c-0.6899999999999995 0-1.25-0.5599999999999987-1.25-1.25v-1.25h10v1.25z m2.5-9.6375c-1.3000000000000007 1.2912500000000016-2.5 1.6574999999999989-2.5 4.846250000000001v1.041249999999998h-2.5v-5l5-5v-2.5l-2.5-2.5-2.5 2.5-2.5-2.5-2.5 2.5-2.5-2.5-2.5 2.5v2.5l5 5v5h-2.5v-1.0412500000000016c0-3.1900000000000013-1.1999999999999993-3.5549999999999997-2.5-4.844999999999999-1.5549999999999997-1.7624999999999993-2.5-4.074999999999999-2.5-6.612500000000001 0-5.522500000000001 4.477499999999999-10 10-10s10 4.477499999999999 10 10c0 2.5375000000000014-0.9450000000000003 4.850000000000001-2.5 6.612500000000001z m-7.5 0.8874999999999993l-5-5v-2.5l2.5 2.5 2.5-2.5 2.5 2.5 2.5-2.5v2.5l-5 5z' })
                )
            );
        }
    }]);

    return GoLightBulb;
}(React.Component);

exports.default = GoLightBulb;
module.exports = exports['default'];