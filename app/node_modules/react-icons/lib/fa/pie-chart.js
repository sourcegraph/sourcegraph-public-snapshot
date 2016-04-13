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

var FaPieChart = function (_React$Component) {
    _inherits(FaPieChart, _React$Component);

    function FaPieChart() {
        _classCallCheck(this, FaPieChart);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaPieChart).apply(this, arguments));
    }

    _createClass(FaPieChart, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.857142857142858 19.865714285714287l12.185714285714287 12.185714285714287q-2.364285714285714 2.411428571428573-5.522857142857141 3.751428571428569t-6.6628571428571455 1.3400000000000034q-4.6657142857142855 0-8.604285714285714-2.299999999999997t-6.238571428571429-6.235714285714295-2.3-8.607142857142854 2.3000000000000003-8.6 6.238571428571429-6.242857142857143 8.604285714285714-2.295714285714286v17.00857142857143z m4.174285714285713 0.13428571428571345h17.254285714285714q0 3.5042857142857144-1.3399999999999963 6.662857142857142t-3.75 5.524285714285718z m15.825714285714291-2.8571428571428577h-17.142857142857146v-17.142857142857142q4.665714285714287 0 8.604285714285716 2.3000000000000003t6.238571428571426 6.237142857142857 2.3000000000000043 8.605714285714285z' })
                )
            );
        }
    }]);

    return FaPieChart;
}(React.Component);

exports.default = FaPieChart;
module.exports = exports['default'];