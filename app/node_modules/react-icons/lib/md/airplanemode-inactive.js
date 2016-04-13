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

var MdAirplanemodeInactive = function (_React$Component) {
    _inherits(MdAirplanemodeInactive, _React$Component);

    function MdAirplanemodeInactive() {
        _classCallCheck(this, MdAirplanemodeInactive);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAirplanemodeInactive).apply(this, arguments));
    }

    _createClass(MdAirplanemodeInactive, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm5 8.75l2.1100000000000003-2.1100000000000003 26.25 26.25-2.1099999999999994 2.1099999999999994-9.61-9.533333333333331v6.173333333333332l3.3599999999999994 2.5v2.5l-5.859999999999999-1.6400000000000006-5.783333333333335 1.6400000000000006v-2.5l3.283333333333335-2.5v-9.14l-13.283333333333333 4.140000000000001v-3.2833333333333314l9.923333333333334-6.25z m16.64 6.25l13.36 8.36v3.2833333333333314l-5.313333333333329-1.6433333333333309-13.04666666666667-13.043333333333333v-6.095000000000001q0-1.0166666666666666 0.7416666666666671-1.7583333333333329t1.7566666666666677-0.7416666666666667 1.7583333333333329 0.7416666666666667 0.7399999999999984 1.7583333333333329v9.138333333333334z' })
                )
            );
        }
    }]);

    return MdAirplanemodeInactive;
}(React.Component);

exports.default = MdAirplanemodeInactive;
module.exports = exports['default'];