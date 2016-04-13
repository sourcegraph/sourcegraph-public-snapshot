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

var MdPrint = function (_React$Component) {
    _inherits(MdPrint, _React$Component);

    function MdPrint() {
        _classCallCheck(this, MdPrint);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPrint).apply(this, arguments));
    }

    _createClass(MdPrint, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 5v6.640000000000001h-20v-6.640000000000001h20z m1.6400000000000006 15q0.7033333333333331 0 1.211666666666666-0.466666666666665t0.509999999999998-1.173333333333332-0.5083333333333329-1.211666666666666-1.2100000000000009-0.5100000000000016-1.1716666666666669 0.5083333333333329-0.466666666666665 1.2100000000000009 0.466666666666665 1.1716666666666669 1.173333333333332 0.466666666666665z m-5 11.64v-8.283333333333331h-13.283333333333333v8.283333333333331h13.283333333333333z m5-18.28q2.0333333333333314 0 3.5166666666666657 1.4833333333333343t1.4833333333333343 3.5166666666666657v10h-6.640000000000001v6.640000000000001h-20v-6.640000000000001h-6.64v-10q0-2.0333333333333314 1.4833333333333334-3.5166666666666657t3.5166666666666657-1.4833333333333343h23.283333333333335z' })
                )
            );
        }
    }]);

    return MdPrint;
}(React.Component);

exports.default = MdPrint;
module.exports = exports['default'];