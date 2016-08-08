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

var GoCircleSlash = function (_React$Component) {
    _inherits(GoCircleSlash, _React$Component);

    function GoCircleSlash() {
        _classCallCheck(this, GoCircleSlash);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoCircleSlash).apply(this, arguments));
    }

    _createClass(GoCircleSlash, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 7.5c-6.905000000000001 0-12.5 5.595000000000001-12.5 12.5s5.594999999999999 12.5 12.5 12.5c6.903749999999999 0 12.5-5.596250000000001 12.5-12.5s-5.596250000000001-12.5-12.5-12.5z m0 5c1.0799999999999983 0 2.0975 0.2375000000000007 3.0249999999999986 0.6475000000000009l-9.877499999999998 9.877500000000001c-0.41000000000000014-0.927500000000002-0.6475000000000009-1.9450000000000003-0.6475000000000009-3.025000000000002 0-4.141249999999999 3.3575-7.5 7.5-7.5z m0 15c-1.0949999999999989 0-2.129999999999999-0.25-3.0700000000000003-0.6687499999999993l9.91875-9.8625c0.4125000000000014 0.9287500000000009 0.6499999999999986 1.9499999999999993 0.6499999999999986 3.03125 0 4.142499999999998-3.357499999999998 7.5-7.5 7.5z' })
                )
            );
        }
    }]);

    return GoCircleSlash;
}(React.Component);

exports.default = GoCircleSlash;
module.exports = exports['default'];