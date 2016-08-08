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

var TiChartLine = function (_React$Component) {
    _inherits(TiChartLine, _React$Component);

    function TiChartLine() {
        _classCallCheck(this, TiChartLine);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiChartLine).apply(this, arguments));
    }

    _createClass(TiChartLine, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm7.916666666666667 25.933333333333334c0.6166666666666663 0.49166666666666714 1.3500000000000005 0.7333333333333343 2.079999999999999 0.7333333333333343 0.9800000000000004 0 1.9466666666666672-0.4299999999999997 2.6050000000000004-1.2533333333333339l4.649999999999999-5.810000000000002 4.0816666666666706 3.063333333333336c1.4400000000000013 1.0800000000000018 3.4800000000000004 0.8200000000000003 4.603333333333332-0.586666666666666l6.666666666666664-8.333333333333334c1.1499999999999986-1.4333333333333336 0.9166666666666643-3.533333333333333-0.5200000000000031-4.683333333333334-1.4383333333333326-1.1500000000000004-3.533333333333335-0.9199999999999999-4.686666666666667 0.5166666666666675l-4.649999999999999 5.811666666666667-4.079999999999995-3.0583333333333353c-1.4400000000000013-1.08-3.4783333333333335-0.8233333333333341-4.6033333333333335 0.5833333333333339l-6.666666666666667 8.333333333333332c-1.1499999999999995 1.4383333333333326-0.916666666666667 3.533333333333335 0.5200000000000005 4.683333333333334z m0.41666666666666696 9.066666666666666h23.333333333333336c0.9216666666666669 0 1.6666666666666643-0.7449999999999974 1.6666666666666643-1.6666666666666643s-0.7449999999999974-1.6666666666666679-1.6666666666666679-1.6666666666666679h-23.333333333333336c-0.9216666666666651 0-1.6666666666666652 0.745000000000001-1.6666666666666652 1.6666666666666679s0.7450000000000001 1.6666666666666643 1.666666666666667 1.6666666666666643z' })
                )
            );
        }
    }]);

    return TiChartLine;
}(React.Component);

exports.default = TiChartLine;
module.exports = exports['default'];