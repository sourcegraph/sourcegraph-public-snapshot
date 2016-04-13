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

var MdStayPrimaryLandscape = function (_React$Component) {
    _inherits(MdStayPrimaryLandscape, _React$Component);

    function MdStayPrimaryLandscape() {
        _classCallCheck(this, MdStayPrimaryLandscape);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdStayPrimaryLandscape).apply(this, arguments));
    }

    _createClass(MdStayPrimaryLandscape, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 11.64h-23.28333333333334v16.716666666666665h23.283333333333335v-16.714999999999996z m-29.921666666666667 0q0-1.3283333333333331 0.9783333333333335-2.3049999999999997t2.3033333333333292-0.9749999999999996h30q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3066666666666666v16.716666666666665q0 1.3299999999999983-1.0166666666666657 2.306666666666665t-2.3433333333333337 0.9733333333333398h-30q-1.3283333333333331 0-2.3433333333333333-0.9766666666666666t-1.0166666666666666-2.3049999999999997z' })
                )
            );
        }
    }]);

    return MdStayPrimaryLandscape;
}(React.Component);

exports.default = MdStayPrimaryLandscape;
module.exports = exports['default'];