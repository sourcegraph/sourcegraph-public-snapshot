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

var FaPlay = function (_React$Component) {
    _inherits(FaPlay, _React$Component);

    function FaPlay() {
        _classCallCheck(this, FaPlay);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaPlay).apply(this, arguments));
    }

    _createClass(FaPlay, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.17857142857143 20.69142857142857l-29.642857142857146 16.47142857142857q-0.5142857142857133 0.29142857142856826-0.8814285714285708 0.06857142857143117t-0.36857142857142833-0.8028571428571425v-32.85714285714286q0-0.5800000000000001 0.36857142857142833-0.8028571428571429t0.8814285714285717 0.06714285714285717l29.642857142857146 16.471428571428575q0.5142857142857125 0.2914285714285718 0.5142857142857125 0.6928571428571431t-0.5142857142857125 0.6914285714285704z' })
                )
            );
        }
    }]);

    return FaPlay;
}(React.Component);

exports.default = FaPlay;
module.exports = exports['default'];