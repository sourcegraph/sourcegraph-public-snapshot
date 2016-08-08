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

var MdShare = function (_React$Component) {
    _inherits(MdShare, _React$Component);

    function MdShare() {
        _classCallCheck(this, MdShare);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdShare).apply(this, arguments));
    }

    _createClass(MdShare, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 26.796666666666667q2.0333333333333314 0 3.4383333333333326 1.4450000000000003t1.4066666666666663 3.3999999999999986q0 2.030000000000001-1.4450000000000003 3.4750000000000014t-3.3999999999999986 1.4450000000000003-3.3966666666666683-1.4450000000000003-1.4450000000000003-3.4766666666666666q0-0.783333333333335 0.07833333333333314-1.0933333333333337l-11.796666666666667-6.875q-1.4833333333333307 1.3283333333333331-3.4399999999999977 1.3283333333333331-2.0300000000000002 0-3.5133333333333336-1.4833333333333343t-1.4833333333333334-3.5166666666666657 1.4833333333333334-3.5166666666666657 3.5166666666666675-1.4833333333333343q1.9499999999999993 0 3.4366666666666674 1.3283333333333331l11.716666666666665-6.796666666666667q-0.15500000000000114-0.7833333333333332-0.15500000000000114-1.1733333333333338 0-2.033333333333333 1.4866666666666681-3.5166666666666666t3.5166666666666657-1.4833333333333334 3.513333333333332 1.4833333333333334 1.4833333333333343 3.5166666666666666-1.4833333333333343 3.5166666666666675-3.5166666666666657 1.4833333333333343q-1.8733333333333348 0-3.4366666666666674-1.4066666666666663l-11.716666666666667 6.875q0.15166666666666906 0.7850000000000001 0.15166666666666906 1.173333333333332t-0.15499999999999936 1.1716666666666669l11.874999999999998 6.875q1.4066666666666663-1.25 3.2833333333333314-1.25z' })
                )
            );
        }
    }]);

    return MdShare;
}(React.Component);

exports.default = MdShare;
module.exports = exports['default'];