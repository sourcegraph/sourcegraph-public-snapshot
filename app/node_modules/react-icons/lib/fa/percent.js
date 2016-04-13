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

var FaPercent = function (_React$Component) {
    _inherits(FaPercent, _React$Component);

    function FaPercent() {
        _classCallCheck(this, FaPercent);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaPercent).apply(this, arguments));
    }

    _createClass(FaPercent, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.42857142857143 28.571428571428573q0-1.1600000000000001-0.8485714285714288-2.008571428571429t-2.008571428571429-0.8485714285714288-2.008571428571429 0.8485714285714288-0.8485714285714288 2.008571428571429 0.8485714285714288 2.008571428571429 2.008571428571429 0.8485714285714288 2.008571428571429-0.8485714285714288 0.8485714285714288-2.008571428571429z m-17.142857142857146-17.142857142857146q1.7763568394002505e-15-1.1599999999999984-0.848571428571427-2.008571428571427t-2.008571428571429-0.8485714285714288-2.008571428571429 0.8485714285714288-0.8485714285714288 2.008571428571429 0.8485714285714288 2.008571428571429 2.008571428571429 0.8485714285714288 2.008571428571429-0.8485714285714288 0.8485714285714288-2.008571428571429z m22.85714285714286 17.142857142857142q0 3.548571428571428-2.5114285714285742 6.060000000000002t-6.059999999999999 2.5114285714285742-6.059999999999999-2.5114285714285742-2.5114285714285742-6.059999999999999 2.5114285714285707-6.059999999999999 6.060000000000002-2.5114285714285742 6.059999999999999 2.5114285714285707 2.5114285714285742 6.060000000000002z m-2.142857142857146-24.285714285714285q0 0.4471428571428584-0.28999999999999915 0.8485714285714296l-23.571428571428573 31.428571428571434q-0.4242857142857126 0.5799999999999983-1.138571428571428 0.5799999999999983h-3.571428571428571q-0.5800000000000001 0-1.0042857142857144-0.42428571428571615t-0.4242857142857144-1.0042857142857144q0-0.4471428571428575 0.29000000000000004-0.8485714285714252l23.571428571428573-31.42857142857143q0.4242857142857126-0.5800000000000023 1.138571428571428-0.5800000000000023h3.5714285714285694q0.5799999999999983 0 1.0042857142857144 0.4242857142857144t0.42428571428571615 1.004285714285714z m-15 7.142857142857144q0 3.548571428571428-2.5114285714285707 6.0600000000000005t-6.0600000000000005 2.5114285714285707-6.0600000000000005-2.5114285714285707-2.511428571428571-6.0600000000000005 2.511428571428571-6.0600000000000005 6.0600000000000005-2.511428571428571 6.0600000000000005 2.511428571428571 2.5114285714285707 6.0600000000000005z' })
                )
            );
        }
    }]);

    return FaPercent;
}(React.Component);

exports.default = FaPercent;
module.exports = exports['default'];