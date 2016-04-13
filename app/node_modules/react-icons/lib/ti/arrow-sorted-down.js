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

var TiArrowSortedDown = function (_React$Component) {
    _inherits(TiArrowSortedDown, _React$Component);

    function TiArrowSortedDown() {
        _classCallCheck(this, TiArrowSortedDown);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowSortedDown).apply(this, arguments));
    }

    _createClass(TiArrowSortedDown, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm9.7 16.2l10.3 10.5 10.3-10.5c0.3999999999999986-0.40000000000000036 0.5-0.9000000000000004 0.5-1.1999999999999993s-0.10000000000000142-0.8000000000000007-0.5-1.1999999999999993c-0.3000000000000007-0.3000000000000007-0.6000000000000014-0.5-1.1000000000000014-0.5h-18.4c-0.5 0-0.8000000000000007 0.1999999999999993-1.0999999999999996 0.5-0.40000000000000036 0.40000000000000036-0.5 0.6999999999999993-0.5 1.1999999999999993s0.09999999999999964 0.8000000000000007 0.5 1.1999999999999993z' })
                )
            );
        }
    }]);

    return TiArrowSortedDown;
}(React.Component);

exports.default = TiArrowSortedDown;
module.exports = exports['default'];