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

var TiLocationArrow = function (_React$Component) {
    _inherits(TiLocationArrow, _React$Component);

    function TiLocationArrow() {
        _classCallCheck(this, TiLocationArrow);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiLocationArrow).apply(this, arguments));
    }

    _createClass(TiLocationArrow, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.28 31.83666666666667c0.5833333333333321 1.7500000000000036 1.6833333333333336 1.8099999999999987 2.4633333333333347 0.14333333333333442l8.849999999999998-18.958333333333336c0.7766666666666673-1.67 0.054999999999999716-2.3900000000000006-1.6133333333333333-1.6116666666666664l-18.96 8.846666666666666c-1.6666666666666652 0.7783333333333324-1.6049999999999986 1.8833333333333329 0.14166666666666927 2.4666666666666686l6.838333333333331 2.2766666666666637 2.280000000000001 6.836666666666666z' })
                )
            );
        }
    }]);

    return TiLocationArrow;
}(React.Component);

exports.default = TiLocationArrow;
module.exports = exports['default'];