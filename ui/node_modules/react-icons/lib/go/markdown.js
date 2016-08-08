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

var GoMarkdown = function (_React$Component) {
    _inherits(GoMarkdown, _React$Component);

    function GoMarkdown() {
        _classCallCheck(this, GoMarkdown);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoMarkdown).apply(this, arguments));
    }

    _createClass(GoMarkdown, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.115 7.5h-34.23125c-1.5899999999999992 0-2.883749999999999 1.2937499999999993-2.883749999999999 2.885v19.231250000000003c0 1.5912500000000023 1.2912499999999998 2.883749999999999 2.8825000000000003 2.883749999999999h34.23125c1.592499999999994 0 2.886249999999997-1.2912500000000016 2.886249999999997-2.8825000000000003v-19.2325c0-1.589999999999998-1.2937500000000028-2.884999999999998-2.884999999999998-2.884999999999998z m-14.615000000000002 19.994999999999997l-5 0.005000000000002558v-7.5l-3.75 4.807500000000001-3.75-4.807500000000001v7.5h-5v-15h5l3.75 5 3.75-5 5-0.005000000000000782v14.999999999999998z m7.465 1.25l-6.215-8.744999999999997h3.75v-7.5h5v7.5h3.75l-6.285 8.745000000000001z' })
                )
            );
        }
    }]);

    return GoMarkdown;
}(React.Component);

exports.default = GoMarkdown;
module.exports = exports['default'];