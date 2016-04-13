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

var FaAutomobile = function (_React$Component) {
    _inherits(FaAutomobile, _React$Component);

    function FaAutomobile() {
        _classCallCheck(this, FaAutomobile);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaAutomobile).apply(this, arguments));
    }

    _createClass(FaAutomobile, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm9.375 23.75q0-1.2875000000000014-0.9175000000000004-2.2074999999999996t-2.2074999999999996-0.9175000000000004-2.2074999999999996 0.9175000000000004-0.9175000000000004 2.2074999999999996 0.9175000000000004 2.2074999999999996 2.2074999999999996 0.9175000000000004 2.2074999999999996-0.9175000000000004 0.9175000000000004-2.2074999999999996z m0.7025000000000006-6.25h19.84375l-1.7375000000000007-6.9725q-0.03999999999999915-0.15625-0.2749999999999986-0.3412500000000005t-0.4087500000000013-0.18624999999999936h-15q-0.1775000000000002 0-0.41249999999999964 0.1875t-0.27250000000000085 0.33999999999999986z m26.7975 6.25q0-1.2875000000000014-0.9174999999999969-2.2074999999999996t-2.207500000000003-0.9175000000000004-2.2074999999999996 0.9175000000000004-0.9175000000000004 2.2074999999999996 0.9175000000000004 2.2074999999999996 2.2074999999999996 0.9175000000000004 2.207500000000003-0.9175000000000004 0.9174999999999969-2.2074999999999996z m3.125-1.875v7.5q0 0.2749999999999986-0.17499999999999716 0.4499999999999993t-0.45000000000000284 0.1750000000000007h-1.875v2.5q0 1.5625-1.09375 2.65625t-2.65625 1.09375-2.65625-1.09375-1.09375-2.65625v-2.5h-20v2.5q0 1.5625-1.09375 2.65625t-2.65625 1.09375-2.65625-1.09375-1.09375-2.65625v-2.5h-1.875q-0.275 0-0.44999999999999996-0.1750000000000007t-0.17500000000000004-0.4499999999999993v-7.5q0-1.8162500000000001 1.2787499999999998-3.0962500000000013t3.0962500000000004-1.2787499999999987h0.5475000000000003l2.05-8.18375q0.4500000000000002-1.8375000000000004 2.032499999999999-3.075t3.495000000000001-1.24125h15q1.9149999999999991 0 3.4974999999999987 1.2400000000000002t2.0312500000000036 3.075000000000001l2.049999999999997 8.184999999999999h0.5499999999999972q1.8149999999999977 0 3.094999999999999 1.2787499999999987t1.2762500000000045 3.0962500000000013z' })
                )
            );
        }
    }]);

    return FaAutomobile;
}(React.Component);

exports.default = FaAutomobile;
module.exports = exports['default'];