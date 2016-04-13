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

var FaDiamond = function (_React$Component) {
    _inherits(FaDiamond, _React$Component);

    function FaDiamond() {
        _classCallCheck(this, FaDiamond);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaDiamond).apply(this, arguments));
    }

    _createClass(FaDiamond, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm4.14125 17.5l12.1675 12.9875-5.858750000000001-12.9875h-6.312499999999999z m15.85875 15.07875l6.81625-15.07875h-13.6325z m-9.4925-17.57875l3.9837500000000006-7.5h-5.116250000000001l-5.625 7.5h6.756250000000001z m13.18375 15.4875l12.1675-12.9875h-6.30875z m-10.35-15.4875h13.31875l-3.9849999999999994-7.5h-5.350000000000001z m16.15 0h6.758749999999999l-5.625-7.5h-5.1175z m2.7550000000000026-9.4925l7.5 10q0.2749999999999986 0.34999999999999964 0.2537499999999966 0.8099999999999987t-0.3325000000000031 0.7912500000000016l-18.75 20q-0.3499999999999943 0.39124999999999943-0.9174999999999969 0.39124999999999943t-0.9175000000000004-0.39124999999999943l-18.75-20q-0.3125-0.3324999999999996-0.3325-0.7912500000000016t0.25375000000000003-0.8100000000000005l7.5-10q0.35000000000000053-0.5074999999999985 0.9962500000000007-0.5074999999999985h22.5q0.6449999999999996 0 0.9962500000000034 0.5075000000000003z' })
                )
            );
        }
    }]);

    return FaDiamond;
}(React.Component);

exports.default = FaDiamond;
module.exports = exports['default'];