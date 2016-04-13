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

var MdPregnantWoman = function (_React$Component) {
    _inherits(MdPregnantWoman, _React$Component);

    function MdPregnantWoman() {
        _classCallCheck(this, MdPregnantWoman);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPregnantWoman).apply(this, arguments));
    }

    _createClass(MdPregnantWoman, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 21.64v6.716666666666669h-5v8.283333333333331h-5v-8.283333333333331h-3.283333333333333v-11.716666666666669q0-2.033333333333333 1.4866666666666664-3.5166666666666657t3.5166666666666657-1.4833333333333343 3.5133333333333354 1.4833333333333343 1.4833333333333343 3.5166666666666657q3.2833333333333314 1.3283333333333331 3.2833333333333314 5z m-11.64-15q0-1.4066666666666663 0.9766666666666666-2.3433333333333337t2.383333333333333-0.9383333333333335 2.3433333333333337 0.9383333333333335 0.9383333333333326 2.3433333333333337-0.9383333333333326 2.383333333333333-2.3433333333333337 0.9766666666666666-2.383333333333333-0.9766666666666666-0.9766666666666666-2.383333333333333z' })
                )
            );
        }
    }]);

    return MdPregnantWoman;
}(React.Component);

exports.default = MdPregnantWoman;
module.exports = exports['default'];