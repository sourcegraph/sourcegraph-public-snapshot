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

var TiHeartHalfOutline = function (_React$Component) {
    _inherits(TiHeartHalfOutline, _React$Component);

    function TiHeartHalfOutline() {
        _classCallCheck(this, TiHeartHalfOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiHeartHalfOutline).apply(this, arguments));
    }

    _createClass(TiHeartHalfOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm3.7 15.7c0 2.1000000000000014 0.2999999999999998 5.5 3.3 8.5 2.6999999999999993 2.6000000000000014 11.5 8.599999999999998 11.8 9.000000000000004 0.3999999999999986 0.10000000000000142 0.6999999999999993 0.29999999999999716 1 0.29999999999999716s0.6999999999999993-0.20000000000000284 1-0.29999999999999716c0.3999999999999986-0.3999999999999986 9.2-6.199999999999999 11.900000000000002-9 3-3 3.299999999999997-6.399999999999999 3.299999999999997-8.5 0-5-4-9-9-9-2.6999999999999993 0-5.300000000000001 1.5-7 3.8000000000000007-1.6999999999999993-2.3000000000000007-4.300000000000001-3.8-7.300000000000001-3.8-4.9 0-9 3.999999999999999-9 9z m16.3 1.6000000000000014c1 0 1.6999999999999993-0.6000000000000014 1.6999999999999993-1.5999999999999996 0-3.200000000000001 2.5-5.700000000000001 5.599999999999998-5.700000000000001s5.700000000000003 2.5 5.700000000000003 5.699999999999999c0 1.8000000000000007-0.29999999999999716 4-2.5 6.099999999999998-2 2-8.2 6.399999999999999-10.5 7.899999999999999v-12.399999999999995z' })
                )
            );
        }
    }]);

    return TiHeartHalfOutline;
}(React.Component);

exports.default = TiHeartHalfOutline;
module.exports = exports['default'];