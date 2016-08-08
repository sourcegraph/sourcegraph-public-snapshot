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

var FaUserPlus = function (_React$Component) {
    _inherits(FaUserPlus, _React$Component);

    function FaUserPlus() {
        _classCallCheck(this, FaUserPlus);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaUserPlus).apply(this, arguments));
    }

    _createClass(FaUserPlus, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.75 20q-3.1050000000000004 0-5.3025-2.197499999999998t-2.1975-5.302500000000002 2.1975-5.3025 5.3025-2.1975 5.302500000000002 2.1975 2.197499999999998 5.3025-2.197499999999998 5.302500000000002-5.302500000000002 2.197499999999998z m18.75 2.5h6.875q0.2537499999999966 0 0.4399999999999977 0.18499999999999872t0.18500000000000227 0.4400000000000013v3.75q0 0.25375000000000014-0.18500000000000227 0.4400000000000013t-0.4399999999999977 0.18499999999999872h-6.875v6.875q0 0.2537499999999966-0.18500000000000227 0.4399999999999977t-0.4399999999999977 0.18500000000000227h-3.75q-0.25375000000000014 0-0.4400000000000013-0.18500000000000227t-0.18499999999999872-0.4399999999999977v-6.875h-6.875q-0.25375000000000014 0-0.4400000000000013-0.18499999999999872t-0.18499999999999872-0.4400000000000013v-3.75q0-0.25375000000000014 0.18499999999999872-0.4400000000000013t0.4400000000000013-0.18499999999999872h6.875v-6.875q0-0.25375000000000014 0.18499999999999872-0.4399999999999995t0.4400000000000013-0.1850000000000005h3.75q0.2537499999999966 0 0.4399999999999977 0.1850000000000005t0.18500000000000227 0.4399999999999995v6.875z m-14.375 4.375q0 1.0150000000000006 0.7424999999999997 1.7575000000000003t1.7575000000000003 0.7424999999999997h5v4.649999999999999q-1.3275000000000006 0.9750000000000014-3.34 0.9750000000000014h-17.07q-2.3625 0-3.7874999999999996-1.3474999999999966t-1.4275000000000002-3.712500000000002q0-1.0337500000000013 0.06875-2.0199999999999996t0.275-2.12875 0.51625-2.1187499999999986 0.8400000000000002-1.9037499999999987 1.2125-1.5824999999999996 1.6687499999999997-1.0462500000000006 2.1775-0.3900000000000041q0.37124999999999986 0 0.7625000000000002 0.3324999999999996 1.5412499999999998 1.1912500000000001 3.0162499999999994 1.7875000000000014t3.2125000000000004 0.5962500000000013 3.2124999999999986-0.5962500000000013 3.0174999999999983-1.7875000000000014q0.39124999999999943-0.3324999999999996 0.7624999999999993-0.3324999999999996 2.5775000000000006 0 4.237500000000001 1.875h-4.354999999999997q-1.0150000000000006 0-1.7575000000000003 0.7424999999999997t-0.7424999999999997 1.7575000000000003v3.75z' })
                )
            );
        }
    }]);

    return FaUserPlus;
}(React.Component);

exports.default = FaUserPlus;
module.exports = exports['default'];