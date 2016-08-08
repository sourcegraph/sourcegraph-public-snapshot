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

var TiVendorApple = function (_React$Component) {
    _inherits(TiVendorApple, _React$Component);

    function TiVendorApple() {
        _classCallCheck(this, TiVendorApple);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiVendorApple).apply(this, arguments));
    }

    _createClass(TiVendorApple, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm19.8 11s-0.10000000000000142-2.6999999999999993 1.5-5l4.699999999999999-2.7s0.1999999999999993 2.7-1.5 5.000000000000001l-4.699999999999999 2.6999999999999993z m9 9.3c0-2.5 1.3999999999999986-4.6 3.400000000000002-6l-1.5-1.5c-0.8999999999999986-0.5-1.8999999999999986-1.0999999999999996-4-1.0999999999999996-2.3999999999999986 0-4 1.5-6.199999999999999 1.5s-3.6999999999999993-1.4000000000000004-5.199999999999999-1.5c-1.0999999999999996 0-2.3000000000000007 0-3.5 0.5-0.8000000000000007 0.3000000000000007-2 1.0999999999999996-2.5999999999999996 2-1 1-2 3.0999999999999996-2.2 5.1-0.20000000000000018 2-0.20000000000000018 3.5 0.2999999999999998 5.400000000000002 0.40000000000000036 1.5 1 3 1.6999999999999993 4.300000000000001 0.5 1 1 2 1.6999999999999993 2.9999999999999964 0.5 0.7000000000000028 1.0999999999999996 1.2999999999999972 1.8000000000000007 1.7999999999999972 0.5 0.3999999999999986 1 0.7000000000000028 1.6999999999999993 1 0.3000000000000007 0 0.8000000000000007 0.20000000000000284 1.3000000000000007 0.20000000000000284 1-0.20000000000000284 2.6999999999999993-1.5 4-1.7999999999999972 0.6999999999999993-0.20000000000000284 1.3000000000000007-0.20000000000000284 2.1999999999999993 0 1.1000000000000014 0.10000000000000142 2.3000000000000007 1.5 3.6000000000000014 1.6000000000000014 1 0.20000000000000284 2 0 2.8999999999999986-0.5 0.6000000000000014-0.29999999999999716 1.1000000000000014-0.7999999999999972 1.6000000000000014-1.5 0.6999999999999993-0.6000000000000014 1.1999999999999993-1.5 1.6999999999999993-2.1000000000000014 0.6999999999999993-1.1999999999999993 1.4999999999999964-2.5 1.8000000000000007-3.8999999999999986-2.6000000000000014-1-4.5-3.5-4.5-6.5z' })
                )
            );
        }
    }]);

    return TiVendorApple;
}(React.Component);

exports.default = TiVendorApple;
module.exports = exports['default'];