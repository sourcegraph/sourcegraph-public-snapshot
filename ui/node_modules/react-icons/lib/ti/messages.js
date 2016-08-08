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

var TiMessages = function (_React$Component) {
    _inherits(TiMessages, _React$Component);

    function TiMessages() {
        _classCallCheck(this, TiMessages);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMessages).apply(this, arguments));
    }

    _createClass(TiMessages, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 11.666666666666668h-5c0-2.75-2.25-5-5-5h-20c-2.7499999999999996-8.881784197001252e-16-5 2.2499999999999982-5 5v11.666666666666668c0 2.75 2.2500000000000004 5 5 5v5l5-5c0 2.75 2.25 5 5 5h13.333333333333336l5 5v-5h1.6666666666666643c2.75 0 5-2.25 5-5v-11.666666666666668c0-2.75-2.25-5-5-5z m-30 13.333333333333332c-0.9033333333333333 0-1.666666666666667-0.7633333333333319-1.666666666666667-1.6666666666666679v-11.666666666666664c0-0.9033333333333342 0.7633333333333336-1.666666666666666 1.666666666666667-1.666666666666666h20c0.9033333333333324 0 1.6666666666666679 0.7633333333333336 1.6666666666666679 1.666666666666666v1.666666666666666h-10.833333333333334c-2.3000000000000007 0-4.166666666666666 1.8666666666666654-4.166666666666666 4.166666666666666v7.5h-6.666666666666668z m31.66666666666667 3.333333333333332c0 0.9033333333333324-0.7633333333333354 1.6666666666666679-1.6666666666666643 1.6666666666666679h-20.000000000000007c-0.9033333333333342 0-1.666666666666666-0.7633333333333319-1.666666666666666-1.6666666666666679v-10.833333333333332c0-1.3783333333333339 1.1216666666666661-2.5 2.5-2.5h19.166666666666664c0.903333333333336 0 1.6666666666666643 0.7633333333333336 1.6666666666666643 1.6666666666666679v11.666666666666668z' })
                )
            );
        }
    }]);

    return TiMessages;
}(React.Component);

exports.default = TiMessages;
module.exports = exports['default'];