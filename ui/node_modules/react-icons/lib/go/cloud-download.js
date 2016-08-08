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

var GoCloudDownload = function (_React$Component) {
    _inherits(GoCloudDownload, _React$Component);

    function GoCloudDownload() {
        _classCallCheck(this, GoCloudDownload);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoCloudDownload).apply(this, arguments));
    }

    _createClass(GoCloudDownload, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.5 12.5c-0.3412500000000023 0-0.6687499999999993 0.054999999999999716-1 0.09999999999999964-1.90625-4.465-6.337500000000002-7.6-11.5-7.6s-9.592500000000001 3.135-11.5 7.6c-0.33000000000000007-0.04499999999999993-0.65625-0.09999999999999964-1-0.09999999999999964-4.1425 0-7.5 3.3575-7.5 7.5s3.3575 7.5 7.5 7.5c0.8025000000000002 0 1.557500000000001-0.16625000000000156 2.280000000000001-0.40500000000000114 1.3312500000000007 1.4750000000000014 3.1624999999999996 2.4662499999999987 5.220000000000001 2.7687500000000007v-2.5474999999999994c-1.9474999999999998-0.40500000000000114-3.5775000000000006-1.6724999999999994-4.4-3.432500000000001-0.8537499999999998 0.6875-1.92 1.1187499999999986-3.0999999999999996 1.1187499999999986-2.76 0-5-2.240000000000002-5-5s2.24-5 5-5c0.9875000000000007 0 1.9000000000000004 0.29625000000000057 2.6724999999999994 0.7874999999999996 0.817499999999999-4.702499999999999 4.889999999999999-8.289999999999997 9.827499999999999-8.289999999999997 4.94125 0 8.975000000000001 3.60375 9.785 8.3125 0.78125-0.5087499999999991 1.7124999999999986-0.8125 2.715-0.8125 2.758749999999999 0 5 2.241250000000001 5 5s-2.241250000000001 5-5 5c-0.3999999999999986 0-0.78125-0.057500000000000995-1.1574999999999989-0.14499999999999957-1.129999999999999 1.5949999999999989-2.9875000000000007 2.6449999999999996-5.092500000000001 2.6449999999999996-0.432500000000001 0-0.8449999999999989-0.0625-1.25-0.15500000000000114v2.5375000000000014c0.40749999999999886 0.06500000000000128 0.8225000000000016 0.11874999999999858 1.25 0.11874999999999858 2.3900000000000006 0 4.550000000000001-0.9624999999999986 6.130000000000003-2.5162499999999994 0.04124999999999801 0.002500000000001279 0.07500000000000284 0.017499999999998295 0.11999999999999744 0.017499999999998295 4.142499999999998 0 7.5-3.3562499999999993 7.5-7.5s-3.3575000000000017-7.5-7.5-7.5z m-10 7.5h-5v12.5h-5l7.5 7.5 7.5-7.5h-5v-12.5z' })
                )
            );
        }
    }]);

    return GoCloudDownload;
}(React.Component);

exports.default = GoCloudDownload;
module.exports = exports['default'];