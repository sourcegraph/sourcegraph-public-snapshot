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

var MdSupervisorAccount = function (_React$Component) {
    _inherits(MdSupervisorAccount, _React$Component);

    function MdSupervisorAccount() {
        _classCallCheck(this, MdSupervisorAccount);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSupervisorAccount).apply(this, arguments));
    }

    _createClass(MdSupervisorAccount, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm15 21.64q1.7166666666666686 0 3.9833333333333343 0.466666666666665-3.9833333333333343 2.193333333333335-3.9833333333333343 5.785v3.75h-11.64v-4.141666666666666q8.881784197001252e-16-1.7966666666666669 2.1500000000000012-3.203333333333333t4.7250000000000005-2.0333333333333314 4.764999999999999-0.6216666666666697z m12.5 1.7199999999999989q2.8900000000000006 0 6.016666666666666 1.25t3.123333333333335 3.2833333333333314v3.75h-18.283333333333335v-3.75q0-2.0333333333333314 3.126666666666665-3.2833333333333314t6.016666666666666-1.25z m-12.5-5q-2.033333333333333 0-3.5166666666666657-1.4833333333333343t-1.4833333333333343-3.518333333333331 1.4833333333333343-3.5166666666666657 3.5166666666666657-1.4800000000000004 3.5166666666666657 1.4833333333333343 1.4833333333333343 3.5166666666666657-1.4833333333333343 3.5166666666666657-3.5166666666666657 1.4833333333333343z m12.5 1.6400000000000006q-1.7166666666666686 0-2.9299999999999997-1.211666666666666t-1.211666666666666-2.9299999999999997 1.211666666666666-2.966666666666667 2.9299999999999997-1.25 2.9299999999999997 1.25 1.211666666666666 2.966666666666667-1.211666666666666 2.9299999999999997-2.9299999999999997 1.211666666666666z' })
                )
            );
        }
    }]);

    return MdSupervisorAccount;
}(React.Component);

exports.default = MdSupervisorAccount;
module.exports = exports['default'];