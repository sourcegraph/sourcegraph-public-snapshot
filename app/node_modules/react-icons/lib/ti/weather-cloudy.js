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

var TiWeatherCloudy = function (_React$Component) {
    _inherits(TiWeatherCloudy, _React$Component);

    function TiWeatherCloudy() {
        _classCallCheck(this, TiWeatherCloudy);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiWeatherCloudy).apply(this, arguments));
    }

    _createClass(TiWeatherCloudy, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.333333333333336 31.666666666666668h-18.333333333333336c-3.6766666666666667 0-6.666666666666667-2.9899999999999984-6.666666666666667-6.666666666666668 0-3.1000000000000014 2.128333333333333-5.716666666666669 4.999999999999999-6.456666666666667v-0.2099999999999973c0-5.516666666666666 4.483333333333334-10 10-10 4.311666666666667 0 8.040000000000003 2.7300000000000004 9.416666666666668 6.691666666666666 4.903333333333336-0.408333333333335 8.916666666666671 3.5216666666666647 8.916666666666671 8.308333333333334 0 4.594999999999999-3.7383333333333297 8.333333333333336-8.333333333333336 8.333333333333336z m-18.491666666666667-10.010000000000002c-1.6800000000000015 0.010000000000001563-3.1750000000000016 1.5050000000000026-3.1750000000000016 3.3433333333333337s1.495 3.333333333333332 3.333333333333333 3.333333333333332h18.333333333333336c2.7566666666666677 0 5-2.2433333333333323 5-5s-2.2433333333333323-5-5-5c-0.40500000000000114 0-0.8333333333333321 0.06666666666666643-1.3500000000000014 0.216666666666665l-1.7916666666666679 0.5116666666666667-0.3099999999999987-1.8383333333333347c-0.543333333333333-3.219999999999997-3.2983333333333356-5.556666666666661-6.548333333333332-5.556666666666661-3.676666666666666 0-6.666666666666668 2.99-6.666666666666668 6.666666666666668 0 0.45333333333333314 0.04499999999999993 0.908333333333335 0.13666666666666671 1.3500000000000014l0.4066666666666663 2-2.366666666666667-0.026666666666667282z' })
                )
            );
        }
    }]);

    return TiWeatherCloudy;
}(React.Component);

exports.default = TiWeatherCloudy;
module.exports = exports['default'];