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

var FaGgCircle = function (_React$Component) {
    _inherits(FaGgCircle, _React$Component);

    function FaGgCircle() {
        _classCallCheck(this, FaGgCircle);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaGgCircle).apply(this, arguments));
    }

    _createClass(FaGgCircle, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.004285714285714 30.22285714285714l6.048571428571428-6.048571428571428-6.2285714285714295-6.228571428571431-1.9628571428571426 1.9657142857142844 4.2857142857142865 4.262857142857143-2.1428571428571423 2.1428571428571423-6.2285714285714295-6.228571428571431 6.2285714285714295-6.225714285714286 0.8928571428571423 0.8928571428571423 1.942857142857143-1.942857142857143-2.8357142857142854-2.8571428571428577-10.134285714285715 10.134285714285713z m7.991428571428571-0.17857142857142705l10.134285714285717-10.13285714285714-10.13285714285714-10.134285714285715-6.048571428571428 6.04857142857143 6.228571428571428 6.228571428571431 1.9628571428571426-1.9657142857142844-4.285714285714285-4.261428571428571 2.1428571428571423-2.1428571428571423 6.228571428571428 6.228571428571428-6.228571428571428 6.225714285714286-0.8928571428571423-0.8928571428571423-1.942857142857143 1.9642857142857153z m16.004285714285714-10.044285714285714q0 4.062857142857144-1.585714285714289 7.767142857142858t-4.261428571428574 6.385714285714286-6.385714285714286 4.261428571428574-7.764285714285705 1.5857142857142819-7.765714285714285-1.585714285714289-6.385714285714285-4.262857142857143-4.261428571428572-6.385714285714286-1.5857142857142892-7.765714285714282 1.5857142857142859-7.767142857142858 4.261428571428572-6.385714285714285 6.385714285714285-4.261428571428572 7.762857142857143-1.5857142857142854 7.768571428571427 1.5857142857142859 6.385714285714286 4.262857142857143 4.261428571428574 6.385714285714286 1.5842857142857127 7.765714285714285z' })
                )
            );
        }
    }]);

    return FaGgCircle;
}(React.Component);

exports.default = FaGgCircle;
module.exports = exports['default'];