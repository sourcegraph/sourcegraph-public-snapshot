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

var MdTune = function (_React$Component) {
    _inherits(MdTune, _React$Component);

    function MdTune() {
        _classCallCheck(this, MdTune);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdTune).apply(this, arguments));
    }

    _createClass(MdTune, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 15v-10h3.3599999999999994v3.3599999999999994h6.640000000000001v3.283333333333333h-6.640000000000001v3.3566666666666674h-3.3599999999999994z m10 6.640000000000001h-16.64v-3.2833333333333314h16.64v3.2833333333333314z m-23.36-6.640000000000001h3.3599999999999994v10h-3.3599999999999994v-3.3599999999999994h-6.640000000000001v-3.2833333333333314h6.640000000000001v-3.356666666666669z m10 20h-3.2833333333333314v-10h3.2833333333333314v3.3599999999999994h13.36v3.2833333333333314h-13.36v3.356666666666669z m-16.64-26.64h16.64v3.283333333333335h-16.64v-3.283333333333333z m0 20h10v3.2833333333333314h-10v-3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdTune;
}(React.Component);

exports.default = MdTune;
module.exports = exports['default'];