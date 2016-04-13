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

var GoHorizontalRule = function (_React$Component) {
    _inherits(GoHorizontalRule, _React$Component);

    function GoHorizontalRule() {
        _classCallCheck(this, GoHorizontalRule);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoHorizontalRule).apply(this, arguments));
    }

    _createClass(GoHorizontalRule, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm9.9975 17.5h5v5h2.5000000000000018v-14.9975h-2.5v7.4975000000000005h-5v-7.4975000000000005h-2.4975000000000023v14.9975h2.4975000000000005v-5z m22.497500000000002 5v-5h-2.4974999999999987v5h2.4974999999999987z m0-7.5v-4.997499999999999h-2.4974999999999987v4.997499999999999h2.4974999999999987z m-7.497500000000002 0v-4.997499999999999h5v-2.5h-7.5v14.997499999999999h2.5v-5h5v-2.5h-5z m-17.497500000000002 17.5h24.994999999999997v-5h-24.994999999999997v5z' })
                )
            );
        }
    }]);

    return GoHorizontalRule;
}(React.Component);

exports.default = GoHorizontalRule;
module.exports = exports['default'];