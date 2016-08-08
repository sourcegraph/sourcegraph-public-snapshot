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

var MdTextFormat = function (_React$Component) {
    _inherits(MdTextFormat, _React$Component);

    function MdTextFormat() {
        _classCallCheck(this, MdTextFormat);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdTextFormat).apply(this, arguments));
    }

    _createClass(MdTextFormat, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 10l-3.125 8.36h6.25z m-4.140000000000001 11.328333333333333l-1.5633333333333326 3.671666666666667h-3.4383333333333344l7.8916666666666675-18.36h2.5l7.886666666666667 18.36h-3.4366666666666674l-1.5633333333333326-3.671666666666667h-8.283333333333333z m-7.5 7.033333333333331h23.283333333333335v3.280000000000001h-23.285000000000004v-3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdTextFormat;
}(React.Component);

exports.default = MdTextFormat;
module.exports = exports['default'];