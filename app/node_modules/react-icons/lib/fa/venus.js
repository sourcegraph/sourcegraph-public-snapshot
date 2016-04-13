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

var FaVenus = function (_React$Component) {
    _inherits(FaVenus, _React$Component);

    function FaVenus() {
        _classCallCheck(this, FaVenus);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaVenus).apply(this, arguments));
    }

    _createClass(FaVenus, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.85714285714286 12.857142857142858q0 4.9328571428571415-3.2928571428571445 8.582857142857144t-8.135714285714286 4.185714285714283v5.802857142857146h5q0.31428571428571317 0 0.5142857142857125 0.1999999999999993t0.1999999999999993 0.514285714285716v1.4285714285714306q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.5142857142857125 0.20000000000000284h-5v5q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.514285714285716 0.19999999999999574h-1.428571428571427q-0.31428571428571317 0-0.514285714285716-0.20000000000000284t-0.1999999999999993-0.5142857142857125v-5h-5.000000000000002q-0.31428571428571495 0-0.5142857142857142-0.20000000000000284t-0.1999999999999993-0.5142857142857125v-1.4285714285714306q0-0.31428571428571317 0.1999999999999993-0.5142857142857125t0.5142857142857142-0.19999999999999574h5.000000000000002v-5.8028571428571425q-3.3485714285714288-0.35714285714285765-6.0600000000000005-2.3000000000000007t-4.151428571428573-5-1.1714285714285717-6.514285714285714q0.24571428571428644-2.9942857142857164 1.797142857142858-5.561428571428573t4.062857142857142-4.1971428571428575 5.48-1.9628571428571424q3.7942857142857136-0.4242857142857144 7.120000000000001 1.2057142857142857t5.265714285714285 4.732857142857142 1.9428571428571466 6.828571428571429z m-22.85714285714286 0q0 4.12857142857143 2.935714285714287 7.064285714285713t7.064285714285713 2.935714285714287 7.064285714285717-2.935714285714287 2.9357142857142833-7.064285714285713-2.935714285714287-7.064285714285715-7.064285714285713-2.9357142857142855-7.064285714285715 2.9357142857142855-2.935714285714285 7.064285714285715z' })
                )
            );
        }
    }]);

    return FaVenus;
}(React.Component);

exports.default = FaVenus;
module.exports = exports['default'];