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

var TiMessageTyping = function (_React$Component) {
    _inherits(TiMessageTyping, _React$Component);

    function TiMessageTyping() {
        _classCallCheck(this, TiMessageTyping);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMessageTyping).apply(this, arguments));
    }

    _createClass(TiMessageTyping, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 10h-21.666666666666664c-2.7500000000000018 0-5.000000000000002 2.25-5.000000000000002 5v11.666666666666668c-4.440892098500626e-16 2.75 2.25 5 5 5h1.666666666666666v5.0000000000000036l5-5h15c2.75 0 5-2.25 5-5v-11.666666666666671c0-2.75-2.25-5-5-5z m1.6666666666666679 16.666666666666668c0 0.9033333333333324-0.7633333333333319 1.6666666666666679-1.6666666666666679 1.6666666666666679h-21.666666666666664c-0.9033333333333351 0-1.6666666666666687-0.7633333333333319-1.6666666666666687-1.6666666666666679v-11.666666666666668c0-0.9033333333333342 0.7633333333333336-1.666666666666666 1.666666666666667-1.666666666666666h21.666666666666664c0.9033333333333324 0 1.6666666666666679 0.7633333333333336 1.6666666666666679 1.666666666666666v11.666666666666668z m-20-2.5c-1.8399999999999999 0-3.333333333333334-1.4933333333333323-3.333333333333334-3.333333333333332s1.493333333333334-3.333333333333332 3.333333333333334-3.333333333333332 3.333333333333334 1.4933333333333323 3.333333333333334 3.333333333333332-1.493333333333334 3.333333333333332-3.333333333333334 3.333333333333332z m0-5c-0.9199999999999999 0-1.666666666666666 0.7466666666666661-1.666666666666666 1.6666666666666679s0.7466666666666661 1.6666666666666679 1.666666666666666 1.6666666666666679 1.666666666666666-0.7466666666666661 1.666666666666666-1.6666666666666679-0.7466666666666661-1.6666666666666679-1.666666666666666-1.6666666666666679z m7.5 5c-1.8399999999999999 0-3.333333333333334-1.4933333333333323-3.333333333333334-3.333333333333332s1.493333333333334-3.333333333333332 3.333333333333334-3.333333333333332 3.333333333333332 1.4933333333333323 3.333333333333332 3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332 3.333333333333332z m0-5c-0.9200000000000017 0-1.6666666666666679 0.7466666666666661-1.6666666666666679 1.6666666666666679s0.7466666666666661 1.6666666666666679 1.6666666666666679 1.6666666666666679 1.6666666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679-0.7466666666666661-1.6666666666666679-1.6666666666666679-1.6666666666666679z m7.5 5c-1.8399999999999999 0-3.333333333333332-1.4933333333333323-3.333333333333332-3.333333333333332s1.4933333333333323-3.333333333333332 3.333333333333332-3.333333333333332 3.333333333333332 1.4933333333333323 3.333333333333332 3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332 3.333333333333332z m0-5c-0.9200000000000017 0-1.6666666666666679 0.7466666666666661-1.6666666666666679 1.6666666666666679s0.7466666666666661 1.6666666666666679 1.6666666666666679 1.6666666666666679 1.6666666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679-0.7466666666666661-1.6666666666666679-1.6666666666666679-1.6666666666666679z' })
                )
            );
        }
    }]);

    return TiMessageTyping;
}(React.Component);

exports.default = TiMessageTyping;
module.exports = exports['default'];