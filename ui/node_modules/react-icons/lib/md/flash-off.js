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

var MdFlashOff = function (_React$Component) {
    _inherits(MdFlashOff, _React$Component);

    function MdFlashOff() {
        _classCallCheck(this, MdFlashOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFlashOff).apply(this, arguments));
    }

    _createClass(MdFlashOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 16.64l-2.578333333333333 4.453333333333333-14.141666666666666-14.14v-3.593333333333333h16.716666666666665l-6.716666666666669 13.283333333333335h6.716666666666669z m-22.89333333333333-11.64l26.173333333333332 26.25-2.1099999999999994 2.1099999999999994-6.875-6.953333333333333-6.016666666666666 10.233333333333334v-15h-5v-6.168333333333335l-8.278333333333336-8.361666666666665z' })
                )
            );
        }
    }]);

    return MdFlashOff;
}(React.Component);

exports.default = MdFlashOff;
module.exports = exports['default'];