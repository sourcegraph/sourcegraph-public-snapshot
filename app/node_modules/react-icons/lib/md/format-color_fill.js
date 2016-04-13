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

var MdFormatColorFill = function (_React$Component) {
    _inherits(MdFormatColorFill, _React$Component);

    function MdFormatColorFill() {
        _classCallCheck(this, MdFormatColorFill);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFormatColorFill).apply(this, arguments));
    }

    _createClass(MdFormatColorFill, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm0 33.36h40v6.640000000000001h-40v-6.640000000000001z m31.640000000000004-14.219999999999999q3.359999999999996 3.671666666666667 3.359999999999996 5.859999999999999 0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657-2.3049999999999997-1.0166666666666657-0.9766666666666666-2.3433333333333337q0-0.9383333333333326 0.8200000000000003-2.421666666666667t1.6000000000000014-2.421666666666667z m-22.966666666666665-2.5h16.013333333333335l-8.046666666666667-7.966666666666669z m18.905-1.7166666666666668q0.783333333333335 0.7799999999999994 0.783333333333335 1.795t-0.783333333333335 1.7199999999999989l-9.14 9.14q-0.783333333333335 0.783333333333335-1.7966666666666669 0.783333333333335-0.9383333333333326 0-1.7166666666666668-0.783333333333335l-9.22-9.14q-0.7833333333333332-0.7033333333333331-0.7833333333333332-1.7166666666666686t0.7833333333333332-1.8000000000000007l8.594999999999995-8.591666666666663-3.9866666666666664-3.983333333333334 2.419999999999998-2.3466666666666662z' })
                )
            );
        }
    }]);

    return MdFormatColorFill;
}(React.Component);

exports.default = MdFormatColorFill;
module.exports = exports['default'];