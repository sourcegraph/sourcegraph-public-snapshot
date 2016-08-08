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

var GoScreenNormal = function (_React$Component) {
    _inherits(GoScreenNormal, _React$Component);

    function GoScreenNormal() {
        _classCallCheck(this, GoScreenNormal);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoScreenNormal).apply(this, arguments));
    }

    _createClass(GoScreenNormal, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm7.4975000000000005 7.4975000000000005h-4.9975000000000005v2.5h7.4975000000000005v-7.4975000000000005h-2.5v4.9975000000000005z m-4.9975000000000005 25.004999999999995h4.9975000000000005v4.997500000000002h2.5v-7.497500000000002h-7.4975000000000005v2.5z m30.002499999999998-25.004999999999995v-4.997500000000002h-2.5v7.4975000000000005h7.497500000000002v-2.5h-4.997500000000002z m-2.5 30.002499999999998h2.5v-4.997500000000002h4.997500000000002v-2.5h-7.497500000000002v7.497500000000002z m-20-10h19.997500000000002v-15h-19.997500000000002v15z m4.997500000000002-10h10v5h-10v-5z' })
                )
            );
        }
    }]);

    return GoScreenNormal;
}(React.Component);

exports.default = GoScreenNormal;
module.exports = exports['default'];