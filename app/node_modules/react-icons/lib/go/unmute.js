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

var GoUnmute = function (_React$Component) {
    _inherits(GoUnmute, _React$Component);

    function GoUnmute() {
        _classCallCheck(this, GoUnmute);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoUnmute).apply(this, arguments));
    }

    _createClass(GoUnmute, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm7.5 15h-5v10h5l10 7.5h2.5v-25h-2.5l-10 7.5z m16.035 1.4649999999999999c-0.4875000000000007-0.4875000000000007-1.2800000000000011-0.4875000000000007-1.7674999999999983 0s-0.4875000000000007 1.2800000000000011 0 1.7674999999999983c0.9750000000000014 0.9750000000000014 0.9750000000000014 2.55875 0 3.535-0.4875000000000007 0.4875000000000007-0.4875000000000007 1.2800000000000011 0 1.7674999999999983s1.2800000000000011 0.4875000000000007 1.7674999999999983 0c1.9525000000000006-1.9525000000000006 1.9525000000000006-5.118749999999999 0-7.071249999999999z m3.5375000000000014-3.5374999999999996c-0.4875000000000007-0.4875000000000007-1.28125-0.4875000000000007-1.7687500000000007 0s-0.4875000000000007 1.28125 0 1.7687500000000007c2.928750000000001 2.928749999999999 2.928750000000001 7.6775 0 10.606250000000001-0.4875000000000007 0.4875000000000007-0.4875000000000007 1.2800000000000011 0 1.7674999999999983s1.2800000000000011 0.4875000000000007 1.7674999999999983 0c3.905000000000001-3.905000000000001 3.905000000000001-10.237499999999997 0-14.1425z m3.5337500000000013-3.5337499999999995c-0.4875000000000007-0.4875000000000007-1.2800000000000011-0.4875000000000007-1.7674999999999983 0s-0.4875000000000007 1.2799999999999994 0 1.7675c4.8825 4.880000000000001 4.8825 12.795000000000003 0 17.674999999999997-0.4875000000000007 0.48999999999999844-0.4875000000000007 1.28125 0 1.7687500000000007s1.28125 0.4875000000000007 1.7687500000000007 0c5.857499999999998-5.857499999999998 5.857499999999998-15.355 0-21.2125z' })
                )
            );
        }
    }]);

    return GoUnmute;
}(React.Component);

exports.default = GoUnmute;
module.exports = exports['default'];