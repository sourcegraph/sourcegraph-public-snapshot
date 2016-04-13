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

var TiZoom = function (_React$Component) {
    _inherits(TiZoom, _React$Component);

    function TiZoom() {
        _classCallCheck(this, TiZoom);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiZoom).apply(this, arguments));
    }

    _createClass(TiZoom, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.666666666666668 6.666666666666667c-6.433333333333334 0-11.666666666666668 5.233333333333333-11.666666666666668 11.666666666666668 0 1.2716666666666683 0.211666666666666 2.491666666666667 0.5899999999999999 3.638333333333332l-2.0999999999999996 2.1033333333333353-1.6799999999999997 1.7416666666666671c-0.9366666666666665 0.9283333333333346-1.4833333333333334 2.241666666666667-1.4833333333333334 3.6416666666666657 0 2.8783333333333303 2.338333333333334 5.216666666666669 5.216666666666666 5.216666666666669 1.2599999999999998 0 2.504999999999999-0.461666666666666 3.505000000000001-1.3049999999999997l0.10500000000000043-0.09166666666666856 0.09999999999999964-0.10000000000000142 3.7716666666666647-3.7666666666666657c1.1466666666666683 0.37666666666666515 2.366666666666667 0.5883333333333347 3.6400000000000006 0.5883333333333347 6.433333333333334 0 11.666666666666664-5.233333333333334 11.666666666666664-11.666666666666668s-5.233333333333334-11.666666666666668-11.666666666666668-11.666666666666668z m0 20c-4.595000000000002 0-8.333333333333334-3.7383333333333333-8.333333333333334-8.333333333333332s3.7383333333333315-8.333333333333334 8.333333333333334-8.333333333333334 8.333333333333332 3.7383333333333333 8.333333333333332 8.333333333333334-3.7383333333333333 8.333333333333336-8.333333333333336 8.333333333333336z m0-15c-3.6750000000000007 0-6.666666666666668 2.99-6.666666666666668 6.666666666666668s2.991666666666667 6.666666666666668 6.666666666666668 6.666666666666668 6.666666666666668-2.9899999999999984 6.666666666666668-6.666666666666668-2.991666666666667-6.666666666666668-6.666666666666668-6.666666666666668z m0 11.666666666666668c-2.7600000000000016 0-5-2.240000000000002-5-5s2.2399999999999984-5 5-5 5 2.24 5 5-2.2399999999999984 5-5 5z' })
                )
            );
        }
    }]);

    return TiZoom;
}(React.Component);

exports.default = TiZoom;
module.exports = exports['default'];