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

var MdPhoneIphone = function (_React$Component) {
    _inherits(MdPhoneIphone, _React$Component);

    function MdPhoneIphone() {
        _classCallCheck(this, MdPhoneIphone);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPhoneIphone).apply(this, arguments));
    }

    _createClass(MdPhoneIphone, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 30v-23.36h-15v23.36h15z m-7.5 6.640000000000001q1.0166666666666657 0 1.7583333333333329-0.7416666666666671t0.7399999999999984-1.759999999999998-0.7416666666666671-1.7583333333333329-1.7600000000000016-0.7433333333333323-1.7583333333333329 0.7416666666666671-0.7433333333333323 1.7566666666666677 0.7416666666666671 1.7583333333333329 1.7566666666666677 0.740000000000002z m6.719999999999999-35q1.7166666666666686 0 2.9299999999999997 1.25t1.2100000000000009 2.966666666666666v28.283333333333335q0 1.7166666666666686-1.2100000000000009 2.9666666666666686t-2.9299999999999997 1.25h-13.36q-1.7166666666666668 0-2.9299999999999997-1.25t-1.211666666666666-2.9666666666666686v-28.283333333333335q0-1.716666666666666 1.211666666666666-2.966666666666666t2.9299999999999997-1.2499999999999996h13.36z' })
                )
            );
        }
    }]);

    return MdPhoneIphone;
}(React.Component);

exports.default = MdPhoneIphone;
module.exports = exports['default'];