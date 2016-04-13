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

var GoMail = function (_React$Component) {
    _inherits(GoMail, _React$Component);

    function GoMail() {
        _classCallCheck(this, GoMail);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoMail).apply(this, arguments));
    }

    _createClass(GoMail, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm2.5 7.5v25h35v-25h-35z m30 2.5l-12.5 10.3125-12.5-10.3125h25z m-27.5 2.5l9.845 7.484999999999999-9.845 7.515000000000001v-15z m2.5 17.5l9.9225-8.056249999999999 2.5775000000000006 1.9624999999999986 2.5749999999999993-1.9574999999999996 9.925 8.05125h-25z m27.5-2.5l-9.85875-7.502499999999998 9.85875-7.497500000000002v15z' })
                )
            );
        }
    }]);

    return GoMail;
}(React.Component);

exports.default = GoMail;
module.exports = exports['default'];