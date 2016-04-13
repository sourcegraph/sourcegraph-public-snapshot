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

var MdDirectionsBus = function (_React$Component) {
    _inherits(MdDirectionsBus, _React$Component);

    function MdDirectionsBus() {
        _classCallCheck(this, MdDirectionsBus);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDirectionsBus).apply(this, arguments));
    }

    _createClass(MdDirectionsBus, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 18.36v-8.36h-20v8.36h20z m-2.5 10q1.0933333333333337 0 1.7966666666666669-0.7416666666666671t0.7033333333333331-1.7566666666666642-0.7033333333333331-1.7583333333333329-1.7966666666666669-0.7399999999999984-1.7966666666666669 0.7416666666666671-0.7033333333333331 1.7566666666666642 0.7033333333333331 1.7583333333333329 1.7966666666666669 0.7433333333333323z m-15 0q1.0933333333333337 0 1.7966666666666669-0.7416666666666671t0.7033333333333331-1.7566666666666642-0.7033333333333331-1.7583333333333329-1.7966666666666669-0.7399999999999984-1.7966666666666669 0.7416666666666671-0.7033333333333331 1.7566666666666642 0.7033333333333331 1.7583333333333329 1.7966666666666669 0.7433333333333323z m-5.86-1.7199999999999989v-16.64q0-3.9833333333333343 3.4383333333333335-5.3133333333333335t9.921666666666667-1.3283333333333327 9.921666666666667 1.3283333333333336 3.4383333333333326 5.313333333333333v16.64q0 2.1883333333333326-1.7166666666666686 3.75v2.9666666666666686q0 0.7049999999999983-0.46999999999999886 1.173333333333332t-1.1716666666666669 0.46666666666666856h-1.6383333333333319q-0.7033333333333331 0-1.211666666666666-0.46666666666666856t-0.5083333333333329-1.1716666666666669v-1.7166666666666686h-13.283333333333333v1.7166666666666686q0 0.7033333333333331-0.5066666666666659 1.1716666666666669t-1.2100000000000009 0.46666666666666856h-1.6433333333333344q-0.7033333333333331 0-1.1716666666666669-0.46666666666666856t-0.4666666666666668-1.1716666666666669v-2.9666666666666686q-1.7200000000000006-1.5666666666666664-1.7200000000000006-3.75z' })
                )
            );
        }
    }]);

    return MdDirectionsBus;
}(React.Component);

exports.default = MdDirectionsBus;
module.exports = exports['default'];