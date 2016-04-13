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

var FaHourglass1 = function (_React$Component) {
    _inherits(FaHourglass1, _React$Component);

    function FaHourglass1() {
        _classCallCheck(this, FaHourglass1);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaHourglass1).apply(this, arguments));
    }

    _createClass(FaHourglass1, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm34.285714285714285 2.857142857142857q0 5.825714285714286-2.3771428571428572 10.3t-5.948571428571427 6.842857142857142q3.571428571428573 2.3657142857142865 5.948571428571427 6.842857142857142t2.3771428571428572 10.300000000000004h2.142857142857146q0.3142857142857167 0 0.5142857142857125 0.20000000000000284t0.20000000000000284 0.5142857142857125v1.4285714285714306q0 0.3142857142857167-0.20000000000000284 0.5142857142857125t-0.5142857142857125 0.20000000000000284h-32.85714285714286q-0.3142857142857123 0-0.514285714285712-0.20000000000000284t-0.20000000000000018-0.5142857142857196v-1.4285714285714306q0-0.3142857142857167 0.20000000000000018-0.5142857142857125t0.5142857142857142-0.19999999999999574h2.142857142857143q0-5.825714285714287 2.377142857142858-10.3t5.948571428571427-6.842857142857145q-3.571428571428571-2.3657142857142865-5.948571428571428-6.842857142857143t-2.3771428571428563-10.299999999999999h-2.142857142857143q-0.3142857142857145-4.440892098500626e-16-0.5142857142857142-0.20000000000000062t-0.20000000000000018-0.5142857142857142v-1.4285714285714286q0-0.31428571428571417 0.20000000000000018-0.5142857142857142t0.5142857142857142-0.19999999999999996h32.85714285714286q0.3142857142857167 0 0.5142857142857125 0.2t0.20000000000000284 0.5142857142857142v1.4285714285714286q0 0.3142857142857145-0.20000000000000284 0.5142857142857142t-0.5142857142857125 0.20000000000000018h-2.142857142857146z m-2.8571428571428577 0h-22.857142857142854q-1.7763568394002505e-15 1.471428571428572 0.1999999999999975 2.857142857142857h22.457142857142856q0.1999999999999993-1.3614285714285712 0.1999999999999993-2.857142857142857z m0 34.285714285714285q0-2.8999999999999986-0.7571428571428562-5.568571428571428t-2.0214285714285722-4.642857142857142-2.8242857142857147-3.3928571428571423-3.2571428571428562-2.109999999999996h-5.135714285714286q-1.6971428571428575 0.6914285714285704-3.2571428571428562 2.1099999999999994t-2.8257142857142856 3.3928571428571423-2.0199999999999996 4.642857142857142-0.7585714285714289 5.568571428571431h22.85714285714286z' })
                )
            );
        }
    }]);

    return FaHourglass1;
}(React.Component);

exports.default = FaHourglass1;
module.exports = exports['default'];