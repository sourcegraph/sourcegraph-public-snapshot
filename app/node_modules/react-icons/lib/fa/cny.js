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

var FaCny = function (_React$Component) {
    _inherits(FaCny, _React$Component);

    function FaCny() {
        _classCallCheck(this, FaCny);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCny).apply(this, arguments));
    }

    _createClass(FaCny, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm22.031428571428574 34.285714285714285h-3.8400000000000034q-0.28999999999999915 0-0.5028571428571418-0.20000000000000284t-0.21142857142856997-0.5142857142857125v-7.3657142857142865h-6.428571428571429q-0.2900000000000009 0-0.5028571428571436-0.1999999999999993t-0.21142857142857174-0.5142857142857125v-2.3000000000000007q0-0.28857142857143003 0.21142857142857174-0.5t0.5028571428571436-0.2142857142857153h6.428571428571429v-1.8957142857142841h-6.428571428571429q-0.2900000000000009 0-0.5028571428571436-0.1999999999999993t-0.21142857142857174-0.5142857142857125v-2.321428571428573q0-0.28999999999999915 0.21142857142857174-0.5028571428571418t0.5028571428571436-0.21142857142856997h4.777142857142858l-7.165714285714287-12.900000000000004q-0.17857142857142883-0.35714285714285676 0-0.714285714285714 0.22285714285714242-0.35999999999999943 0.6257142857142863-0.35999999999999943h4.328571428571427q0.42571428571428527 0 0.6485714285714295 0.3999999999999999l4.800000000000001 9.485714285714288q0.4228571428571435 0.8499999999999996 1.2485714285714273 2.791428571428572 0.2228571428571442-0.5357142857142865 0.6814285714285724-1.5171428571428578t0.6142857142857139-1.361428571428572l4.2642857142857125-9.370000000000001q0.17571428571428882-0.4285714285714275 0.6428571428571495-0.4285714285714275h4.262857142857143q0.379999999999999 0 0.6028571428571432 0.3571428571428572 0.1999999999999993 0.31428571428571406 0.022857142857141355 0.6928571428571431l-6.985714285714284 12.924285714285716h4.797142857142859q0.28999999999999915 0 0.5028571428571418 0.21142857142856997t0.21142857142856997 0.5028571428571418v2.321428571428573q0 0.31428571428571317-0.21142857142856997 0.5142857142857125t-0.5028571428571418 0.1999999999999993h-6.4714285714285715v1.8971428571428568h6.4714285714285715q0.28999999999999915 0 0.5028571428571418 0.21142857142856997t0.21142857142856997 0.5028571428571418v2.3000000000000007q0 0.3114285714285714-0.21142857142856997 0.5114285714285707t-0.5028571428571418 0.1999999999999993h-6.4714285714285715v7.367142857142859q0 0.28999999999999915-0.2142857142857153 0.5028571428571453t-0.5 0.21142857142856997z' })
                )
            );
        }
    }]);

    return FaCny;
}(React.Component);

exports.default = FaCny;
module.exports = exports['default'];