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

var FaTry = function (_React$Component) {
    _inherits(FaTry, _React$Component);

    function FaTry() {
        _classCallCheck(this, FaTry);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTry).apply(this, arguments));
    }

    _createClass(FaTry, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.85714285714286 18.571428571428573q0 4.262857142857143-2.1099999999999994 7.879999999999999t-5.725714285714286 5.725714285714282-7.878571428571433 2.1085714285714303h-3.571428571428571q-0.31428571428571495 0-0.5142857142857142-0.20000000000000284t-0.1999999999999993-0.5142857142857125v-13.637142857142855l-4.800000000000001 1.4714285714285715q-0.06714285714285673 0.024285714285714022-0.20000000000000018 0.024285714285714022-0.2242857142857142 0-0.42571428571428616-0.13285714285714434-0.28857142857142737-0.22571428571428243-0.28857142857142737-0.5828571428571401v-2.8571428571428577q0-0.5114285714285707 0.5114285714285716-0.6900000000000013l5.200000000000001-1.5857142857142854v-2.072857142857142l-4.797142857142859 1.4699999999999989q-0.06714285714285673 0.024285714285714022-0.20000000000000018 0.024285714285714022-0.2242857142857142 0-0.42571428571428616-0.13428571428571345-0.28857142857142737-0.22285714285714064-0.28857142857142737-0.5799999999999983v-2.8571428571428577q0-0.5142857142857142 0.5114285714285716-0.6914285714285722l5.200000000000001-1.5857142857142854v-5.577142857142857q0-0.31428571428571406 0.20285714285714285-0.5142857142857142t0.5142857142857142-0.20000000000000018h3.5714285714285694q0.3114285714285714 0 0.5114285714285707 0.20000000000000018t0.1999999999999993 0.5142857142857142v4.037142857142857l8.37142857142857-2.588571428571428q0.33428571428571274-0.11142857142857121 0.6257142857142846 0.11142857142857121t0.28999999999999915 0.5800000000000001v2.8571428571428568q0 0.5142857142857142-0.514285714285716 0.6914285714285722l-8.769999999999992 2.6999999999999975v2.0771428571428583l8.368571428571428-2.59q0.33428571428571274-0.1114285714285721 0.6257142857142846 0.1114285714285721t0.2914285714285718 0.5785714285714274v2.8571428571428577q0 0.5142857142857142-0.514285714285716 0.6928571428571431l-8.771428571428572 2.6999999999999993v10.87142857142857q4.195714285714285-0.28999999999999915 7.097142857142856-3.37142857142857t2.8999999999999986-7.32q0-0.3114285714285714 0.20285714285714107-0.5114285714285707t0.5142857142857125-0.1999999999999993h3.5714285714285694q0.3114285714285714 0 0.5114285714285742 0.1999999999999993t0.20000000000000284 0.5142857142857125z' })
                )
            );
        }
    }]);

    return FaTry;
}(React.Component);

exports.default = FaTry;
module.exports = exports['default'];