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

var TiBackspaceOutline = function (_React$Component) {
    _inherits(TiBackspaceOutline, _React$Component);

    function TiBackspaceOutline() {
        _classCallCheck(this, TiBackspaceOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiBackspaceOutline).apply(this, arguments));
    }

    _createClass(TiBackspaceOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.666666666666668 35h-16.666666666666668c-2.3933333333333344 0-5.241666666666667-1.4666666666666686-6.628333333333334-3.41l-4.366666666666667-6.111666666666668c-1.0466666666666669-1.4666666666666686-1.9266666666666667-2.6999999999999993-1.9783333333333335-2.7666666666666657-0.41000000000000014-0.5749999999999993-0.41666666666666674-1.4916666666666671-0.013333333333333197-2.0700000000000003l2.006666666666668-2.8083333333333336 4.346666666666667-6.091666666666665c1.3916666666666675-1.9416666666666664 4.243333333333332-3.408333333333333 6.633333333333333-3.408333333333333h16.666666666666668c2.756666666666664 0 5.0000000000000036 2.243333333333334 5.0000000000000036 5v16.666666666666664c0 2.7566666666666677-2.2433333333333323 5-5 5z m-26.283333333333335-13.333333333333332l1.3416666666666677 1.8733333333333348 4.363333333333334 6.111666666666668c0.7499999999999982 1.0566666666666649 2.616666666666667 2.014999999999997 3.9116666666666653 2.014999999999997h16.666666666666668c0.9216666666666633 0 1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679v-16.666666666666664c0-0.9166666666666679-0.7449999999999974-1.6666666666666679-1.6666666666666679-1.6666666666666679h-16.666666666666668c-1.291666666666666 0-3.16 0.9600000000000009-3.916666666666668 2.0166666666666657l-4.346666666666667 6.083333333333332-1.3533333333333317 1.9000000000000021z m17.46166666666667 0l4.41-4.41c0.32333333333333414-0.32333333333333414 0.32333333333333414-0.8533333333333317 0-1.1783333333333346-0.3249999999999993-0.32333333333333414-0.8550000000000004-0.32333333333333414-1.1783333333333346 0l-4.41 4.41-4.41-4.41c-0.3249999999999993-0.32333333333333414-0.8550000000000004-0.32333333333333414-1.1783333333333346 0-0.3249999999999993 0.3249999999999993-0.3249999999999993 0.8550000000000004 0 1.1783333333333346l4.41 4.41-4.41 4.41c-0.3249999999999993 0.3249999999999993-0.3249999999999993 0.8550000000000004 0 1.1783333333333346 0.1616666666666653 0.163333333333334 0.375 0.245000000000001 0.5883333333333347 0.245000000000001s0.42666666666666586-0.08333333333333215 0.5899999999999999-0.2433333333333323l4.41-4.411666666666669 4.41 4.41c0.163333333333334 0.163333333333334 0.37666666666666515 0.245000000000001 0.5899999999999999 0.245000000000001s0.42666666666666586-0.08333333333333215 0.5899999999999999-0.2433333333333323c0.32333333333333414-0.32333333333333414 0.32333333333333414-0.8533333333333317 0-1.1783333333333346l-4.411666666666665-4.411666666666669z' })
                )
            );
        }
    }]);

    return TiBackspaceOutline;
}(React.Component);

exports.default = TiBackspaceOutline;
module.exports = exports['default'];