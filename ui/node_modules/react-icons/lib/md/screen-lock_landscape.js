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

var MdScreenLockLandscape = function (_React$Component) {
    _inherits(MdScreenLockLandscape, _React$Component);

    function MdScreenLockLandscape() {
        _classCallCheck(this, MdScreenLockLandscape);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdScreenLockLandscape).apply(this, arguments));
    }

    _createClass(MdScreenLockLandscape, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.966666666666665 16.64v1.7166666666666686h4.066666666666666v-1.7166666666666686q0-0.7833333333333332-0.5883333333333347-1.3666666666666671t-1.4449999999999967-0.5866666666666678-1.4450000000000003 0.586666666666666-0.586666666666666 1.3666666666666654z m-1.326666666666668 10q-0.7033333333333331 0-1.1716666666666669-0.466666666666665t-0.46833333333333016-1.1733333333333356v-5q0-0.7033333333333331 0.4666666666666668-1.1716666666666669t1.1733333333333338-0.466666666666665v-1.7216666666666676q0-1.3283333333333331 0.9766666666666666-2.3049999999999997t2.383333333333333-0.9749999999999996 2.383333333333333 0.9383333333333326 0.9766666666666666 2.341666666666667v1.7166666666666686q0.7033333333333331 0 1.1716666666666669 0.46999999999999886t0.4683333333333337 1.173333333333332v5q0 0.7033333333333331-0.466666666666665 1.1716666666666669t-1.173333333333332 0.466666666666665h-6.716666666666669z m15 1.7199999999999989v-16.71666666666667h-23.28333333333333v16.71666666666667h23.283333333333335z m3.360000000000003-20q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3049999999999997v16.716666666666665q0 1.3300000000000018-1.0166666666666657 2.3066666666666684t-2.3433333333333337 0.975000000000005h-30q-1.3283333333333331 0-2.3433333333333333-0.9766666666666666t-1.0166666666666666-2.3066666666666684v-16.71333333333334q0-1.3299999999999983 1.0166666666666666-2.306666666666665t2.3433333333333333-0.9766666666666648h30z' })
                )
            );
        }
    }]);

    return MdScreenLockLandscape;
}(React.Component);

exports.default = MdScreenLockLandscape;
module.exports = exports['default'];