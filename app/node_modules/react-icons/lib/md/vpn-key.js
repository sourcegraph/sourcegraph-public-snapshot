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

var MdVpnKey = function (_React$Component) {
    _inherits(MdVpnKey, _React$Component);

    function MdVpnKey() {
        _classCallCheck(this, MdVpnKey);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdVpnKey).apply(this, arguments));
    }

    _createClass(MdVpnKey, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11.64 23.36q1.3283333333333331 0 2.3433333333333337-1.0166666666666657t1.0166666666666657-2.34-1.0166666666666657-2.341666666666665-2.3433333333333337-1.0166666666666657-2.3049999999999997 1.0166666666666657-0.9749999999999996 2.338333333333331 0.9766666666666666 2.344999999999999 2.3066666666666666 1.0166666666666657z m9.453333333333333-6.719999999999999h17.266666666666666v6.716666666666669h-3.3616666666666646v6.643333333333331h-6.641666666666666v-6.641666666666666h-7.266666666666666q-1.0133333333333319 2.9666666666666686-3.591666666666665 4.805t-5.858333333333338 1.836666666666666q-4.140000000000001 0-7.07-2.929999999999996t-2.93-7.070000000000004 2.93-7.07 7.07-2.9299999999999997q3.283333333333333 0 5.859999999999999 1.836666666666666t3.5933333333333337 4.805z' })
                )
            );
        }
    }]);

    return MdVpnKey;
}(React.Component);

exports.default = MdVpnKey;
module.exports = exports['default'];