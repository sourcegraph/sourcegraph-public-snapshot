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

var MdFlightLand = function (_React$Component) {
    _inherits(MdFlightLand, _React$Component);

    function MdFlightLand() {
        _classCallCheck(this, MdFlightLand);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFlightLand).apply(this, arguments));
    }

    _createClass(MdFlightLand, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.36 24.063333333333336q-6.406666666666663-1.7966666666666704-16.093333333333334-4.29666666666667l-2.6566666666666645-0.783333333333335v-8.589999999999998l2.421666666666667 0.625 1.5616666666666674 3.9066666666666663 8.283333333333333 2.1883333333333344v-13.75l3.1999999999999993 0.8600000000000003 4.611666666666665 15 8.828333333333333 2.3433333333333337q1.0166666666666657 0.3116666666666674 1.5249999999999986 1.2100000000000009t0.27333333333333343 1.913333333333334q-0.31666666666667 1.0166666666666657-1.173333333333332 1.4833333333333343t-1.875 0.23666666666666814z m-19.22 7.576666666666668h31.71666666666667v3.359999999999996h-31.715000000000003v-3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdFlightLand;
}(React.Component);

exports.default = MdFlightLand;
module.exports = exports['default'];