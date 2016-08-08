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

var MdCameraFront = function (_React$Component) {
    _inherits(MdCameraFront, _React$Component);

    function MdCameraFront() {
        _classCallCheck(this, MdCameraFront);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCameraFront).apply(this, arguments));
    }

    _createClass(MdCameraFront, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11.64 3.3600000000000003v17.5q0-1.875 2.8499999999999996-3.046666666666667t5.51-1.1716666666666669 5.508333333333333 1.1716666666666669 2.8500000000000014 3.046666666666667v-17.5h-16.715000000000003z m16.72-3.3600000000000003q1.3283333333333331 0 2.3049999999999997 1.0166666666666666t0.9750000000000014 2.341666666666667v23.283333333333335q0 1.326666666666668-0.9766666666666666 2.341666666666665t-2.306666666666665 1.0166666666666657h-11.716666666666669l5 5-5 5v-3.3616666666666646h-8.283333333333333v-3.2833333333333314h8.283333333333333v-3.355000000000004h-5q-1.33 0-2.3066666666666666-1.0166666666666657t-0.9766666666666666-2.3416666666666686v-23.286666666666665q0-1.326666666666667 0.9766666666666666-2.3416666666666672t2.3049999999999997-1.0166666666666666h16.71666666666666z m-8.36 13.360000000000001q-1.3283333333333331 0-2.3049999999999997-1.0166666666666675t-0.9766666666666666-2.341666666666667 0.9783333333333317-2.341666666666667 2.3033333333333346-1.0183333333333335 2.3433333333333337 1.0166666666666666 1.0166666666666657 2.3433333333333337-1.0166666666666657 2.3450000000000006-2.3433333333333337 1.0166666666666657z m3.3599999999999994 20h8.283333333333331v3.2833333333333314h-8.283333333333331v-3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdCameraFront;
}(React.Component);

exports.default = MdCameraFront;
module.exports = exports['default'];