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

var TiWeatherNight = function (_React$Component) {
    _inherits(TiWeatherNight, _React$Component);

    function TiWeatherNight() {
        _classCallCheck(this, TiWeatherNight);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiWeatherNight).apply(this, arguments));
    }

    _createClass(TiWeatherNight, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.5 33.333333333333336c-1.4333333333333336 0-2.8499999999999996-0.25-4.205-0.75l-2.1950000000000003-0.8099999999999987 1.4833333333333325-1.8116666666666674c2.125-2.6000000000000014 3.25-5.756666666666668 3.25-9.128333333333334s-1.125-6.528333333333332-3.25-9.128333333333334l-1.4833333333333343-1.8116666666666674 2.1933333333333334-0.8083333333333336c1.3566666666666674-0.5016666666666669 2.7733333333333388-0.7516666666666669 4.206666666666669-0.7516666666666669 6.893333333333334 0 12.5 5.606666666666667 12.5 12.500000000000002s-5.606666666666666 12.5-12.5 12.5z m-0.8500000000000014-3.373333333333335c0.283333333333335 0.026666666666667282 0.5666666666666664 0.03999999999999915 0.8500000000000014 0.03999999999999915 5.053333333333335 0 9.166666666666668-4.113333333333333 9.166666666666668-9.166666666666668s-4.113333333333333-9.166666666666664-9.166666666666668-9.166666666666664c-0.283333333333335 0-0.5666666666666664 0.013333333333333641-0.8500000000000014 0.03999999999999915 1.6533333333333324 2.741666666666667 2.5166666666666657 5.849999999999998 2.5166666666666657 9.126666666666669s-0.8633333333333333 6.383333333333333-2.5166666666666657 9.126666666666665z' })
                )
            );
        }
    }]);

    return TiWeatherNight;
}(React.Component);

exports.default = TiWeatherNight;
module.exports = exports['default'];