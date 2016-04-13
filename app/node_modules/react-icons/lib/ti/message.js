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

var TiMessage = function (_React$Component) {
    _inherits(TiMessage, _React$Component);

    function TiMessage() {
        _classCallCheck(this, TiMessage);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMessage).apply(this, arguments));
    }

    _createClass(TiMessage, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 11.666666666666668c0.9033333333333324 0 1.6666666666666679 0.7633333333333336 1.6666666666666679 1.666666666666666v11.666666666666666c0 0.9033333333333324-0.7633333333333319 1.6666666666666679-1.6666666666666679 1.6666666666666679h-14.716666666666667l-0.2833333333333332 0.283333333333335v-0.283333333333335h-5c-0.9033333333333342 0-1.666666666666666-0.7633333333333319-1.666666666666666-1.6666666666666679v-11.666666666666666c0-0.9033333333333342 0.7633333333333336-1.666666666666666 1.666666666666666-1.666666666666666h20z m0-3.333333333333334h-20c-2.75 0-5 2.25-5 5v11.666666666666666c0 2.75 2.25 5 5 5h1.666666666666666v5l4.999999999999998-5h13.333333333333336c2.75 0 5-2.25 5-5v-11.666666666666666c0-2.75-2.25-5-5-5z' })
                )
            );
        }
    }]);

    return TiMessage;
}(React.Component);

exports.default = TiMessage;
module.exports = exports['default'];