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

var MdBatteryUnknown = function (_React$Component) {
    _inherits(MdBatteryUnknown, _React$Component);

    function MdBatteryUnknown() {
        _classCallCheck(this, MdBatteryUnknown);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBatteryUnknown).apply(this, arguments));
    }

    _createClass(MdBatteryUnknown, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.828333333333337 21.171666666666667q1.1716666666666633-1.1716666666666669 1.1716666666666633-2.8133333333333326 0-2.0333333333333314-1.4833333333333343-3.5166666666666657t-3.5166666666666657-1.4816666666666674-3.5166666666666657 1.4833333333333325-1.4833333333333343 3.5166666666666657h2.5q0-1.0166666666666657 0.7033333333333331-1.7583333333333329t1.7966666666666669-0.7433333333333323 1.7966666666666669 0.7416666666666671 0.7033333333333331 1.7583333333333329-0.7033333333333331 1.7166666666666686l-1.5633333333333326 1.5666666666666664q-1.5633333333333326 1.5616666666666674-1.5633333333333326 3.3583333333333343h2.6566666666666663q0-1.25 1.4066666666666663-2.6566666666666663z m-2.2666666666666657 8.75v-3.2049999999999983h-3.123333333333335v3.2049999999999983h3.125z m4.533333333333335-23.283333333333335q0.9366666666666674 0 1.6000000000000014 0.666666666666667t0.663333333333334 1.6000000000000005v25.545q0 0.9399999999999977-0.663333333333334 1.56666666666667t-1.6000000000000014 0.6233333333333348h-12.191666666666672q-0.9383333333333326 0-1.5999999999999996-0.625t-0.6666666666666661-1.56666666666667v-25.54166666666667q0-0.9383333333333317 0.6666666666666661-1.5999999999999979t1.5999999999999996-0.666666666666667h2.7333333333333325v-3.280000000000001h6.720000000000002v3.283333333333333h2.7333333333333343z' })
                )
            );
        }
    }]);

    return MdBatteryUnknown;
}(React.Component);

exports.default = MdBatteryUnknown;
module.exports = exports['default'];