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

var FaMapO = function (_React$Component) {
    _inherits(FaMapO, _React$Component);

    function FaMapO() {
        _classCallCheck(this, FaMapO);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMapO).apply(this, arguments));
    }

    _createClass(FaMapO, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm39.45375 2.7150000000000003q0.5462500000000006 0.39124999999999943 0.5462500000000006 1.0349999999999997v27.5q0 0.39124999999999943-0.21249999999999858 0.7025000000000006t-0.5675000000000026 0.45000000000000284l-12.5 5q-0.46875 0.21249999999999858-0.9375 0l-12.032499999999999-4.8075000000000045-12.02875 4.805000000000007q-0.1962499999999996 0.09999999999999432-0.4712499999999995 0.09999999999999432-0.37 0-0.7000000000000001-0.21625000000000227-0.5499999999999999-0.39124999999999943-0.5499999999999999-1.0337499999999977v-27.5q0-0.39250000000000007 0.21625-0.7037499999999994t0.5662499999999999-0.4500000000000002l12.5-5q0.46875-0.2124999999999999 0.9375 0l12.03 4.8075 12.033749999999998-4.80625q0.625-0.2537500000000006 1.1724999999999994 0.11749999999999972z m-25.07875 2.6374999999999997v24.80375l11.25 4.4925v-24.80625z m-11.875 4.235v24.805l10.625-4.237500000000001v-24.805z m35 20.82v-24.8025l-10.625 4.237500000000001v24.807499999999997z' })
                )
            );
        }
    }]);

    return FaMapO;
}(React.Component);

exports.default = FaMapO;
module.exports = exports['default'];