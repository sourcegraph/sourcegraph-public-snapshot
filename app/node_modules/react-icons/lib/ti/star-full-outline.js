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

var TiStarFullOutline = function (_React$Component) {
    _inherits(TiStarFullOutline, _React$Component);

    function TiStarFullOutline() {
        _classCallCheck(this, TiStarFullOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiStarFullOutline).apply(this, arguments));
    }

    _createClass(TiStarFullOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm5.2 18.8l5.999999999999999 5.5-1.6999999999999993 7.699999999999999c-0.1999999999999993 1 0.1999999999999993 2 1 2.5 0.3000000000000007 0.29999999999999716 0.8000000000000007 0.5 1.3000000000000007 0.5 0.40000000000000036 0 0.6999999999999993 0 1-0.20000000000000284 0 0 0.1999999999999993 0 0.1999999999999993-0.10000000000000142l6.800000000000001-3.8999999999999986 6.900000000000002 3.8999999999999986s0.10000000000000142 0 0.10000000000000142 0.10000000000000142c0.8999999999999986 0.3999999999999986 1.8999999999999986 0.3999999999999986 2.5-0.10000000000000142 0.8999999999999986-0.5 1.1999999999999993-1.5 1-2.5l-1.6000000000000014-7.699999999999999c0.6000000000000014-0.5 1.6000000000000014-1.5 2.6000000000000014-2.5l3.200000000000003-2.8000000000000007 0.20000000000000284-0.1999999999999993c0.6000000000000014-0.6999999999999993 0.7999999999999972-1.6999999999999993 0.5-2.5s-1-1.5-2-1.6999999999999993h-0.20000000000000995l-7.800000000000001-0.8000000000000007-3.1999999999999993-7.199999999999997s0-0.09999999999999964-0.1999999999999993-0.09999999999999964c-0.10000000000000142-1.2000000000000002-1-1.7000000000000002-1.8000000000000007-1.7000000000000002s-1.6999999999999993 0.5-2.1999999999999993 1.2999999999999998c0 0 0 0.20000000000000018-0.10000000000000142 0.20000000000000018l-3.1999999999999993 7.199999999999999-7.8 0.8000000000000007h-0.20000000000000018c-0.7999999999999998 0.1999999999999993-1.7000000000000002 0.8000000000000007-2 1.6999999999999993-0.20000000000000018 1 0 2 0.7000000000000002 2.6000000000000014z' })
                )
            );
        }
    }]);

    return TiStarFullOutline;
}(React.Component);

exports.default = TiStarFullOutline;
module.exports = exports['default'];