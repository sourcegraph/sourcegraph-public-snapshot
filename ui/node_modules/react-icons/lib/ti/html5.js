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

var TiHtml5 = function (_React$Component) {
    _inherits(TiHtml5, _React$Component);

    function TiHtml5() {
        _classCallCheck(this, TiHtml5);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiHtml5).apply(this, arguments));
    }

    _createClass(TiHtml5, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.8 5.8l1.1999999999999993 1.9000000000000004 1.1999999999999993-1.8999999999999995v2.5h1.6000000000000014v-5h-1.6000000000000014l-1.1999999999999993 1.8999999999999995-1-1.9h-1.8000000000000007v5h1.6000000000000014z m8.900000000000002 2.500000000000001v-1.6000000000000005h-2.3999999999999986v-3.4000000000000004h-1.6000000000000014v5.000000000000001z m-14.399999999999999 0h1.6999999999999957v-3.3000000000000007h1.5v-1.7000000000000002h-4.699999999999999v1.7000000000000002h1.5z m-5.300000000000004-1.6000000000000005h1.5v1.6000000000000005h1.6999999999999993v-5h-1.6999999999999993v1.6999999999999993h-1.5v-1.7000000000000002h-1.6999999999999993v5.000000000000001h1.6999999999999993z m-2.6999999999999993 3.3l2 24 9.7 2.700000000000003 9.7-2.700000000000003 2-24h-23.4z m18.9 7.699999999999999h-11.399999999999999l0.3999999999999986 3h10.599999999999998l-0.8000000000000007 9.099999999999998-6 1.6999999999999993-6-1.6999999999999993-0.5-4.800000000000001h3l0.1999999999999993 2.5 3.3000000000000007 0.8000000000000007 3.3000000000000007-0.8000000000000007 0.3999999999999986-3.8000000000000007h-10.399999999999995l-0.8000000000000007-8.9h15l-0.3000000000000007 2.9000000000000004z' })
                )
            );
        }
    }]);

    return TiHtml5;
}(React.Component);

exports.default = TiHtml5;
module.exports = exports['default'];