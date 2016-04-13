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

var MdBatteryFull = function (_React$Component) {
    _inherits(MdBatteryFull, _React$Component);

    function MdBatteryFull() {
        _classCallCheck(this, MdBatteryFull);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBatteryFull).apply(this, arguments));
    }

    _createClass(MdBatteryFull, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.093333333333334 6.640000000000001q0.9383333333333326 0 1.6000000000000014 0.6633333333333331t0.6666666666666679 1.6000000000000005v25.549999999999997q0 0.9366666666666674-0.6666666666666679 1.5616666666666674t-1.6000000000000014 0.6233333333333348h-12.188333333333333q-0.9383333333333326 0-1.5999999999999996-0.625t-0.6666666666666661-1.56666666666667v-25.54q0-0.9383333333333317 0.6666666666666661-1.5999999999999979t1.5999999999999996-0.666666666666667h2.7333333333333343v-3.280000000000001h6.719999999999999v3.283333333333333h2.7333333333333343z' })
                )
            );
        }
    }]);

    return MdBatteryFull;
}(React.Component);

exports.default = MdBatteryFull;
module.exports = exports['default'];