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

var TiFlag = function (_React$Component) {
    _inherits(TiFlag, _React$Component);

    function TiFlag() {
        _classCallCheck(this, TiFlag);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiFlag).apply(this, arguments));
    }

    _createClass(TiFlag, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30.638333333333332 7.196666666666666c-0.6233333333333348-0.25833333333333375-1.3399999999999999-0.11666666666666625-1.8166666666666664 0.36166666666666636-2.1066666666666656 2.1049999999999995-5.533333333333331 2.1066666666666665-7.643333333333334 0-3.408333333333335-3.4050000000000002-8.950000000000001-3.4050000000000002-12.356666666666667 0-0.3133333333333326 0.31166666666666654-0.4883333333333333 0.7366666666666672-0.4883333333333333 1.1783333333333337v21.66666666666667c0 0.9200000000000017 0.7449999999999992 1.6666666666666643 1.666666666666666 1.6666666666666643s1.666666666666666-0.7466666666666661 1.666666666666666-1.6666666666666679v-7.586666666666666c2.116666666666667-1.663333333333334 5.200000000000001-1.5199999999999996 7.154999999999999 0.43333333333333357 3.408333333333335 3.4033333333333324 8.95 3.4033333333333324 12.356666666666666 0 0.3133333333333326-0.31666666666666643 0.4883333333333333-0.7383333333333333 0.4883333333333333-1.1799999999999997v-13.333333333333334c0-0.6750000000000007-0.4066666666666663-1.2833333333333332-1.0283333333333324-1.54z' })
                )
            );
        }
    }]);

    return TiFlag;
}(React.Component);

exports.default = TiFlag;
module.exports = exports['default'];