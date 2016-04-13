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

var MdKeyboard = function (_React$Component) {
    _inherits(MdKeyboard, _React$Component);

    function MdKeyboard() {
        _classCallCheck(this, MdKeyboard);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdKeyboard).apply(this, arguments));
    }

    _createClass(MdKeyboard, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 16.64v-3.283333333333333h-3.283333333333335v3.283333333333333h3.2833333333333314z m0 5v-3.2833333333333314h-3.283333333333335v3.2833333333333314h3.2833333333333314z m-5-5v-3.283333333333333h-3.283333333333335v3.283333333333333h3.2833333333333314z m0 5v-3.2833333333333314h-3.283333333333335v3.2833333333333314h3.2833333333333314z m0 6.719999999999999v-3.3599999999999994h-13.283333333333333v3.3599999999999994h13.283333333333333z m-15-11.719999999999999v-3.283333333333333h-3.283333333333333v3.283333333333333h3.283333333333333z m0 5v-3.2833333333333314h-3.283333333333333v3.2833333333333314h3.283333333333333z m1.7200000000000006-3.280000000000001v3.2833333333333314h3.283333333333333v-3.2833333333333314h-3.283333333333333z m0-5v3.2833333333333314h3.283333333333333v-3.283333333333333h-3.283333333333333z m5.000000000000002 5v3.2833333333333314h3.2833333333333314v-3.2833333333333314h-3.2833333333333314z m0-5v3.2833333333333314h3.2833333333333314v-3.283333333333333h-3.2833333333333314z m15-5q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.3049999999999997v16.716666666666665q0 1.3300000000000018-0.9766666666666666 2.3066666666666684t-2.306666666666665 0.9750000000000014h-26.713333333333342q-1.330000000000001 0-2.3066666666666675-0.9766666666666666t-0.9766666666666666-2.306666666666665v-16.71333333333334q0-1.3299999999999983 0.9766666666666666-2.306666666666665t2.3050000000000006-0.9766666666666666h26.71666666666667z' })
                )
            );
        }
    }]);

    return MdKeyboard;
}(React.Component);

exports.default = MdKeyboard;
module.exports = exports['default'];