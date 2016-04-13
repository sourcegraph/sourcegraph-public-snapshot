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

var TiTick = function (_React$Component) {
    _inherits(TiTick, _React$Component);

    function TiTick() {
        _classCallCheck(this, TiTick);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiTick).apply(this, arguments));
    }

    _createClass(TiTick, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.28666666666667 10.416666666666668c-1.6116666666666681-0.8949999999999996-3.6416666666666657-0.3116666666666674-4.533333333333335 1.2966666666666669l-6.186666666666667 11.136666666666667-3.543333333333333-3.541666666666668c-1.3000000000000007-1.3000000000000007-3.411666666666667-1.3000000000000007-4.713333333333333 0s-1.3000000000000007 3.4116666666666653 0 4.713333333333335l6.666666666666666 6.666666666666668c0.629999999999999 0.6333333333333329 1.4800000000000004 0.9783333333333317 2.3566666666666656 0.9783333333333317 0.15333333333333243 0 0.30833333333333357-0.010000000000001563 0.461666666666666-0.03333333333333499 1.033333333333335-0.14499999999999957 1.9433333333333316-0.7666666666666657 2.4499999999999993-1.6833333333333336l8.333333333333336-15c0.8966666666666683-1.6083333333333343 0.31666666666666643-3.6366666666666667-1.2916666666666679-4.533333333333335z' })
                )
            );
        }
    }]);

    return TiTick;
}(React.Component);

exports.default = TiTick;
module.exports = exports['default'];