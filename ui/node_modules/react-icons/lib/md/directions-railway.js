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

var MdDirectionsRailway = function (_React$Component) {
    _inherits(MdDirectionsRailway, _React$Component);

    function MdDirectionsRailway() {
        _classCallCheck(this, MdDirectionsRailway);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDirectionsRailway).apply(this, arguments));
    }

    _createClass(MdDirectionsRailway, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 16.64v-8.283333333333333h-20v8.283333333333333h20z m-10 11.719999999999999q1.3283333333333331 0 2.3433333333333337-1.0166666666666657t1.0166666666666657-2.3416666666666686-1.0166666666666657-2.3416666666666686-2.3433333333333337-1.0166666666666657-2.3433333333333337 1.0166666666666657-1.0166666666666657 2.3400000000000034 1.0166666666666657 2.344999999999999 2.3433333333333337 1.0166666666666657z m-13.360000000000001-2.5v-17.5q0-3.9833333333333343 3.4383333333333344-5.3500000000000005t9.921666666666667-1.3683333333333323 9.921666666666667 1.3666666666666667 3.4383333333333326 5.350000000000001v17.5q0 2.421666666666667-1.7166666666666686 4.100000000000001t-4.141666666666666 1.6833333333333336l2.5 2.4999999999999964v0.8583333333333343h-20.001666666666665v-0.8583333333333343l2.5-2.5q-2.42 0-4.138333333333334-1.6799999999999997t-1.7166666666666668-4.100000000000001z' })
                )
            );
        }
    }]);

    return MdDirectionsRailway;
}(React.Component);

exports.default = MdDirectionsRailway;
module.exports = exports['default'];