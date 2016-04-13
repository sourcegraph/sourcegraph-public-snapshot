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

var MdAirlineSeatReclineExtra = function (_React$Component) {
    _inherits(MdAirlineSeatReclineExtra, _React$Component);

    function MdAirlineSeatReclineExtra() {
        _classCallCheck(this, MdAirlineSeatReclineExtra);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAirlineSeatReclineExtra).apply(this, arguments));
    }

    _createClass(MdAirlineSeatReclineExtra, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.03333333333333 25l9.608333333333334 7.5-2.5 2.5-6.328333333333333-5h-11.406666666666666q-1.7966666666666669 0-3.163333333333334-1.1333333333333329t-1.7599999999999998-2.9299999999999997l-2.261666666666665-9.843333333333334q-0.2333333333333325-1.5633333333333326 0.663333333333334-2.8133333333333326t2.383333333333333-1.5633333333333326h0.07833333333333314q1.7166666666666668-0.2333333333333325 2.8900000000000006 0.7033333333333331l2.736666666666668 2.1099999999999994q3.75 2.8900000000000006 7.813333333333333 2.1099999999999994v3.5933333333333337q-3.75 0.625-8.593333333333334-2.0333333333333314l1.7166666666666686 6.800000000000001h8.126666666666665z m-0.3933333333333344 6.640000000000001v3.3599999999999994h-11.716666666666667q-3.126666666666667 0-5.470000000000001-1.9916666666666671t-2.8133333333333335-5.038333333333334l-3.283333333333333-16.328333333333337h3.283333333333333l3.2833333333333323 15.783333333333335q0.3116666666666674 1.7950000000000017 1.7166666666666668 3.0066666666666677t3.283333333333333 1.2100000000000009h11.716666666666665z m-17.733333333333327-22.266666666666666q-1.0933333333333355-0.7800000000000011-1.328333333333335-2.1466666666666683t0.5466666666666669-2.461666666666667 2.1500000000000004-1.3666666666666667 2.460000000000001 0.5066666666666668q1.0950000000000006 0.8599999999999999 1.3666666666666671 2.226666666666667t-0.5066666666666659 2.461666666666667-2.1883333333333344 1.3283333333333331-2.5-0.5466666666666669z' })
                )
            );
        }
    }]);

    return MdAirlineSeatReclineExtra;
}(React.Component);

exports.default = MdAirlineSeatReclineExtra;
module.exports = exports['default'];