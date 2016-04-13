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

var MdGridOff = function (_React$Component) {
    _inherits(MdGridOff, _React$Component);

    function MdGridOff() {
        _classCallCheck(this, MdGridOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdGridOff).apply(this, arguments));
    }

    _createClass(MdGridOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 33.36h2.421666666666667l-2.421666666666667-2.421666666666667v2.421666666666667z m-3.280000000000001 0v-5.783333333333335l-0.9383333333333326-0.9366666666666674h-5.783333333333335v6.716666666666665h6.720000000000002z m-10-10v-5.783333333333335l-0.9383333333333326-0.9366666666666674h-5.783333333333334v6.716666666666669h6.7200000000000015z m0 10v-6.716666666666669h-6.716666666666668v6.716666666666669h6.716666666666668z m-6.719999999999999-22.421666666666667v2.421666666666667h2.421666666666667z m10 10v2.421666666666667h2.421666666666667z m-14.530000000000001-18.828333333333333l35.78333333333333 35.78333333333333-2.1133333333333297 2.106666666666669-3.3616666666666646-3.3583333333333343h-25.78q-1.3266666666666689 0-2.3033333333333346-0.9766666666666666t-0.9750000000000001-2.306666666666665v-25.781666666666666l-3.3600000000000008-3.3566666666666674z m24.53 4.530000000000001v6.716666666666669h6.716666666666669v-6.716666666666668h-6.716666666666669z m-13.280000000000001 0h-2.421666666666667l-3.3600000000000003-3.283333333333333h25.78333333333333q1.326666666666668 0 2.3033333333333346 0.9783333333333335t0.9750000000000014 2.3066666666666666v25.78333333333334l-3.2833333333333314-3.361666666666668v-2.421666666666667h-2.4200000000000017l-3.3599999999999994-3.283333333333335h5.783333333333331v-6.716666666666669h-6.719999999999999v5.783333333333335l-3.2833333333333314-3.361666666666668v-2.423333333333332h-2.4200000000000017l-3.3599999999999994-3.283333333333333h5.783333333333335v-6.716666666666668h-6.720000000000002v5.783333333333334l-3.283333333333333-3.3616666666666664v-2.421666666666667z' })
                )
            );
        }
    }]);

    return MdGridOff;
}(React.Component);

exports.default = MdGridOff;
module.exports = exports['default'];