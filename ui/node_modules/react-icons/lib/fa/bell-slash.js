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

var FaBellSlash = function (_React$Component) {
    _inherits(FaBellSlash, _React$Component);

    function FaBellSlash() {
        _classCallCheck(this, FaBellSlash);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaBellSlash).apply(this, arguments));
    }

    _createClass(FaBellSlash, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30.43 19.14q1.1912500000000001 6.953749999999999 5.82 10.86 0 1.0124999999999993-0.7424999999999997 1.7562500000000014t-1.7575000000000003 0.7437499999999986h-8.75q0 2.0687500000000014-1.4649999999999999 3.5337499999999977t-3.535 1.4662500000000023-3.5249999999999986-1.4562499999999972-1.4749999999999996-3.5249999999999986z m-10.43 16.7975q0.3125 0 0.3125-0.3125t-0.3125-0.3125q-1.1524999999999999 0-1.9825000000000017-0.8299999999999983t-0.8299999999999983-1.9825000000000017q0-0.3125-0.3125-0.3125t-0.3125 0.3125q0 1.4249999999999972 1.0062500000000014 2.4312499999999986t2.4312499999999986 1.0062500000000014z m19.57-31.25q0.15625 0.19500000000000028 0.146250000000002 0.4587500000000002t-0.2049999999999983 0.4399999999999995l-36.56125 31.67875q-0.19625000000000314 0.15625-0.4600000000000031 0.13750000000000284t-0.41999999999999993-0.21625000000000227l-1.6412499999999999-1.875q-0.15625-0.19500000000000028-0.14625-0.458750000000002t0.20500000000000002-0.4200000000000017l3.6325000000000003-3.1449999999999996q-0.3700000000000001-0.6249999999999964-0.3700000000000001-1.2874999999999979 0.9749999999999996-0.8200000000000003 1.7750000000000004-1.71875t1.6624999999999996-2.333750000000002 1.4537499999999994-3.0962500000000013 0.9749999999999996-4.022500000000001 0.3837500000000009-5.078749999999996q0-2.96875 2.2837499999999995-5.5175t5.996250000000002-3.0962499999999995q-0.15500000000000114-0.37375000000000025-0.15500000000000114-0.7612500000000004 0-0.78125 0.5462500000000006-1.3275000000000001t1.3287499999999994-0.5474999999999999 1.3249999999999993 0.5474999999999999 0.5500000000000007 1.3275000000000001q0 0.3912500000000003-0.15749999999999886 0.7625000000000002 2.4212500000000006 0.34999999999999964 4.2775 1.6100000000000003t2.8900000000000006 3.075000000000001l8.162499999999998-7.0875q0.19624999999999915-0.1575000000000002 0.46000000000000085-0.13750000000000018t0.4200000000000017 0.21499999999999986z' })
                )
            );
        }
    }]);

    return FaBellSlash;
}(React.Component);

exports.default = FaBellSlash;
module.exports = exports['default'];