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

var TiArrowUnsorted = function (_React$Component) {
    _inherits(TiArrowUnsorted, _React$Component);

    function TiArrowUnsorted() {
        _classCallCheck(this, TiArrowUnsorted);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowUnsorted).apply(this, arguments));
    }

    _createClass(TiArrowUnsorted, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30.3 15.5l-10.3-10.5-10.3 10.5c-0.40000000000000036 0.3000000000000007-0.5 0.6999999999999993-0.5 1.1999999999999993s0.09999999999999964 0.8000000000000007 0.5 1.1000000000000014c0.3000000000000007 0.3999999999999986 0.5999999999999996 0.5 1.0999999999999996 0.5h18.4c0.5 0 0.8000000000000007-0.10000000000000142 1.1000000000000014-0.5 0.3999999999999986-0.3000000000000007 0.5-0.8000000000000007 0.5-1.1000000000000014s-0.10000000000000142-0.9000000000000004-0.5-1.1999999999999993z m-20.6 9l10.3 10.5 10.3-10.5c0.3999999999999986-0.3000000000000007 0.5-0.8000000000000007 0.5-1.1999999999999993s-0.10000000000000142-0.8000000000000007-0.5-1.1000000000000014c-0.3000000000000007-0.3999999999999986-0.6000000000000014-0.5-1.1000000000000014-0.5h-18.4c-0.5 0-0.8000000000000007 0.10000000000000142-1.0999999999999996 0.5-0.40000000000000036 0.3000000000000007-0.5 0.8000000000000007-0.5 1.1000000000000014s0.09999999999999964 0.8999999999999986 0.5 1.1999999999999993z' })
                )
            );
        }
    }]);

    return TiArrowUnsorted;
}(React.Component);

exports.default = TiArrowUnsorted;
module.exports = exports['default'];