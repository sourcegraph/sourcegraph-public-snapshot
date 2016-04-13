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

var MdNfc = function (_React$Component) {
    _inherits(MdNfc, _React$Component);

    function MdNfc() {
        _classCallCheck(this, MdNfc);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdNfc).apply(this, arguments));
    }

    _createClass(MdNfc, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 10v20h-20v-20h6.640000000000001v3.3599999999999994h-3.283333333333333v13.283333333333331h13.283333333333333v-13.283333333333333h-5v3.7500000000000018q1.7166666666666686 0.9383333333333326 1.7166666666666686 2.8900000000000006 0 1.3283333333333331-1.0133333333333319 2.3433333333333337t-2.3433333333333337 1.0166666666666657-2.3433333333333337-1.0166666666666657-1.0166666666666693-2.3433333333333337q0-1.9533333333333331 1.7199999999999989-2.8900000000000006v-3.75q0-1.3283333333333331 0.9766666666666666-2.3433333333333337t2.3049999999999997-1.0166666666666657h8.358333333333334z m3.3599999999999994 23.36v-26.716666666666665h-26.716666666666665v26.716666666666665h26.716666666666665z m0-30q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.3049999999999997v26.71666666666667q0 1.3299999999999983-0.9766666666666666 2.306666666666665t-2.306666666666665 0.9750000000000014h-26.713333333333335q-1.330000000000001 0-2.3066666666666675-0.9766666666666666t-0.9766666666666666-2.306666666666665v-26.713333333333335q0-1.330000000000001 0.9766666666666666-2.3066666666666675t2.3050000000000006-0.9766666666666666h26.71666666666667z' })
                )
            );
        }
    }]);

    return MdNfc;
}(React.Component);

exports.default = MdNfc;
module.exports = exports['default'];