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

var MdAirlineSeatLegroomExtra = function (_React$Component) {
    _inherits(MdAirlineSeatLegroomExtra, _React$Component);

    function MdAirlineSeatLegroomExtra() {
        _classCallCheck(this, MdAirlineSeatLegroomExtra);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAirlineSeatLegroomExtra).apply(this, arguments));
    }

    _createClass(MdAirlineSeatLegroomExtra, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm38.04666666666667 28.75q0.5466666666666669 0.9383333333333326 0.1566666666666663 1.9533333333333331t-1.3283333333333331 1.4833333333333343l-6.171666666666667 2.81666666666667-5.703333333333333-11.643333333333338h-11.639999999999999q-2.033333333333333 0-3.5166666666666675-1.4833333333333343t-1.4833333333333343-3.5166666666666657v-13.36h10v10h5.859999999999999q2.1099999999999994 0 2.9666666666666686 1.8766666666666652l5.626666666666665 11.64 1.875-0.8599999999999994q0.9383333333333326-0.39000000000000057 1.913333333333334-0.07833333333333314t1.4466666666666654 1.1716666666666669z m-31.406666666666666-8.75q0 2.0333333333333314 1.4833333333333343 3.5166666666666657t3.5166666666666657 1.4833333333333343h10v3.3599999999999994h-10q-3.4383333333333344 0-5.86-2.461666666666666t-2.4216666666666664-5.898333333333333v-15h3.2833333333333337v15z' })
                )
            );
        }
    }]);

    return MdAirlineSeatLegroomExtra;
}(React.Component);

exports.default = MdAirlineSeatLegroomExtra;
module.exports = exports['default'];