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

var MdFlashAuto = function (_React$Component) {
    _inherits(MdFlashAuto, _React$Component);

    function MdFlashAuto() {
        _classCallCheck(this, MdFlashAuto);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFlashAuto).apply(this, arguments));
    }

    _createClass(MdFlashAuto, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.046666666666667 12.733333333333333h3.9066666666666663l-1.9533333333333331-6.089999999999999z m3.5933333333333337-9.373333333333333l5.390000000000001 15h-3.203333333333333l-1.1716666666666669-3.3599999999999994h-5.313333333333333l-1.1716666666666669 3.3599999999999994h-3.2033333333333367l5.390000000000001-15h3.283333333333335z m-26.64 8.881784197001252e-16h16.64l-6.640000000000001 15h6.640000000000001l-11.64 20v-15h-5v-20z' })
                )
            );
        }
    }]);

    return MdFlashAuto;
}(React.Component);

exports.default = MdFlashAuto;
module.exports = exports['default'];