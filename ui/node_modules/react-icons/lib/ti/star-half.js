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

var TiStarHalf = function (_React$Component) {
    _inherits(TiStarHalf, _React$Component);

    function TiStarHalf() {
        _classCallCheck(this, TiStarHalf);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiStarHalf).apply(this, arguments));
    }

    _createClass(TiStarHalf, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm19.2 7.2c-1.5 3.1000000000000005-3.6999999999999993 7.999999999999999-3.6999999999999993 7.999999999999999s-5.199999999999999 0.5999999999999996-8.7 1c-0.2999999999999998 0-0.5999999999999996 0.3000000000000007-0.5999999999999996 0.5-0.20000000000000018 0.3000000000000007 0 0.6000000000000014 0.09999999999999964 0.8000000000000007 2.6999999999999993 2.3000000000000007 6.5 6 6.5 6s-1 5.199999999999999-1.8000000000000007 8.700000000000003c0 0.29999999999999716 0 0.6000000000000014 0.3000000000000007 0.7999999999999972 0.40000000000000036 0.29999999999999716 0.6999999999999993 0.29999999999999716 1 0.20000000000000284 3-1.6999999999999993 7.699999999999999-4.399999999999999 7.699999999999999-4.399999999999999v-22.100000000000005c-0.3000000000000007 8.881784197001252e-16-0.6999999999999993 0.3000000000000007-0.8000000000000007 0.5000000000000009z' })
                )
            );
        }
    }]);

    return TiStarHalf;
}(React.Component);

exports.default = TiStarHalf;
module.exports = exports['default'];