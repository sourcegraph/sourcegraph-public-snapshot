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

var MdPhonelinkOff = function (_React$Component) {
    _inherits(MdPhonelinkOff, _React$Component);

    function MdPhonelinkOff() {
        _classCallCheck(this, MdPhonelinkOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPhonelinkOff).apply(this, arguments));
    }

    _createClass(MdPhonelinkOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm38.36 13.360000000000001q0.7033333333333331 0 1.1716666666666669 0.4666666666666668t0.4683333333333337 1.173333333333332v16.64q0 0.7033333333333331-0.46666666666666856 1.211666666666666t-1.173333333333332 0.509999999999998h-0.3133333333333326l-5-5h3.5933333333333337v-11.716666666666669h-6.640000000000001v8.670000000000002l-3.3599999999999994-3.3599999999999994v-6.954999999999998q0-0.7033333333333331 0.5083333333333329-1.1716666666666669t1.2100000000000009-0.4666666666666668h10z m-31.72-2.8933333333333344v17.891666666666666h17.89z m-3.4366666666666674-7.733333333333333q6.33 6.333333333333335 18.008333333333336 18.05t14.33666666666667 14.373333333333331l-2.1116666666666646 2.1099999999999994-3.9066666666666663-3.9066666666666663h-29.53000000000001v-5h3.3600000000000003v-18.36q0-1.1716666666666669 0.7833333333333332-2.1100000000000003l-3.0516666666666667-3.046666666666666z m33.436666666666675 7.266666666666667h-21.95333333333334l-3.3599999999999994-3.3599999999999994h25.313333333333333v3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdPhonelinkOff;
}(React.Component);

exports.default = MdPhonelinkOff;
module.exports = exports['default'];