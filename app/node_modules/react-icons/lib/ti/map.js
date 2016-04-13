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

var TiMap = function (_React$Component) {
    _inherits(TiMap, _React$Component);

    function TiMap() {
        _classCallCheck(this, TiMap);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMap).apply(this, arguments));
    }

    _createClass(TiMap, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.971666666666664 5.126666666666667c-0.6216666666666697-0.25833333333333375-1.3400000000000034-0.11666666666666625-1.81666666666667 0.36166666666666636l-6.444999999999993 6.445-7.166666666666668-5.733333333333333c-0.6616666666666653-0.5283333333333333-1.6233333333333348-0.47999999999999954-2.2216666666666676 0.12166666666666703l-7.5 7.5c-0.3133333333333308 0.3116666666666674-0.4883333333333315 0.7333333333333343-0.4883333333333315 1.1783333333333328v16.666666666666668c0 0.6733333333333356 0.40499999999999936 1.283333333333335 1.0283333333333342 1.5400000000000027 0.206666666666667 0.08833333333333115 0.4233333333333338 0.12666666666666515 0.6383333333333336 0.12666666666666515 0.43333333333333357 0 0.8599999999999994-0.1700000000000017 1.1783333333333328-0.48833333333333684l6.445-6.445 7.166666666666668 5.733333333333334c0.6616666666666653 0.528333333333336 1.620000000000001 0.47666666666666657 2.2216666666666676-0.1216666666666697l7.5-7.5c0.3133333333333326-0.3116666666666674 0.48833333333333684-0.7333333333333343 0.48833333333333684-1.1783333333333346v-16.66666666666666c0-0.6733333333333347-0.40500000000000114-1.283333333333334-1.028333333333336-1.540000000000001z m-22.304999999999996 22.516666666666666v-11.953333333333331l5-5v12.066666666666668c-0.11666666666666714 0.07166666666666544-5 4.886666666666667-5 4.886666666666667z m6.875-4.778333333333332c-0.06333333333333258-0.05000000000000071-0.14000000000000057-0.07666666666666799-0.20833333333333215-0.11666666666666714v-12.443333333333332s6.600000000000001 5.238333333333333 6.666666666666668 5.276666666666667v12.455000000000002l-6.458333333333332-5.166666666666668z m13.125-0.22333333333332916l-5 5v-12.058333333333335c0.11666666666666714-0.07166666666666721 5-4.893333333333333 5-4.893333333333333v11.953333333333333z' })
                )
            );
        }
    }]);

    return TiMap;
}(React.Component);

exports.default = TiMap;
module.exports = exports['default'];