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

var GoCalendar = function (_React$Component) {
    _inherits(GoCalendar, _React$Component);

    function GoCalendar() {
        _classCallCheck(this, GoCalendar);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoCalendar).apply(this, arguments));
    }

    _createClass(GoCalendar, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.5 20h-2.5v5h2.5v-5z m-5 0h-2.5v5h2.5v-5z m5-7.5h-2.5v5h2.5v-5z m5 7.5h-2.5v5h2.5v-5z m-10 7.5h-2.5v5h2.5v-5z m7.5-27.5h-2.5v5h2.5v-5z m-20 0h-2.5v5h2.5v-5z m22.5 12.5h-2.5v5h2.5v-5z m-10 0h-2.5v5h2.5v-5z m-10 15h-2.5v5h2.5v-5z m-5-7.5h-2.5v5h2.5v-5z m5 0h-2.5v5h2.5v-5z m20-17.5v5h-7.5v-5h-12.5v5h-7.5v-5h-5v35h37.5v-35h-5z m2.5 32.5h-32.5v-25h32.5v25z m-27.5-7.5h-2.5v5h2.5v-5z m10-15h-2.5v5h2.5v-5z m0 15h-2.5v5h2.5v-5z m-5-15h-2.5v5h2.5v-5z m5 7.5h-2.5v5h2.5v-5z m10 7.5h-2.5v5h2.5v-5z' })
                )
            );
        }
    }]);

    return GoCalendar;
}(React.Component);

exports.default = GoCalendar;
module.exports = exports['default'];