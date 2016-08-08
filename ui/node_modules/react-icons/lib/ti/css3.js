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

var TiCss3 = function (_React$Component) {
    _inherits(TiCss3, _React$Component);

    function TiCss3() {
        _classCallCheck(this, TiCss3);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiCss3).apply(this, arguments));
    }

    _createClass(TiCss3, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm9.5 5.7l-1 5.3h20.5l-0.6999999999999993 3.5h-20.5l-1.0000000000000009 5.300000000000001h20.5l-1.1000000000000014 6-8.400000000000002 2.8999999999999986-7.1-2.8999999999999986 0.5-2.6000000000000014h-5l-1.1999999999999975 6.300000000000001 11.8 4.799999999999997 13.7-4.800000000000001 1.7999999999999972-9.7 0.3999999999999986-2 2.3000000000000043-12.099999999999998h-25.5z' })
                )
            );
        }
    }]);

    return TiCss3;
}(React.Component);

exports.default = TiCss3;
module.exports = exports['default'];