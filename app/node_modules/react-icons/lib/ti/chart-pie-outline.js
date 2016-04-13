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

var TiChartPieOutline = function (_React$Component) {
    _inherits(TiChartPieOutline, _React$Component);

    function TiChartPieOutline() {
        _classCallCheck(this, TiChartPieOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiChartPieOutline).apply(this, arguments));
    }

    _createClass(TiChartPieOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30.378333333333334 12.683333333333334l0.9283333333333346-0.9333333333333336c0.6600000000000001-0.6600000000000001 1.0116666666666667-1.5716666666666672 0.9700000000000024-2.5066666666666677s-0.47666666666666657-1.8066666666666666-1.1950000000000003-2.4050000000000002c-3.5500000000000007-2.9583333333333335-7.853333333333332-4.746666666666667-12.45-5.161666666666667-0.09833333333333982-0.004999999999998339-0.1983333333333377-0.009999999999998233-0.29833333333333556-0.009999999999998233-0.8283333333333331 0-1.6333333333333329 0.31000000000000005-2.25 0.8716666666666668-0.6900000000000013 0.6333333333333333-1.0833333333333357 1.525-1.0833333333333357 2.4616666666666664v3.716666666666667c-6.095000000000001 1.4399999999999995-10.555000000000001 6.9783333333333335-10.555000000000001 13.34166666666667 0 7.578333333333333 6.1466666666666665 13.740000000000002 13.706666666666665 13.740000000000002 2.6583333333333314 0 5.183333333333334-0.7999999999999972 7.403333333333336-2.2433333333333323 0.461666666666666 0.23666666666666458 0.9716666666666676 0.37666666666666515 1.5 0.37666666666666515 0.061666666666667425 0 0.120000000000001-0.0033333333333303017 0.18333333333333357-0.006666666666667709 0.9466666666666654-0.04999999999999716 1.8283333333333331-0.5083333333333329 2.419999999999998-1.25 2.366666666666667-2.9666666666666686 3.673333333333332-6.699999999999999 3.673333333333332-10.494999999999997 0.0033333333333303017-3.388333333333332-1.0416666666666643-6.713333333333333-2.9499999999999993-9.5z m-12.228333333333332 19.783333333333335c-5.7283333333333335 0-10.371666666666666-4.656666666666666-10.371666666666666-10.408333333333331 0-5.228333333333332 3.861666666666668-9.55 8.888333333333335-10.273333333333333v10.683333333333334l7.683333333333334 7.921666666666667c-1.7300000000000004 1.3000000000000007-3.871666666666666 2.0766666666666644-6.199999999999999 2.0766666666666644z m0.18333333333333357-12.425v-15.041666666666668c4.016666666666666 0.3633333333333333 7.678333333333335 1.955 10.61 4.4l-10.61 10.643333333333333z m0.3566666666666656 2.116666666666667l8.366666666666667-8.383333333333335c1.8366666666666625 2.3116666666666656 2.9433333333333316 5.2383333333333315 2.9433333333333316 8.408333333333333 0 3.1883333333333326-1.1066666666666656 6.108333333333334-2.9499999999999993 8.416666666666668l-8.363333333333333-8.443333333333335z' })
                )
            );
        }
    }]);

    return TiChartPieOutline;
}(React.Component);

exports.default = TiChartPieOutline;
module.exports = exports['default'];