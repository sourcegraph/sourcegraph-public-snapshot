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

var GoLink = function (_React$Component) {
    _inherits(GoLink, _React$Component);

    function GoLink() {
        _classCallCheck(this, GoLink);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoLink).apply(this, arguments));
    }

    _createClass(GoLink, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 10h-5.391249999999999c1.875 1.25 3.6325000000000003 3.4749999999999996 4.18 5h1.1750000000000007c2.537499999999998 0 5.0000000000000036 2.5 5.0000000000000036 5s-2.539999999999999 5-5 5h-7.5c-2.4624999999999986 0-5-2.5-5-5 0-0.8999999999999986 0.27250000000000085-1.7575000000000003 0.7025000000000006-2.5h-5.350000000000001c-0.19624999999999915 0.8200000000000003-0.3125 1.6412499999999994-0.3125 2.5 0 5 4.960000000000001 10 9.96 10h7.5362499999999955s10-5 10-10-5-10-10-10z m-18.7875 15h-1.1749999999999972c-2.5374999999999996 0-5-2.5-5-5s2.54-5 5-5h7.5c2.4624999999999986 0 5 2.5 5 5 0 0.8999999999999986-0.27250000000000085 1.7575000000000003-0.7025000000000006 2.5h5.350000000000001c0.19624999999999915-0.8200000000000003 0.3125-1.6412499999999994 0.3125-2.5 0-5-4.960000000000001-10-9.96-10h-7.537500000000001s-10 5-10 10 5 10 10 10h5.390000000000001c-1.875-1.25-3.6325000000000003-3.4750000000000014-4.18-5z' })
                )
            );
        }
    }]);

    return GoLink;
}(React.Component);

exports.default = GoLink;
module.exports = exports['default'];