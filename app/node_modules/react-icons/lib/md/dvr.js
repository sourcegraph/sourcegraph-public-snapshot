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

var MdDvr = function (_React$Component) {
    _inherits(MdDvr, _React$Component);

    function MdDvr() {
        _classCallCheck(this, MdDvr);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDvr).apply(this, arguments));
    }

    _createClass(MdDvr, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11.64 20v3.3599999999999994h-3.283333333333333v-3.3599999999999994h3.283333333333333z m0-6.640000000000001v3.2833333333333314h-3.283333333333333v-3.283333333333333h3.283333333333333z m20 6.640000000000001v3.3599999999999994h-18.283333333333335v-3.3599999999999994h18.283333333333335z m0-6.640000000000001v3.2833333333333314h-18.283333333333335v-3.283333333333333h18.283333333333335z m3.3599999999999994 15v-20h-30v20h30z m0-23.36q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3416666666666677l-0.0799999999999983 20q0 1.3283333333333331-0.9766666666666666 2.3049999999999997t-2.3033333333333346 0.9766666666666701h-8.361666666666668v3.359999999999996h-13.283333333333333v-3.3599999999999994h-8.354999999999999q-1.3283333333333331 0-2.3433333333333333-0.9766666666666666t-1.0166666666666666-2.3049999999999997v-20q0-1.3283333333333331 1.0166666666666666-2.3433333333333337t2.3400000000000003-1.0150000000000006h30.000000000000004z' })
                )
            );
        }
    }]);

    return MdDvr;
}(React.Component);

exports.default = MdDvr;
module.exports = exports['default'];